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
	"github.com/pkg/profile"
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
	x := profile.Start(profile.TraceProfile, profile.ProfilePath("."), profile.NoShutdownHook)

	flag.Parse()

	// Config
	log.Println("Opening config file: ", *filename)
	config := readConfig(*filename)
	log.Printf("Config loaded")

	// Running
	res := make(chan TargetStatus)
	warn := make(chan TargetStatus)
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	initStorage()

	for _, target := range config.Targets {
		ticker := makeTicker(target, res)
		ticker.startTarget(ctx)
	}

	// HTTP

	go startHttp(*httpPort, state)
	go startBrowser(*httpPort, fmt.Sprintf("http://localhost:%d/status", *httpPort))

	go store(ctx, res, warn)

	<-signalChan
	x.Stop()
	log.WithFields(log.Fields{"type": "Main"}).Println("This is the end")

	server.Shutdown(ctx)
	log.WithFields(log.Fields{"type": "server"}).Println("This is the end")
	cancel()
	time.Sleep(25 * time.Second)
}
