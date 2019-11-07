package pubsub

//type ContextWrap struct {
//	IPubSubscriber
//	done func()
//}
//
//func NewContextContextWrap(done func()) *ContextWrap {
//	s := new(ContextWrap)
//	s.done = done
//
//	return s
//}
//
//func (s *ContextWrap) Wrap(sub IPubSubscriber) IPubSubscriber {
//	s.IPubSubscriber = sub
//	return s
//}
//
//func (s *ContextWrap) Done() {
//	s.done()
//	s.IPubSubscriber.Done()
//}
