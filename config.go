package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	MailServer     string
	MailServerPort int
	Sender         string
	Mails          []string
	Targets        []Target
}

// Opening (or creating) config file in JSON format
func readConfig(filename string) Config {

	config := Config{
		MailServer:     "smpt.smals.be",
		MailServerPort: 25,
		Sender:         "pingos@smals.be",
		Mails:          []string{"xavier.zebier@smals.be", "patrick.alin@smals.be"},
		Targets:        []Target{Target{Name: "Local HTTP Server", Addr: "localhost:80"}},
	}

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		// unaccessible or not existing file -> creatoin
		file, err = os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}

		// config file just created
		err := json.NewEncoder(file).Encode(config)
		if err != nil {
			log.Fatal(err)
		}

	} else {

		err = json.NewDecoder(file).Decode(&config)

		if err != nil {
			log.Fatal(err)
		}
	}
	return config
}
