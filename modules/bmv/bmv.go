package bmv

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/modules/debugsettings"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/pubsub"
)

type BasicMessageValidator struct {
	// bootTime is used to set the
	bootTime time.Time
	// Anything before this timestamp is ignored
	preBootFilter time.Time

	// input is where all the incoming messages com from.
	input *pubsub.SubChannel
	// times is where each dblock's timestamp comes from for our time filter
	times *pubsub.SubChannel

	// The various publishers for various messages sorted by type
	groups []msgPub

	// The rest of the messages
	// pubList []pubsub.IPublisher
	// pubs    map[byte]pubsub.IPublisher
	output pubsub.IPublisher

	replay *MsgReplay

	NodeName string

	// Settings
	// Updates to regex filter
	inputRegexUpdates <-chan interface{}
	inputRegex        *regexp.Regexp
	netState          *debugsettings.Subscribe_ByValue_Bool_type
}

type msgPub struct {
	Name  string
	Types []byte
}

func NewBasicMessageValidator(nodeName string, replay *MsgReplay) *BasicMessageValidator {
	b := new(BasicMessageValidator)
	b.NodeName = nodeName

	// subscribed in b.Subscribe()
	b.input = pubsub.SubFactory.Channel(1000)
	b.times = pubsub.SubFactory.Channel(100)
	b.bootTime = time.Now()
	// 20min grace period
	b.preBootFilter = b.bootTime.Add(-20 * time.Minute)

	b.output = pubsub.PubFactory.Threaded(1000).Publish(pubsub.GetPath(b.NodeName, "bmv", "output"), pubsub.PubMultiWrap())

	b.replay = replay
	return b
}

func (b *BasicMessageValidator) Subscribe() {
	// TODO: Find actual paths
	b.input = b.input.Subscribe(pubsub.GetPath(b.NodeName, "bmv", "input"))
	b.times = b.times.Subscribe(pubsub.GetPath(b.NodeName, "blocktime"))

	sub := debugsettings.GetSettings(b.NodeName).InputRegexC()
	b.inputRegexUpdates = sub.Channel()

	b.netState = debugsettings.GetSettings(b.NodeName).NetStatOffV()
}

func (b *BasicMessageValidator) ClosePublishing() {
	_ = b.output.Close()
}

func (b *BasicMessageValidator) Run(ctx context.Context) {
	go b.output.Start()
	for {
		select {
		case <-ctx.Done():
			return
		case v := <-b.inputRegexUpdates:
			if re, ok := v.(*regexp.Regexp); ok {
				b.inputRegex = re
			}
		case blockTime := <-b.times.Updates:
			b.replay.Recenter(blockTime.(time.Time))
		case data := <-b.input.Updates:
			msg, ok := data.(interfaces.IMsg)
			if !ok {
				continue
			}

			if b.netState.Read() { // drop received message if he is off
				// fnode.State.LogMessage("NetworkInputs", "API drop, X'd by simCtrl", msg)
				continue // Toss any inputs from API
			}

			if msg.GetTimestamp().GetTime().Before(b.preBootFilter) {
				continue // Prior to our boot, we ignore
			}

			// Pre-Checks
			if msg.GetHash().IsHashNil() {
				// fnode.State.LogMessage("badEvents", "Nil hash from APIQueue", msg)
				continue
			}

			if b.replay.UpdateReplay(msg) < 0 {
				continue // Already seen
			}

			if b.inputRegexReject(msg) {
				continue // Input regex rejected msg
			}

			// TODO: Missing things that were here before and must be done somewhere:
			// 		- Block replay. Replays from previous blocks. We could bootstrap
			//			them into the existing replay filter here. We need to bootstrap the times anyway.

			if msg.WellFormed() {
				TotalMessagesReceived.WithLabelValues(string(msg.Type())).Inc()
				b.Write(msg)
			}
		}
	}
}

func (b *BasicMessageValidator) Write(msg interfaces.IMsg) {
	b.output.Write(msg)
}

// TODO: The prior regex check also included state like leader height and minute. Those were stripped. Is that ok?
// inputRegexReject allows the developer to drop certain messages. If the message is supposed to be dropped, the
// return is true. The default case is a return of 'false', meaning no messages are targeted to be dropped.
func (b *BasicMessageValidator) inputRegexReject(msg interfaces.IMsg) bool {
	if b.inputRegex != nil {
		t := ""
		if mm, ok := msg.(*messages.MissingMsgResponse); ok {
			t = fmt.Sprintf(mm.MsgResponse.String())
		} else {
			t = fmt.Sprintf(msg.String())
		}

		if mm, ok := msg.(*messages.MissingMsgResponse); ok {
			if eom, ok := mm.MsgResponse.(*messages.EOM); ok {
				t2 := fmt.Sprintf(eom.String())
				messageResult := b.inputRegex.MatchString(t2)
				if messageResult {
					// fnode.State.LogMessage("NetworkInputs", "Drop, matched filter Regex", msg)
					return true
				}
			}
		}
		messageResult := b.inputRegex.MatchString(t)
		if messageResult {
			// fnode.State.LogMessage("NetworkInputs", "Drop, matched filter Regex", msg)
			return true
		}
	}
	return false
}
