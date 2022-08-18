package supervisor

// Publisher publishes tasks carrying the provided function to the queueManager,
// Which will pass the task on to a worker which runs the task using the provided parameters.
type Publisher struct {
	queueManager *queueManager
	function     *func(p *Publisher, d interface{}, rch chan interface{})
}

// newPublisher creates a new instance of Publisher.
func newPublisher(qm *queueManager, f *func(p *Publisher, d interface{}, rch chan interface{})) *Publisher {
	return &Publisher{
		qm,
		f,
	}
}

// Publish publishes a new task with the function provided during the creation of the Publisher.
// The task will be run using the data provided.
func (p *Publisher) Publish(data interface{}) {
	p.queueManager.addToQueue(newUnitOfWork(p.function, data))
}

// unitOfWork carries the function to be executed and the data that it should be executed with.
// A worker receives an instance of unitOfWork and uses it to run a single worker cycle.
type unitOfWork struct {
	Function *func(p *Publisher, d interface{}, rch chan interface{})
	Data     interface{}
}

// newUnitOfWork creates a new instance of unitOfWork.
func newUnitOfWork(f *func(p *Publisher, d interface{}, rch chan interface{}), d interface{}) *unitOfWork {
	return &unitOfWork{
		f,
		d,
	}
}
