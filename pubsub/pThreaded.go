package pubsub

// Threaded handles all writes on a separate go routine.
type Threaded struct {
	PubBase

	inputs chan interface{}
}

func NewThreadedPublisherPublisher(buffer int) *Threaded {
	p := new(Threaded)
	p.inputs = make(chan interface{}, buffer)

	return p
}

func (p *Threaded) Close() error {
	close(p.inputs)
	return nil
}

func (p *Threaded) Write(o interface{}) {
	p.inputs <- o
}

func (p *Threaded) Run() {
	for in := range p.inputs { // Run until close
		p.write(in)
	}
	// Close when out of things to write and channel is closed
	_ = p.PubBase.Close()
}

func (p *Threaded) write(o interface{}) {
	p.PubBase.Write(o)
}

func (p *Threaded) Publish(path string) *Threaded {
	globalPublish(path, p)
	return p
}
