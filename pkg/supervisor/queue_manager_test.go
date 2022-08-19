package supervisor

import (
	"testing"
)

const queueManagerTestData = "Queue manager test data"

func TestQueueManager_IsQueueEmpty(t *testing.T) {
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	qm := newQueueManager()
	p := newPublisher(qm, &f)

	if qm.IsQueueEmpty() != true {
		t.Errorf("Expected queue to be empty instead contains %d tasks", len(qm.queue))
	}
	p.Publish(queueManagerTestData)
	if qm.IsQueueEmpty() == true {
		t.Error("Expected queue to contain 1 task, instead returns empty")
	}
}

func TestQueueManager_AddToQueue(t *testing.T) {
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	qm := newQueueManager()
	p := newPublisher(qm, &f)

	qm.AddToQueue(newUnitOfWork(&f, p))
	if len(qm.queue) != 1 {
		t.Errorf("Expected queue to contain 1 task, instead contains %d", len(qm.queue))
	}
	for i := 0; i < 4; i++ {
		qm.AddToQueue(newUnitOfWork(&f, p))
	}
	if len(qm.queue) != 5 {
		t.Errorf("Expected queue to contain 5 tasks, instead contains %d", len(qm.queue))
	}
}

func TestQueueManager_RemoveFromQueue(t *testing.T) {
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	qm := newQueueManager()
	p := newPublisher(qm, &f)

	p.Publish(publisherTestData)
	qm.RemoveFromQueue(qm.getNextInQueue())
	if len(qm.queue) != 0 {
		t.Errorf("Expected queue to contain 0 tasks, instead contains %d", len(qm.queue))
	}
	for i := 0; i < 5; i++ {
		qm.AddToQueue(newUnitOfWork(&f, p))
	}
	for i := 4; i != 0; i-- {
		qm.RemoveFromQueue(qm.getNextInQueue())
		if len(qm.queue) != i {
			t.Errorf("Expected queue to contain %d tasks, instead contains %d", i, len(qm.queue))
		}
	}
}

func TestQueueManager_Start(t *testing.T) {
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	qm := newQueueManager()
	p := newPublisher(qm, &f)

	p.Publish(publisherTestData)
	qm.Start()
	r := <-qm.requestChannel
	if r.Data != publisherTestData || r.Function != &f {
		t.Errorf(
			"Response doesn't match published message: expected %s, %v got %s, %v",
			publisherTestData, &f, r.Data, r.Function,
		)
	}
	qm.Shutdown()
}
