package supervisor

import "testing"

const publisherTestData = "Publisher test data"

func TestPublisher_Publish(t *testing.T) {
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	p := newPublisher(newQueueManager(), &f)
	p.Publish(publisherTestData)
	uoe := p.queueManager.getNextInQueue()
	if uoe == nil {
		t.Error("Could not fetch task from queueManager despite publish")
	}
	if uoe.Function != &f {
		t.Errorf("Expected function pointer %v got %v", &f, uoe.Function)
	}
	if uoe.Data != publisherTestData {
		t.Errorf("Expected unitOfWork data %s got %s", publisherTestData, uoe.Data)
	}
}
