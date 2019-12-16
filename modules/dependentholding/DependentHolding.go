package dependentholding

import (
	"context"
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/pubsub"
)

type DependentHolding struct {
	common.Name
	inMsgs  *pubsub.SubChannel // incoming messages from BMV
	outMsgs pubsub.IPublisher  // outgoing messages to VMs
}

func NewDependentHolding(parent common.NamedObject, instance int) *DependentHolding {
	b := new(DependentHolding)
	b.NameInit(parent, fmt.Sprintf("dependentHolding%d", instance), reflect.TypeOf(b).String())
	b.inMsgs = pubsub.SubFactory.Channel(100) //.Subscribe("path?")
	// All dependent holdings in an fnode publish into one multiwrap
	path := pubsub.GetPath(b.Name.GetParentName(), "dependentholding", "msgout")
	b.outMsgs = pubsub.PubFactory.Threaded(100).Publish(path, pubsub.PubMultiWrap())
	return b
}

func (b *DependentHolding) Publish() {
	go b.outMsgs.Start()
}
func (b *DependentHolding) Subscribe() {
	// TODO: Find actual paths
	b.inMsgs = b.inMsgs.Subscribe(pubsub.GetPath(b.GetParentName(), "bmv", "rest"))
}

func (b *DependentHolding) ClosePublishing() {
	_ = b.outMsgs.Close()
}

func (b *DependentHolding) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-b.inMsgs.Channel():
			b.outMsgs.Write(data)
		}
	}
}
