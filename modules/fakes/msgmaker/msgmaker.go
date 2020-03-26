package msgmaker

import (
	"context"
	"math/rand"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages"

	"github.com/FactomProject/factomd/modules/pubsub"
	"go.uber.org/ratelimit"
)

// FakeMsgMaker will generate msgs at a given rate
type FakeMsgMaker struct {
	pubpath     string
	pub         pubsub.IPublisher
	rateLimiter ratelimit.Limiter
	replayLimit ratelimit.Limiter
}

func NewFakeMsgMaker(pub pubsub.IPublisher, rate int) *FakeMsgMaker {
	f := new(FakeMsgMaker)
	f.pubpath = pub.Path()
	f.pub = pub
	f.rateLimiter = ratelimit.New(rate)
	f.replayLimit = ratelimit.New(rate)

	return f
}

func (f *FakeMsgMaker) SetReplayRate(rate int) {
	f.replayLimit = ratelimit.New(rate)
}

func (f *FakeMsgMaker) Run(ctx context.Context, window time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		f.rateLimiter.Take()
		msg := new(messages.Bounce)
		slip := rand.Int63n(int64(window)*2) - int64(window)
		ts := time.Now().Add(time.Duration(slip))
		msg.Timestamp = primitives.NewTimestampFromSeconds(uint32(ts.Unix()))
		msg.Data = make([]byte, 32)
		rand.Read(msg.Data)

		f.pub.Write(msg)
	}
}

func (f *FakeMsgMaker) RunReplays(ctx context.Context, amt int) {
	hits := make(map[[32]byte]int) // TODO: Maybe make this not a mem leak?
	sub := pubsub.SubFactory.Channel(100).Subscribe(f.pubpath)
	c := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case omsg := <-c:
			msg := omsg.(interfaces.IMsg)
			if _, ok := hits[msg.GetMsgHash().Fixed()]; ok {
				continue
			}
			hits[msg.GetMsgHash().Fixed()] += 1
			f.replayLimit.Take()
			for i := 0; i < amt; i++ {
				f.pub.Write(msg)
			}
		}
	}
}
