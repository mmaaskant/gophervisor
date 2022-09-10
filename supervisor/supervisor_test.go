package supervisor

import (
	"testing"
)

const supervisorTestData = "Supervisor test data"
const supervisorTestAmountOfWorkers = 10
const supervisorAmountOfTasks = 10000

func TestNewSupervisor(t *testing.T) {
	sv := NewSupervisor(supervisorTestAmountOfWorkers)
	p, rch := sv.Register(func(p *Publisher, d any, rch chan any) {
		rch <- d
	})
	p.Publish(supervisorTestData)
	if d := <-rch; d != supervisorTestData {
		t.Errorf("Expected %s got %v", supervisorTestData, d)
	}
	for i := 0; i < supervisorAmountOfTasks; i++ {
		p.Publish(supervisorTestData)
	}
	go sv.Shutdown()
	counter := 0
	for range rch {
		counter++
	}
	if counter != supervisorAmountOfTasks {
		t.Errorf("Expected %d tasks got %d", supervisorAmountOfTasks, counter)
	}
}
