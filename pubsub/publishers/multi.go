package publishers

// SimpleMultiPublish is a very basic idea of keeping track of multiple
// writers. The close functionality only happens if ALL writers close
// the publish.
type SimpleMultiPublish struct {
	*Threaded

	total int
}

// TODO: Do this better. This doesn't keep track very well
func NewSimpleMultiPublish(buffer int) *SimpleMultiPublish {
	p := new(SimpleMultiPublish)
	p.Threaded = NewThreadedPublisherPublisher(buffer)

	return p
}

func (m *SimpleMultiPublish) NewPublisher() *SimpleMultiPublish {
	m.total++
	return m
}

func (m *SimpleMultiPublish) Close() error {
	m.total--
	if m.total <= 0 {
		return m.Threaded.Close()
	}
	return nil
}
