package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	httpstat "github.com/tcnksm/go-httpstat"
)

func startTarget(t Target, res chan TargetStatus, end chan int, jobsQueue chan Job) {
	go runTarget(t, res, end, &jobsQueue)
}

func runTarget(t Target, res chan TargetStatus, end chan int, jobsQueue *chan Job) {

	log.Println("starting runtarget on ", t.Name)
	if t.Interval == 0 {
		t.Interval = 1
	}
	ticker := time.Tick(time.Duration(t.Interval) * time.Second)
	for {
		log.Println("posting", t)
		*jobsQueue <- Job{t}
		// waiting to ticker
		<-ticker
	}

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

	if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "http", Error: fmt.Sprintf("http checker error : %s", err)}
	}
	res.Body.Close()
	stats := httpstats(&result, time.Now())
	//result.End(time.Now())

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
