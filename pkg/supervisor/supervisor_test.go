package supervisor

import (
	"fmt"
	"math/rand"
	"testing"
)

const supervisorTestData = "Supervisor test data"
const supervisorTestAmountOfWorkers = 5
const supervisorAmountOfTasks = 10000

func TestNewSupervisor(t *testing.T) {
	sv := NewSupervisor(supervisorTestAmountOfWorkers)
	p, rch := sv.Register(func(p *Publisher, d interface{}, rch chan interface{}) {
		rch <- d
	})
	p.Publish(supervisorTestData)
	if d := <-rch; d != supervisorTestData {
		t.Errorf("Expected %s got %v", supervisorTestData, d)
	}
	for i := 0; i < supervisorAmountOfTasks; i++ {
		s := fmt.Sprintf("%s-%d", supervisorTestData, rand.Intn(10000))
		p.Publish(s)
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
