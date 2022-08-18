package supervisor

import (
	"sync"
)

// Signaler tracks the status of all workers within the Supervisor's worker pool.
// On request Signaler is able to broadcast a graceful shutdown that will trigger once all workers are done and the queue is empty.
type signaler struct {
	totalAmountOfWorkers int
	numberOfDoneWorkers  int
	isShuttingDown       bool
	isDoneChannel        chan bool
	shutdownChannel      chan struct{}
	mutex                sync.Mutex
	queueManager         *queueManager
}

// newSignaler Returns a new instance of signaler.
func newSignaler(queueManager *queueManager, numberOfWorkers int) *signaler {
	return &signaler{
		numberOfWorkers,
		numberOfWorkers,
		false,
		make(chan bool),
		make(chan struct{}),
		sync.Mutex{},
		queueManager,
	}
}

// start starts a new GoRoutine that monitors the status of Supervisor's workers.
// If all workers are done, the queue is empty and the Shutdown signal has been received it shuts down the queueManager.
func (s *signaler) start() {
	go func() {
		for done := range s.isDoneChannel {
			s.mutex.Lock()
			if done {
				s.numberOfDoneWorkers += 1
			} else {
				s.numberOfDoneWorkers -= 1
			}
			if s.isShuttingDown && s.isDone() {
				s.queueManager.shutdown()
				s.shutdownChannel <- struct{}{}
			}
			s.mutex.Unlock()
		}
	}()
}

// shutdown starts the graceful shutdown of the Supervisor's workers.
func (s *signaler) shutdown() {
	s.isShuttingDown = true
	<-s.shutdownChannel
}

// isDone check if all workers are done and if the queue is empty.
func (s *signaler) isDone() bool {
	return (s.totalAmountOfWorkers == s.numberOfDoneWorkers) && s.queueManager.isQueueEmpty()
}
