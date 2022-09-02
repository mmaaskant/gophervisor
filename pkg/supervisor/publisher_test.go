package supervisor

import (
	"testing"
)

const publisherTestData = "Publisher test data"

func TestPublisher_Publish(t *testing.T) {
	f := func(p *Publisher, d interface{}, rch chan interface{}) {}
	p := newPublisher(newQueueManager(), &f)
	for i := 0; i <= 100; i++ {
		p.Publish(publisherTestData)
	}
	counter := 0
	for uoe := p.queueManager.queueIterator.Next(); uoe != nil; uoe = p.queueManager.queueIterator.Next() {
		counter++
	}
	if counter != 100 {
		t.Errorf("Expected a 100 iterated UOEs, got %d", counter)
	}
	p.Publish(publisherTestData)
	uoe := p.queueManager.queueIterator.Next()
	if uoe == nil {
		t.Error("Could not fetch task from queueManager despite publish")
	}
	if uoe.Function != &f {
		t.Errorf("Expected function pointer %v got %v", &f, uoe.Function)
	}
	if uoe.Data != publisherTestData {
		t.Errorf("Expected unitOfWork data %s got %s", publisherTestData, uoe.Data)
	}
	if uoe.Data != publisherTestData {
		t.Errorf("Expected UOE to contain %s, got %s", publisherTestData, uoe.Data)
	}
}
