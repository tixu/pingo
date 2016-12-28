package main

import (
	"flag"
	"fmt"

	"os/exec"
	"runtime"

	log "github.com/Sirupsen/logrus"
)

// Init config

var filename = flag.String("f", "config.json", "JSON configuration file")
var httpPort = flag.Int("p", 8888, "HTTP port")

var testers map[string]testFun = map[string]testFun{
	"dial": dialTest,
	"http": httpTest,
}

var (
	Version string
	Build   string
)

func startBrowser(port int, url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows", "darwin":
		err = exec.Command("cmd", "/c", "start", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

// Main function
func main() {

	log.Println("=============================")
	log.Println("Version: ", Version)
	log.Println("Git commit hash: ", Build)
	log.Println("=============================")

	flag.Parse()

	// Config
	log.Println("Opening config file: ", *filename)
	config := readConfig(*filename)
	log.Printf("Config loaded")

	// Running
	res := make(chan TargetStatus)
	jobsQueue := make(chan Job)
	end := make(chan int)

	dispatcher := NewDispatcher(jobsQueue, res, config.WorkerNumber)
	dispatcher.Run()

	state := NewState()

	for _, target := range config.Targets {
		startTarget(target, res, end, jobsQueue)
	}

	// HTTP

	go startHttp(*httpPort, state)
	go startBrowser(*httpPort, fmt.Sprintf("http://localhost:%d/status", *httpPort))

	for {
		select {

		case status := <-res:
			if s, ok := state.State[status.Target]; ok {
				log.Println("target  found ", status.Target)
				if s.Online != status.Online {

					s.Online = status.Online
					s.Since = status.Since
					s.Error = status.Error
					//	s.Stats = status.Stats
					go sendMail(s, config)
				}
				s.LastCheck = status.Since
				s.Stats = status.Stats
				status = s
			} else {
				log.Println("target not found ", status.Target)
				status.LastCheck = status.Since
				state.State[status.Target] = status
			}
			log.Println("pingo ===>", status)

			state.State[status.Target] = status

		}
	}

}
