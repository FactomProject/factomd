package blockmaker

import (
	"time"

	"github.com/FactomProject/factomd/pubsub"
)

// FakeBlockMaker will generate timestamps as if factom blocks are being
// generated.
type FakeBlockMaker struct {
	pub pubsub.IPublisher
}

func NewFakeBlockMaker() *FakeBlockMaker {
	f := new(FakeBlockMaker)
	f.pub = pubsub.PubFactory.Base().Publish("/blocktime")

	return f
}

func (f *FakeBlockMaker) Publish(blockTime time.Time) {
	f.pub.Write(blockTime)
}

// Bootstrap will send over pastN blocks from the given timestamp
func (f *FakeBlockMaker) Bootstrap(pastN int, current time.Time) {
	for i := pastN; i >= 0; i-- {
		f.Publish(current.Add(-1 * time.Duration(i) * time.Minute * 10))
	}
}
