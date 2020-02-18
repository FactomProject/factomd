package pubsub

var SubFactory subFactory

type subFactory struct{}

func (subFactory) BEChannel(buffer int) *SubChannel { return NewBestEffortSubChannel(buffer) }
func (subFactory) Channel(buffer int) *SubChannel   { return NewSubChannel(buffer) }
func (subFactory) Value() *SubValue                 { return NewSubValue() }
func (subFactory) UnsafeValue() *UnsafeSubValue     { return NewUnsafeSubValue() }
func (subFactory) Counter() *SubCounter             { return NewSubCounter() }
func (subFactory) Config(update func(o map[string]interface{})) *SubConfig {
	return NewSubConfig(update)
}
func (subFactory) PrometheusCounter(name string, help ...string) *SubPrometheusCounter {
	return NewSubPrometheusCounter(name, help...)
}

var PubFactory pubFactory

type pubFactory struct{}

func (pubFactory) Base() *PubBase { return new(PubBase) }
func (pubFactory) RoundRobin(buffer int) *PubSelector {
	return NewPubSelector(buffer, &RoundRobinSelector{})
}
func (pubFactory) MsgSplit(buffer int) *PubSelector {
	return NewPubSelector(buffer, &MsgSplitSelector{})
}
func (pubFactory) Threaded(buffer int) *PubThreaded { return NewPubThreaded(buffer) }
