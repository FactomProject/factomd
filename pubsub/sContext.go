package pubsub

// Context allows an external function call to be bound to the Done()
// function of the subscriber.
type Context struct {
	IPubSubscriber
	done func()
}

func NewContext(subscriber IPubSubscriber, done func()) *Context {
	s := new(Context)
	s.IPubSubscriber = subscriber
	s.done = done

	return s
}

func (s *Context) Done() {
	s.done()
	s.IPubSubscriber.Done()
}

func (s *Context) Subscribe(path string) *Context {
	globalSubscribe(path, s)
	return s
}
