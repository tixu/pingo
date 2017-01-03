package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"os/exec"
	"os/signal"
	"runtime"

	"net/http"
	_ "net/http/pprof"

	"time"

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
	server  *http.Server
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

	flag.Parse()

	// Config
	log.Println("Opening config file: ", *filename)
	config := readConfig(*filename)
	log.Printf("Config loaded")

	// Running
	res := make(chan TargetStatus)
	jobsQueue := make(chan Job)

	dispatcher := NewDispatcher(jobsQueue, &res, config.WorkerNumber)
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	dispatcher.Run(ctx)

	state := NewState()

	for _, target := range config.Targets {
		ticker := makeTicker(target)
		ticker.startTarget(ctx, jobsQueue)
	}

	// HTTP

	go startHttp(*httpPort, state)
	go startBrowser(*httpPort, fmt.Sprintf("http://localhost:%d/status", *httpPort))

	go func(ctx context.Context) {
		for {
			select {

			case status := <-res:
				if s, ok := state.State[status.Target]; ok {
					log.WithFields(log.Fields{"type": "Aggregator"}).Println("target  found ", status.Target)
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
					log.WithFields(log.Fields{"type": "Aggregator"}).Println("target not found ", status.Target)
					status.LastCheck = status.Since
					state.State[status.Target] = status
				}
				log.WithFields(log.Fields{"type": "Aggregator"}).Println("Status to send", status)

				state.State[status.Target] = status

			case <-ctx.Done():
				log.WithFields(log.Fields{"type": "Aggregator"}).Println("stopping")
				return
			}
		}

	}(ctx)

	<-signalChan
	log.WithFields(log.Fields{"type": "Main"}).Println("This is the end")
	server.Shutdown(ctx)
	log.WithFields(log.Fields{"type": "server"}).Println("This is the end")
	cancel()
	time.Sleep(25 * time.Second)
}
