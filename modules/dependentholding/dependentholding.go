package dependentholding

//go:generate go run ../../factomgenerate/generate.go

import (
	"context"
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/pubsubtypes"
	"github.com/FactomProject/factomd/generated"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/telemetry"
	"github.com/FactomProject/factomd/worker"
)

type heldMessage struct {
	dependentHash [32]byte
	offset        int
}

type DependentHolding struct {
	common.Name
	logging    interfaces.Log
	holding    map[[32]byte][]interfaces.IMsg
	dependents map[[32]byte]heldMessage // used to avoid duplicate entries & track position in holding

	// New DependentHolding
	outMsgs       *generated.Publish_PubBase_IMsg_type            // outgoing messages to leader
	inMsgs        *generated.Subscribe_ByChannel_HoldRequest_type // incoming messages from BMV
	metDependency *generated.Subscribe_ByChannel_Hash_type        // hash of met dependencies
}

func NewDependentHolding(parent *worker.Thread, instance int) *DependentHolding {
	b := new(DependentHolding)
	b.logging = parent.Log
	b.NameInit(parent, fmt.Sprintf("dependentHolding%d", instance), reflect.TypeOf(b).String())

	// publishers
	outPath := pubsub.GetPath(b.Name.GetParentName(), "leader", "fromDependentHolding")
	b.outMsgs = generated.Publish_PubBase_IMsg(pubsub.PubFactory.Threaded(100).Publish(outPath))

	// subscribers
	b.inMsgs = generated.Subscribe_ByChannel_HoldRequest(pubsub.SubFactory.Channel(100))
	b.metDependency = generated.Subscribe_ByChannel_HoldRequest(pubsub.SubFactory.Channel(100))
	return b
}

// access gauge w/ proper labels
func (b *DependentHolding) metric(msg interfaces.IMsg) telemetry.Gauge {
	return telemetry.MapSize.WithLabelValues(b.GetName(), msg.Label())
}

func (b *DependentHolding) Publish() {
	go b.outMsgs.Start()
}

func (b *DependentHolding) Subscribe() {
	// TODO: Find actual paths
	inPath := pubsub.GetPath(b.GetParentName(), "leader", "toDependentHolding")
	b.inMsgs.Subscribe(inPath)
	metPath := pubsub.GetPath(b.GetParentName(), "metDependency")
	b.metDependency.Subscribe(metPath)
}

func (b *DependentHolding) ClosePublishing() {
	_ = b.outMsgs.Close()
}

// Get a single msg from dependent holding
func (l *DependentHolding) GetDependentMsg(h [32]byte) interfaces.IMsg {
	d, ok := l.dependents[h]
	if !ok {
		return nil
	}
	m := l.holding[d.dependentHash][d.offset]
	return m
}

// remove a single msg from  dependent holding (done when we add it to the process list).
func (l *DependentHolding) RemoveDependentMsg(h [32]byte, reason string) {
	d, ok := l.dependents[h]
	if !ok {
		return
	}
	msg := l.holding[d.dependentHash][d.offset]
	l.holding[d.dependentHash][d.offset] = nil
	delete(l.dependents, h)
	l.logging.LogMessage("NewDependentHolding", "delete "+reason, msg)
	return
}

// Add a message to a dependent holding list
func (l *DependentHolding) Add(h [32]byte, msg interfaces.IMsg) bool {
	_, found := l.dependents[msg.GetMsgHash().Fixed()]
	if found {
		return false
	}
	l.metric(msg).Inc()
	l.logging.LogMessage("NewDependentHolding", fmt.Sprintf("add[%x]", h[:6]), msg)

	if l.holding[h] == nil {
		l.holding[h] = []interfaces.IMsg{msg}
	} else {
		l.holding[h] = append(l.holding[h], msg)
	}

	l.dependents[msg.GetMsgHash().Fixed()] = heldMessage{h, len(l.holding[h]) - 1}
	return true
}

func (b *DependentHolding) nextDependency(iMsg interfaces.IMsg) *[32]byte {
	var check chan error = make(chan error)
	switch iMsg.Type() {
	case constants.COMMIT_ENTRY_MSG, constants.COMMIT_CHAIN_MSG:
	case constants.REVEAL_ENTRY_MSG:
		b.checkCommits.Write(pubsubtypes.CommitRequest{iMsg, check})
		err := <-check
		if err != nil {
			return iMsg.GetHash().PFixed() // reveal waiting on commit
		}
	}
	return nil
}

func (b *DependentHolding) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-b.inMsgs.Channel():
			iMsg := data.(interfaces.IMsg)
			h := b.nextDependency(iMsg)
			if h != nil {
				b.Add(*h, iMsg) // add dependency
			} else {
				b.outMsgs.Write(iMsg) // send it on to process
			}
		}
	}
}
