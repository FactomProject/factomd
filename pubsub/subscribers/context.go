package subscribers

import "github.com/Emyrk/pubsub/pubregistry"

// Context allows an external function call to be bound to the Done()
// function of the subscriber.
type Context struct {
	pubregistry.IPubSubscriber
	done func()
}

func NewContext(subscriber pubregistry.IPubSubscriber, done func()) *Context {
	s := new(Context)
	s.IPubSubscriber = subscriber
	s.done = done

	return s
}

func (s *Context) Done() {
	s.done()
	s.IPubSubscriber.Done()
}
