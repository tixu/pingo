package main

import (
	"context"
	"sync"

	log "github.com/Sirupsen/logrus"
)

var state *State

func initStorage() {
	state = NewState()
}

type State struct {
	Lock  sync.Mutex
	State map[Target]TargetStatus
}

func NewState() *State {
	s := new(State)
	s.State = make(map[Target]TargetStatus)
	return s
}

func store(ctx context.Context, res chan TargetStatus, warn chan TargetStatus) {
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
					//warn <- s
					//go sendMail(s, config)
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

}
