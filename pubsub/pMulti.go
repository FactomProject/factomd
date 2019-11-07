package pubsub

// PubSimpleMulti is a very basic idea of keeping track of multiple
// writers. The close functionality only happens if ALL writers close
// the publish.
type PubSimpleMulti struct {
	*PubThreaded

	total int
}

// TODO: Do this better. This doesn't keep track very well
func NewSimpleMultiPublish(buffer int) *PubSimpleMulti {
	p := new(PubSimpleMulti)
	p.PubThreaded = NewThreadedPublisherPublisher(buffer)

	return p
}

func (m *PubSimpleMulti) Publish(path string) *PubSimpleMulti {
	globalPublish(path, m)
	return m
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
