package supervisor

// Supervisor starts, registers and shuts down workers.
type Supervisor struct {
	numberOfWorkers  int
	queueManager     *queueManager
	signaler         *signaler
	publishers       map[*func(p *Publisher, d interface{}, rch chan interface{})]*Publisher
	responseChannels map[*func(p *Publisher, d interface{}, rch chan interface{})]chan interface{}
}

// NewSupervisor creates a new Supervisor instance and amount of workers equal to numberOfWorkers.
func NewSupervisor(numberOfWorkers int) *Supervisor {
	qm := newQueueManager()
	sv := &Supervisor{
		numberOfWorkers,
		qm,
		newSignaler(qm, numberOfWorkers),
		make(map[*func(p *Publisher, d interface{}, rch chan interface{})]*Publisher),
		make(map[*func(p *Publisher, d interface{}, rch chan interface{})]chan interface{}),
	}
	sv.start()
	return sv
}

// start Starts the signaler which monitors workers and starts the workers shortly after.
func (sv *Supervisor) start() {
	sv.signaler.Start()
	sv.startWorkers()
}

// startWorkers starts an amount of workers based on Supervisor's numberOfWorkers variable.
func (sv *Supervisor) startWorkers() {
	for i := 0; i < sv.numberOfWorkers; i++ {
		go func() {
			for uoe := range sv.queueManager.requestChannel {
				sv.signaler.isDoneChannel <- false
				(*uoe.Function)(sv.publishers[uoe.Function], uoe.Data, sv.responseChannels[uoe.Function])
				sv.queueManager.RemoveFromQueue(uoe)
				sv.signaler.isDoneChannel <- true
			}
		}()
	}
}

// Register registers a new function and returns an instance of Publisher.
// This Publisher instance publishes new tasks that'll be run within the provided function.
func (sv *Supervisor) Register(
	f func(p *Publisher, d interface{}, rch chan interface{}),
) (*Publisher, chan interface{}) {
	rch := make(chan interface{})
	p := newPublisher(sv.queueManager, &f)
	sv.publishers[&f] = p
	sv.responseChannels[&f] = rch
	return p, rch
}

// Shutdown gracefully shuts down all workers.
// Run Shutdown in a separate routine in case you do not want to wait for Shutdown to finish.
// No new tasks should be published after Shutdown has been called.
func (sv *Supervisor) Shutdown() {
	sv.signaler.Shutdown()
	for _, ch := range sv.responseChannels {
		close(ch)
	}
}
