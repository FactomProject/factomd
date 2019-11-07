package pubsub

type SubWrapContext struct {
	IPubSubscriber
	done func()
}

func NewContextWrap(done func()) *SubWrapContext {
	s := new(SubWrapContext)
	s.done = done

	return s
}

func (s *SubWrapContext) Wrap(sub IPubSubscriber) IPubSubscriber {
	s.IPubSubscriber = sub
	return s
}

func (s *SubWrapContext) Done() {
	s.done()
	s.IPubSubscriber.Done()
}

func (s *SubWrapContext) Subscribe(path string) *SubWrapContext {
	globalSubscribe(path, s)
	return s
}
