package supervisor

import (
	"sync"
)

// queueManager manages all incoming tasks and fans them out over the available workers.
// Tasks are added to a queue and tracked, once a task is completed it is removed from this queue.
// If there is no available worker for a message a routine is started that'll publish the backed up messages.
type queueManager struct {
	mutex          sync.Mutex
	isRunning      bool
	queue          map[*unitOfWork]bool
	requestChannel chan *unitOfWork
}

// newQueueManager creates a new instance of queueManager.
func newQueueManager() *queueManager {
	qm := &queueManager{
		sync.Mutex{},
		false,
		make(map[*unitOfWork]bool),
		make(chan *unitOfWork),
	}
	return qm
}

// Start Starts a GoRoutine that publishes all backed up tasks and shuts down once these tasks have been consumed.
func (qm *queueManager) Start() {
	if qm.isRunning {
		return
	}
	qm.isRunning = true
	go func() {
		for {
			uoe := qm.getNextInQueue()
			if uoe == nil {
				qm.isRunning = false
				break
			}
			qm.requestChannel <- uoe
		}
	}()
}

// getNextInQueue locks the queue and attempts to get the oldest task that as not been published yet.
// If there is no task, it returns nil.
func (qm *queueManager) getNextInQueue() *unitOfWork {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()
	for uoe, s := range qm.queue {
		if s == false {
			qm.queue[uoe] = true
			return uoe
		}
	}
	return nil
}

// AddToQueue looks the queue and publishes a task.
// If there is no available worker to take the task, it calls Start and handles the backed up message(s).
func (qm *queueManager) AddToQueue(uoe *unitOfWork) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()
	qm.queue[uoe] = false
	select {
	case qm.requestChannel <- uoe:
		return
	default:
		qm.Start()
	}
}

// RemoveFromQueue locks the queue and removes a completed task from the queue.
func (qm *queueManager) RemoveFromQueue(uoe *unitOfWork) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()
	delete(qm.queue, uoe)
}

// IsQueueEmpty locks the queue and checks if it is empty.
func (qm *queueManager) IsQueueEmpty() bool {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()
	return len(qm.queue) == 0
}

func (qm *queueManager) Shutdown() {
	close(qm.requestChannel)
}
