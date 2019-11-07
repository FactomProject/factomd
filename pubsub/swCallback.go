package pubsub

// SubWrapCallback allows an external function call to be bound to the Write()
// function of the subscriber.
type SubWrapCallback struct {
	IPubSubscriber
	SubWrapBase

	// BeforeWrite is called before a write. If an error is thrown, the
	// value is rejected by the subscriber
	BeforeWrite func(o interface{}) error
	AfterWrite  func(o interface{})
}

// NewCallback
func NewCallback(before func(o interface{}) error, after func(o interface{})) *SubWrapCallback {
	s := new(SubWrapCallback)

	if before == nil {
		s.BeforeWrite = func(o interface{}) error { return nil }
	} else {
		s.BeforeWrite = before
	}
	if after == nil {
		s.AfterWrite = func(o interface{}) {}
	} else {
		s.AfterWrite = after
	}

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
	s.SetBase(sub)
	s.IPubSubscriber = sub
	return s
}
