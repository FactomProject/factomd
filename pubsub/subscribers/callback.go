package subscribers

import "github.com/Emyrk/pubsub/pubregistry"

// Callback allows an external function call to be bound to the Write()
// function of the subscriber.
type Callback struct {
	pubregistry.IPubSubscriber

	// BeforeWrite is called before a write. If an error is thrown, the
	// value is rejected by the subscriber
	BeforeWrite func(o interface{}) error
	AfterWrite  func(o interface{})
}

// NewCallback
//	Params:
//		subscriber
//		callback
func NewCallback(subscriber pubregistry.IPubSubscriber) *Callback {
	s := new(Callback)
	s.IPubSubscriber = subscriber
	// Default to no ops
	s.BeforeWrite = func(o interface{}) error { return nil }
	s.AfterWrite = func(o interface{}) {}

	return s
}

func (s *Callback) Write(o interface{}) {
	if s.BeforeWrite(o) != nil {
		return
	}
	s.IPubSubscriber.Write(o)
	s.AfterWrite(o)
}
