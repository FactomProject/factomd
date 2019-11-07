package pubsub

// SubWrapCallback allows an external function call to be bound to the Write()
// function of the subscriber.
type SubWrapCallback struct {
	IPubSubscriber

	// BeforeWrite is called before a write. If an error is thrown, the
	// value is rejected by the subscriber
	BeforeWrite func(o interface{}) error
	AfterWrite  func(o interface{})
}

// NewCallback
//	Params:
//		subscriber
//		callback
func NewCallback(subscriber IPubSubscriber) *SubWrapCallback {
	s := new(SubWrapCallback)
	s.IPubSubscriber = subscriber
	// Default to no ops
	s.BeforeWrite = func(o interface{}) error { return nil }
	s.AfterWrite = func(o interface{}) {}

	return s
}

func (s *SubWrapCallback) write(o interface{}) {
	if s.BeforeWrite(o) != nil {
		return
	}
	s.IPubSubscriber.write(o)
	s.AfterWrite(o)
}

func (s *SubWrapCallback) Subscribe(path string) *SubWrapCallback {
	globalSubscribe(path, s)
	return s
}

func (s *SubWrapCallback) Wrap(sub IPubSubscriber) IPubSubscriber {
	s.IPubSubscriber = sub
	return s
}
