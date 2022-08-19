package supervisor

import (
	"sync"
)

// Signaler tracks the status of all workers within the Supervisor's worker pool.
// On request Signaler is able to broadcast a graceful Shutdown that will trigger once all workers are done and the queue is empty.
type signaler struct {
	totalAmountOfWorkers    int
	numberOfDoneWorkers     int
	isShuttingDown          bool
	isDoneChannel           chan bool
	startShutdownChannel    chan struct{}
	finishedShutdownChannel chan struct{}
	mutex                   sync.Mutex
	queueManager            *queueManager
}

// newSignaler Returns a new instance of signaler.
func newSignaler(queueManager *queueManager, numberOfWorkers int) *signaler {
	return &signaler{
		numberOfWorkers,
		numberOfWorkers,
		false,
		make(chan bool),
		make(chan struct{}),
		make(chan struct{}),
		sync.Mutex{},
		queueManager,
	}
}

// Start starts a new GoRoutine that monitors the status of Supervisor's workers.
// If all workers are done, the queue is empty and the Shutdown signal has been received it shuts down the queueManager.
func (s *signaler) Start() {
	go func() {
		for {
			s.mutex.Lock()
			if s.isShuttingDown && s.isDone() {
				s.queueManager.Shutdown()
				s.finishedShutdownChannel <- struct{}{}
				close(s.startShutdownChannel)
				close(s.finishedShutdownChannel)
				return
			}
			select {
			case <-s.startShutdownChannel:
				s.isShuttingDown = true
			case done := <-s.isDoneChannel:
				if done {
					s.numberOfDoneWorkers += 1
				} else {
					s.numberOfDoneWorkers -= 1
				}
			}
			s.mutex.Unlock()
		}
	}()
}

// Shutdown starts the graceful Shutdown of the Supervisor's workers.
func (s *signaler) Shutdown() {
	s.startShutdownChannel <- struct{}{}
	<-s.finishedShutdownChannel
}

// isDone check if all workers are done and if the queue is empty.
func (s *signaler) isDone() bool {
	return (s.totalAmountOfWorkers == s.numberOfDoneWorkers) && s.queueManager.IsQueueEmpty()
}
