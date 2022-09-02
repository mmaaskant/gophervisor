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
		t.Errorf("Expected queue to be empty instead contains %d tasks", len(qm.queueIterator.queue))
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

	p.Publish(queueManagerTestData)
	if qm.queueIterator.Len() != 1 {
		t.Errorf("Expected queue to contain 1 task, instead contains %d", len(qm.queueIterator.queue))
	}
	for i := 0; i < 4; i++ {
		p.Publish(queueManagerTestData)
	}
	if qm.queueIterator.Len() != 5 {
		t.Errorf("Expected queue to contain 5 tasks, instead contains %d", len(qm.queueIterator.queue))
	}
}

func TestQueueManager_Start(t *testing.T) {
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	qm := newQueueManager()
	p := newPublisher(qm, &f)
	p.Publish(publisherTestData)
	r := <-qm.requestChannel
	qm.RemoveFromQueue(r)
	if r.Data != publisherTestData || r.Function != &f {
		t.Errorf(
			"Response doesn't match published message: expected %s, %v got %s, %v",
			publisherTestData, &f, r.Data, r.Function,
		)
	}
	qm.Shutdown()
}
