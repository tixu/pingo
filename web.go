package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
)

type page struct {
	Version     string
	Hash        string
	StateHolder *State
}

func Now() string {
	return time.Now().Format("02/01/2006 15:04:05")
}

// Init of the Web Page template.
var tpl = template.Must(template.New("main").Delims("<%", "%>").Funcs(template.FuncMap{"Now": Now, "json": json.Marshal}).ParseFiles("./tmpl/status.tmpl"))

func startHttp(port int, state *State) {

	p := page{Version: Version, Hash: Build, StateHolder: state}
	mux := http.NewServeMux()
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {

		state.Lock.Lock()

		defer state.Lock.Unlock()

		err := tpl.ExecuteTemplate(w, "status.tmpl", p)
		if err != nil {
			log.Fatal(err)
		}
	})

	s := fmt.Sprintf(":%d", port)

	server = &http.Server{Addr: s, Handler: mux}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
