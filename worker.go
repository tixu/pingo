package main

import (
	"context"

	log "github.com/Sirupsen/logrus"
	utils "github.com/tixu/pingo/utils"
)

type Job struct {
	target Target
}

// Worker represents the worker that executes the job
type Worker struct {
	Id            string
	WorkerPool    chan chan Job
	JobChannel    chan Job
	ResultChannel *chan TargetStatus
}

func NewWorker(workerPool chan chan Job, res *chan TargetStatus) *Worker {
	return &Worker{
		Id:            utils.GetRandomName(0),
		WorkerPool:    workerPool,
		JobChannel:    make(chan Job),
		ResultChannel: res,
	}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w *Worker) Start(ctx context.Context) {

	go func(ctx context.Context) {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				log.WithFields(log.Fields{"type": "Worker", "name": w.Id}).Infof("%s receidved job %+v \n", w.Id, job)
				status := testers[job.target.Test](job.target)
				*w.ResultChannel <- status

			case <-ctx.Done():
				log.WithFields(log.Fields{"type": "Worker", "name": w.Id}).Infoln("stopping", w.Id)
				return
			}
		}
	}(ctx)
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	ID         string
	WorkerPool chan chan Job
	JobQueue   chan Job
	ResChannel *chan TargetStatus
	maxWorkers int
}

func NewDispatcher(jobqueue chan Job, res *chan TargetStatus, maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{ID: utils.GetRandomName(0), WorkerPool: pool, maxWorkers: maxWorkers, JobQueue: jobqueue, ResChannel: res}
}

func (d *Dispatcher) Run(ctx context.Context) {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool, d.ResChannel)
		log.WithFields(log.Fields{"type": "Dispatcher", "name": d.ID}).Infoln("starting worker ", worker.Id)
		worker.Start(ctx)
	}

	go d.dispatch(ctx)
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	log.Println("starting to dispatch")
	for {
		select {
		case job := <-d.JobQueue:
			// a job request has been received
			log.WithFields(log.Fields{"type": "Dispatcher", "name": d.ID}).Infoln("a job request has been received for", job)
			go func() {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job

			}()
		case <-ctx.Done():
			log.WithFields(log.Fields{"type": "Dispatcher", "name": d.ID}).Infoln("stopping")
			return
		}
	}
}
