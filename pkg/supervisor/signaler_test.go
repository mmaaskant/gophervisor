package supervisor

import (
	"testing"
	"time"
)

const signalerTestData = "Signaler test data"

func TestSignaler_Start(t *testing.T) {
	amountOfWorkers := 10
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	qm := newQueueManager()
	p := newPublisher(qm, &f)
	s := newSignaler(qm, amountOfWorkers)

	s.Start()
	for i := 0; i < amountOfWorkers; i++ {
		s.isDoneChannel <- false
	}
	time.Sleep(1 * time.Millisecond)
	if s.isDone() != false {
		t.Error("Expected signaler to signal not done due to running workers, got done")
	}
	for i := 0; i < amountOfWorkers; i++ {
		s.isDoneChannel <- true
	}
	time.Sleep(1 * time.Millisecond)
	if s.isDone() != true {
		t.Error("Expected signaler to signal done due to finished workers, got not done")
	}
	p.Publish(signalerTestData)
	if s.isDone() != false {
		t.Error("Expected signaler to signal not done due to queue not being empty, got done")
	}
	qm.Start()
	qm.Shutdown()
	qm.RemoveFromQueue(qm.getNextInQueue())
	if s.isDone() != true {
		t.Error("Expected signaler to signal done due to queue being empty, got not done")
	}
}
