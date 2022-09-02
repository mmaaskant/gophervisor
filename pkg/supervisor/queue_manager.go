package supervisor

import (
	"sync"
)

// queueManager manages all incoming tasks and fans them out over the available workers.
// Tasks are added to a queue and tracked, once a task is completed it is removed from this queue.
// If there is no available worker for a message a routine is started that'll publish the backed up messages.
type queueManager struct {
	startShutdownChannel chan struct{}
	hasShutdownChannel   chan struct{}
	requestChannel       chan *unitOfWork
	queueIterator        *queueIterator
	runningMutex         *sync.RWMutex
	running              bool
}

// newQueueManager creates a new instance of queueManager.
func newQueueManager() (qm *queueManager) {
	qm = &queueManager{
		make(chan struct{}),
		make(chan struct{}),
		make(chan *unitOfWork),
		newQueueIterator(),
		&sync.RWMutex{},
		false,
	}
	qm.startShutdownListener()
	return qm
}

// startShutdownListener listens for a shutdown signal,
// once it has been received it'll attempt to shut down the QueueManager and closing its channels.
// If the queue is not empty the moment this shutdown is requested, it will gracefully wait before terminating.
func (qm *queueManager) startShutdownListener() {
	go func() {
		var isShuttingDown bool
	loop:
		for {
			select {
			case <-qm.startShutdownChannel:
				if qm.IsQueueEmpty() {
					break loop
				} else {
					isShuttingDown = true
				}
			case <-qm.queueIterator.isIdleChannel:
				if isShuttingDown && qm.IsQueueEmpty() {
					break loop
				}
			}
		}
		qm.hasShutdownChannel <- struct{}{}
	}()
}

// Start Starts a GoRoutine that publishes all backed up tasks and shuts down once these tasks have been consumed.
func (qm *queueManager) Start() {
	qm.setIsRunning(true)
	go func() {
		defer qm.setIsRunning(false)
		for uoe := qm.queueIterator.Next(); uoe != nil; uoe = qm.queueIterator.Next() {
			qm.requestChannel <- uoe
		}
	}()
}

// running returns the state of the backed up message GoRoutine and ensures it can not run in parallel.
func (qm *queueManager) isRunning() bool {
	qm.runningMutex.RLock()
	defer qm.runningMutex.RUnlock()
	return qm.running
}

// setIsRunning sets the running value of the queueManager.
func (qm *queueManager) setIsRunning(b bool) {
	qm.runningMutex.Lock()
	defer qm.runningMutex.Unlock()
	qm.running = b
}

// AddToQueue looks the queue and publishes a task.
// If there is no available worker to take the task, it calls Start and handles the backed up message(s).
func (qm *queueManager) AddToQueue(uoe *unitOfWork) {
	qm.queueIterator.Register(uoe)
	select {
	case qm.requestChannel <- uoe:
	default:
		qm.queueIterator.MarkAsPending(uoe)
		if !qm.isRunning() {
			qm.Start()
		}
	}
}

// RemoveFromQueue removes to given unitOfWork from the queueIterator.
func (qm *queueManager) RemoveFromQueue(uoe *unitOfWork) {
	qm.queueIterator.Delete(uoe)
}

// IsQueueEmpty locks the queueIterator and checks if it is empty.
func (qm *queueManager) IsQueueEmpty() bool {
	return qm.queueIterator.isEmpty()
}

// Shutdown starts the graceful shutdown of the queueManager.
// Once the queue is empty, all queueManager routines and their channels will be closed.
func (qm *queueManager) Shutdown() {
	qm.startShutdownChannel <- struct{}{}
	<-qm.hasShutdownChannel
	close(qm.hasShutdownChannel)
}

// queueIterator tracks all published unitOfWork instances until they are complete.
// If the queueIterator is empty after processing tasks it will send an idle signal to the shutdown listener.
type queueIterator struct {
	*sync.Mutex
	queue         map[*unitOfWork]bool
	isIdleChannel chan struct{}
}

// newQueueIterator returns a new instance of queueIterator.
func newQueueIterator() *queueIterator {
	return &queueIterator{
		&sync.Mutex{},
		make(map[*unitOfWork]bool),
		make(chan struct{}),
	}
}

// Next returns the next message in the queue, roughly fetched based on the order they were added.
func (qi *queueIterator) Next() *unitOfWork {
	qi.Lock()
	defer qi.Unlock()
	for uoe, status := range qi.queue {
		if status == false {
			qi.queue[uoe] = true
			return uoe
		}
	}
	return nil
}

// Register registers an unitOfWork instance to the queue, so it can be fetched, tracked and ran.
func (qi *queueIterator) Register(uoe *unitOfWork) {
	qi.Lock()
	defer qi.Unlock()
	qi.queue[uoe] = true
}

// MarkAsPending is used to mark a message to be stored in memory and pushed to a worker once a slot is available.
func (qi *queueIterator) MarkAsPending(uoe *unitOfWork) {
	qi.Lock()
	defer qi.Unlock()
	qi.queue[uoe] = false
}

// Delete deletes an instance of unitOfWork from the queue.
// If the queue is empty after this delete, it will send an idle signal to the shutdown listener.
func (qi *queueIterator) Delete(uoe *unitOfWork) {
	qi.Lock()
	defer qi.Unlock()
	delete(qi.queue, uoe)
	if len(qi.queue) == 0 {
		select {
		case qi.isIdleChannel <- struct{}{}:
		default:
		}
	}
}

// isEmpty checks if the queue is currently empty.
func (qi *queueIterator) isEmpty() bool {
	qi.Lock()
	defer qi.Unlock()
	return len(qi.queue) == 0
}

// Len returns the amount of unitOfWork instances that are currently in the queue.
func (qi *queueIterator) Len() int {
	qi.Lock()
	defer qi.Unlock()
	return len(qi.queue)
}
