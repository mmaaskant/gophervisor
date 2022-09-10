package supervisor

// Publisher publishes tasks carrying the provided function to the queueManager,
// Which will pass the task on to a worker which runs the task using the provided parameters.
type Publisher struct {
	queueManager *queueManager
	function     *func(p *Publisher, d any, rch chan any)
}

// newPublisher creates a new instance of Publisher.
func newPublisher(qm *queueManager, f *func(p *Publisher, d any, rch chan any)) *Publisher {
	return &Publisher{
		qm,
		f,
	}
}

// Publish publishes a new task with the function provided during the creation of the Publisher.
// The task will be run using the data provided.
func (p *Publisher) Publish(data any) {
	p.queueManager.AddToQueue(newUnitOfWork(p.function, data))
}

// unitOfWork carries the function to be executed and the data that it should be executed with.
// A worker receives an instance of unitOfWork and uses it to run a single worker cycle.
type unitOfWork struct {
	Function *func(p *Publisher, d any, rch chan any)
	Data     any
}

// newUnitOfWork creates a new instance of unitOfWork.
func newUnitOfWork(f *func(p *Publisher, d any, rch chan any), d any) *unitOfWork {
	return &unitOfWork{
		f,
		d,
	}
}
