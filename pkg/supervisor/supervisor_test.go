package supervisor

import "testing"

const supervisorTestData = "Supervisor test data"
const supervisorTestAmountOfWorkers = 10

func TestNewSupervisor(t *testing.T) {
	sv := NewSupervisor(supervisorTestAmountOfWorkers)
	p, rch := sv.Register(func(p *Publisher, d interface{}, rch chan interface{}) {
		rch <- supervisorTestData
	})
	p.Publish(supervisorTestData)
	if d := <-rch; d != supervisorTestData {
		t.Errorf("Expected %s got %v", supervisorTestData, d)
	}
	for i := 0; i < 10; i++ {
		for i2 := 0; i2 < 1000; i2++ {
			p.Publish(supervisorTestData)
		}
	}
	go sv.Shutdown()
	for d := range rch {
		if d != supervisorTestData {
			t.Errorf("Expected %s got %v", supervisorTestData, d)
			break
		}
	}
}
