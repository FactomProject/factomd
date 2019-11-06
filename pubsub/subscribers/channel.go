package subscribers

type Channel struct {
	Base
	Updates chan interface{}
}

func NewChannelBasedSubscriber(buffer int) *Channel {
	s := new(Channel)
	s.Updates = make(chan interface{}, buffer)

	return s
}

func (s *Channel) Write(o interface{}) {
	s.Updates <- o
}

func (s *Channel) Done() {
	close(s.Updates)
}

// Sub side

func (s *Channel) Channel() <-chan interface{} {
	return s.Updates
}

func (s *Channel) Read() (v interface{}) {
	v = <-s.Updates
	return v
}

func (s *Channel) ReadWithFlag() (v interface{}, open bool) {
	v, open = <-s.Updates
	return v, open
}
