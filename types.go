package main

import "time"

type testFun func(t Target) TargetStatus

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

type TargetStatus struct {
	Target    Target
	Online    bool
	Since     time.Time
	LastCheck time.Time
	Test      string
	Error     string
	Stats     map[string]time.Duration
}
