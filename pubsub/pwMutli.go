package pubsub

import "sync"

// PubMultiWrapper is a very basic idea of keeping track of multiple
// writers. The close functionality only happens if ALL writers close
// the publish.
type PubMultiWrapper struct {
	IPublisher
	PubWrapBase

	start sync.Once

	pubLock sync.Mutex
	total   int
}

type IMultiPublisher interface {
	IPublisherWrapper
	NewPublisher(orig IPublisher) IMultiPublisher
	TotalPublishing() int
}

// TODO: Do this better. This doesn't keep track very well
func PubMultiWrap() *PubMultiWrapper {
	p := new(PubMultiWrapper)

	return p
}

func (m *PubMultiWrapper) Wrap(p IPublisher) IPublisherWrapper {
	m.SetBase(p)
	m.IPublisher = p
	return m
}

func (m *PubMultiWrapper) Start() {
	m.start.Do(func() {
		// Only run the threaded run once.
		// If many publishers try to start the thread, only
		// 1 thread will be started.
		m.IPublisher.Start()
	})
}

func (m *PubMultiWrapper) Publish(path string) IPublisherWrapper {
	// Multi might need to return the existing multi
	pub := globalReg.FindPublisher(path)
	if pub == nil {
		// First multi, initiate the register
		globalPublish(path, m)
		return m
	}
	// Publisher already exists
	multi, ok := pub.(IMultiPublisher)
	if !ok {
		panic("tried to register a multi on a path that another publisher type exists")
	}

	return multi.NewPublisher(m)
}

// Close is a permanent operation. You cannot reopen a publisher.
func (m *PubMultiWrapper) Close() error {
	m.pubLock.Lock()
	defer m.pubLock.Unlock()

	m.total--
	if m.total <= 0 {
		return m.IPublisher.Close()
	}
	return nil
}

// -- Wrapper funcs

func (m *PubMultiWrapper) NewPublisher(orig IPublisher) IMultiPublisher {
	m.pubLock.Lock()
	defer m.pubLock.Unlock()

	m.total++
	// We ignore the original publisher, and return a copy of our own.
	// TODO: Should we ensure the original has the same params as us?
	return m
}

func (m *PubMultiWrapper) TotalPublishing() int {
	m.pubLock.Lock()
	defer m.pubLock.Unlock()

	return m.total
}
