package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/tixu/pingo/utils"
)

type Job struct {
	target Target
}

// Worker represents the worker that executes the job
type Worker struct {
	Id            string
	WorkerPool    chan chan Job
	JobChannel    chan Job
	ResultChannel chan TargetStatus
	quit          chan bool
}

func NewWorker(workerPool chan chan Job, res chan TargetStatus) Worker {
	return Worker{
		Id:            utils.GetRandomName(0),
		WorkerPool:    workerPool,
		JobChannel:    make(chan Job),
		ResultChannel: res,
		quit:          make(chan bool)}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	log.Println("starting worker")
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				log.Printf("%s receidved job %+v", w.Id, job)

				var status TargetStatus
				f := testers[job.target.Test]
				if f != nil {
					status = httpTest(job.target)
					w.ResultChannel <- status
				}

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	JobQueue   chan Job
	ResChannel chan TargetStatus
	maxWorkers int
}

func NewDispatcher(jobqueue chan Job, res chan TargetStatus, maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers, JobQueue: jobqueue, ResChannel: res}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool, d.ResChannel)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	log.Println("starting to dispatch")
	for {
		select {
		case job := <-d.JobQueue:
			// a job request has been received
			fmt.Println("a job request has been received for", job)
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}
