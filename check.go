package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type Target struct {

	// Name of the Target
	Name string
	// Address (ex: "localhost:80" of the target
	Addr string
	// Polling interval, in seconds
	Interval int64
	//Type of test http or dial
	Test string
}

type testFun func(t *Target) TargetStatus

type TargetStatus struct {
	Target    *Target
	Online    bool
	Since     time.Time
	LastCheck time.Time
	Test      string
	Error     string
}

func startTarget(t Target, res chan TargetStatus, end chan int) {
	go runTarget(t, res, end)
}

func runTarget(t Target, res chan TargetStatus, end chan int) {

	log.Println("starting runtarget on ", t.Name)
	if t.Interval == 0 {
		t.Interval = 1
	}
	ticker := time.Tick(time.Duration(t.Interval) * time.Second)
	for {
		// Polling
		log.Println("pinging", t.Addr)
		var status TargetStatus
		f := testers[t.Test]
		if f != nil {
			status = httpTest(&t)
			res <- status
		}
		// waiting to ticker
		<-ticker
	}

}

func dialTest(t *Target) TargetStatus {
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "dial", Error: fmt.Sprintf("dial checker error : %s", err)}
	}
	conn.Close()
	return TargetStatus{Target: t, Online: true, Since: time.Now(), Error: ""}

}

func httpTest(t *Target) TargetStatus {

	res, err := http.Get(t.Addr)
	if err != nil {

		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "http", Error: fmt.Sprintf("http checker error : %s", err)}
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("bad status code %d", res.StatusCode)
		return TargetStatus{Target: t, Online: false, Since: time.Now(), Test: "http", Error: fmt.Sprintf("http checker error : %s", err)}
	}
	return TargetStatus{Target: t, Online: true, Since: time.Now(), Test: "http", Error: ""}

}
