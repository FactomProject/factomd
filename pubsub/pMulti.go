package pubsub

// PubSimpleMulti is a very basic idea of keeping track of multiple
// writers. The close functionality only happens if ALL writers close
// the publish.
type PubSimpleMulti struct {
	*PubThreaded

	total int
}

// TODO: Do this better. This doesn't keep track very well
func NewPubMulti(buffer int) *PubSimpleMulti {
	p := new(PubSimpleMulti)
	p.PubThreaded = NewPubThreaded(buffer)

	return p
}

func (m *PubSimpleMulti) Publish(path string) *PubSimpleMulti {
	// Multi might need to return the existing multi
	pub := globalReg.FindPublisher(path)
	if pub == nil {
		// First multi, initiate the register
		globalPublish(path, m)
		return m
	}
	// Publisher already exists
	multi, ok := pub.(*PubSimpleMulti)
	if !ok {
		panic("tried to register a multi on a path that another publisher type exists")
	}

	return multi.NewPublisher()
}

func (m *PubSimpleMulti) NewPublisher() *PubSimpleMulti {
	m.total++
	return m
}

func (m *PubSimpleMulti) Close() error {
	m.total--
	if m.total <= 0 {
		return m.PubThreaded.Close()
	}
	return nil
}
