package pubsub

// SubEmbedded is for making subscribers outside this package.
type SubEmbedded struct {
	Done           func()
	Write          func(o interface{})
	SetUnsubscribe func(unsub func())

	// Some default fields
	SubbedPublisher IReadOnlyPublisher
}

func (s *SubEmbedded) setUnsubscribe(unsub func()) { s.SetUnsubscribe(unsub) }
func (s *SubEmbedded) write(o interface{})         { s.Write(o) }
func (s *SubEmbedded) done()                       { s.Done() }

// Defaults
func (s *SubEmbedded) setPublisher(pub IReadOnlyPublisher) { s.SubbedPublisher = pub }
func (s *SubEmbedded) Publisher() IReadOnlyPublisher       { return s.SubbedPublisher }
