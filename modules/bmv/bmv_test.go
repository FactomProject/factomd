package bmv

import (
	"context"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/pubsub"

	"github.com/FactomProject/factomd/modules/fakes/blockmaker"
	"github.com/FactomProject/factomd/modules/fakes/msgmaker"
)

func TestBasicMessageValidator_SingleRun(t *testing.T) {
	bmv := NewBasicMessageValidator()
	blocks := blockmaker.NewFakeBlockMaker()
	msgs := msgmaker.NewFakeMsgMaker("/msgs", 10)
	bmv.Subscribe()

	// Bootstrap the bmv replay filter
	ctx, cancel := context.WithCancel(context.Background())
	sub := pubsub.SubFactory.Channel(100).Subscribe("/bmv")

	// Get our msgs coming through
	go bmv.Run(ctx)
	blocks.Bootstrap(6, time.Now().Add(-62*time.Minute))

	// Get some msgs going
	go msgs.Run(ctx, 10)
	go msgs.RunReplays(ctx, 3) // Get replays going

	hits := make(map[[32]byte]int)
	for i := 0; i < 25; i++ { // Run for 25 unique messages
		select {
		case o := <-sub.Updates:
			msg := o.(interfaces.IMsg)
			_, ok := hits[msg.GetMsgHash().Fixed()]
			if ok {
				t.Error("replay msg got through!")
			} else {
				hits[msg.GetMsgHash().Fixed()] += 1
			}
		default:
			i--
			time.Sleep(25 * time.Millisecond)
		}
	}
	cancel()

}
