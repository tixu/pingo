package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	httpstat "github.com/tcnksm/go-httpstat"
	"github.com/tixu/pingo/utils"
)

type Ticker struct {
	ID     string
	ticker <-chan time.Time
	target Target
}

func makeTicker(target Target) *Ticker {

	ticker := make(<-chan time.Time)
	id := utils.GetRandomName(0)
	return &Ticker{ID: id, ticker: ticker, target: target}
}

func (tik *Ticker) startTarget(ctx context.Context, jobsQueue chan Job) {
	log.WithFields(log.Fields{"type": "Ticker", "name": tik.ID}).Infoln("starting ticker")
	go tik.runTarget(ctx, &jobsQueue)
}

func (tik *Ticker) runTarget(ctx context.Context, jobsQueue *chan Job) {

	log.WithFields(log.Fields{"type": "Ticker", "name": tik.ID}).Infoln("starting target", tik.target.Name)
	if tik.target.Interval == 0 {
		panic("interval not allowed")
	}

	tik.ticker = time.Tick(time.Duration(tik.target.Interval) * time.Second)

	go func(ctx context.Context, t *Target) {
		for {
			select {
			case <-tik.ticker:
				log.WithFields(log.Fields{"type": "Ticker", "name": tik.ID}).Infoln("posting", t)
				*jobsQueue <- Job{*t}
				// waiting to ticker
			case <-ctx.Done():
				log.WithFields(log.Fields{"type": "Ticker", "name": tik.ID}).Infoln("stopping ticker ")
				return
			}

		}
	}(ctx, &tik.target)

}

func dialTest(t Target) TargetStatus {
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "dial", Error: fmt.Sprintf("dial checker error : %s", err)}
	}
	conn.Close()
	return TargetStatus{Target: t, Online: true, Since: time.Now(), Error: ""}

}

func httpTest(t Target) TargetStatus {

	req, err := http.NewRequest("GET", t.Addr, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create go-httpstat powered context and pass it to http.Request
	var result httpstat.Result
	ctx := httpstat.WithHTTPStat(req.Context(), &result)
	req = req.WithContext(ctx)

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "http", Error: fmt.Sprintf("http checker error : %s", err)}

	}

	defer res.Body.Close()
	if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "http", Error: fmt.Sprintf("http checker error : %s", err)}
	}

	stats := httpstats(&result, time.Now())

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("bad status code %d", res.StatusCode)
		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "http", Error: fmt.Sprintf("http checker error : %s", err), Stats: stats}
	}
	return TargetStatus{Target: t, Online: true, Since: time.Now(), Test: "http", Error: "", Stats: stats}

}

func httpstats(r *httpstat.Result, t time.Time) map[string]time.Duration {
	return map[string]time.Duration{
		"DNSLookup":        r.DNSLookup / time.Millisecond,
		"TCPConnection":    r.TCPConnection / time.Millisecond,
		"TLSHandshake":     r.TLSHandshake / time.Millisecond,
		"ServerProcessing": r.ServerProcessing / time.Millisecond,
		"ContentTransfer":  r.ContentTransfer(t) / time.Millisecond,

		//"NameLookup":    r.NameLookup / time.Millisecond,
		//"Connect":       r.Connect / time.Millisecond,
		//"Pretransfer":   r.Connect / time.Millisecond,
		//"StartTransfer": r.StartTransfer / time.Millisecond,
		"Total": r.Total(t) / time.Millisecond,
	}
}
