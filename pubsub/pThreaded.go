package pubsub

// PubThreaded handles all writes on a separate go routine.
type PubThreaded struct {
	PubBase

	inputs chan interface{}
}

func NewPubThreaded(buffer int) *PubThreaded {
	p := new(PubThreaded)
	p.inputs = make(chan interface{}, buffer)

	return p
}

func (p *PubThreaded) Close() error {
	close(p.inputs)
	return nil
}

func (p *PubThreaded) Write(o interface{}) {
	p.inputs <- o
}

func (p *PubThreaded) Run() {
	for in := range p.inputs { // Run until close
		p.write(in)
	}
	// Close when out of things to write and channel is closed
	_ = p.PubBase.Close()
}

func (p *PubThreaded) write(o interface{}) {
	p.PubBase.Write(o)
}

func (p *PubThreaded) Publish(path string) *PubThreaded {
	globalPublish(path, p)
	return p
}
