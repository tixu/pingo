package main

import (
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
		return TargetStatus{Target: t, Online: false, Since: time.Now()}
	}
	conn.Close()
	return TargetStatus{Target: t, Online: true, Since: time.Now()}

}

func httpTest(t *Target) TargetStatus {

	res, err := http.Get(t.Addr)
	if err != nil {
		log.Println(err)
		return TargetStatus{Target: t, Online: false, Since: time.Now()}
	}

	if res.StatusCode != http.StatusOK {
		log.Println("bad code", res.StatusCode)
		return TargetStatus{Target: t, Online: false, Since: time.Now()}
	}
	return TargetStatus{Target: t, Online: true, Since: time.Now()}

}
