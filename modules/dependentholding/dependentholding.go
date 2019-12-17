package dependentholding

import (
	"context"
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/generated"
	"github.com/FactomProject/factomd/pubsub"
)

type heldMessage struct {
	dependentHash [32]byte
	offset        int
}

type DependentHolding struct {
	common.Name

	holding    map[[32]byte][]interfaces.IMsg
	dependents map[[32]byte]heldMessage // used to avoid duplicate entries & track position in holding

	// New DependentHolding
	inMsgs  *generated.Subscribe_ByChannel_IMsg_type // incoming messages from BMV
	outMsgs *generated.Publish_PubBase_IMsg_type     // outgoing messages to VMs
	//fctMessages        *generated.Publish_PubBase_IMsg_type
	//gossipMessages     *generated.Publish_PubBase_IMsg_type
	//heights            *generated.Subscribe_ByChannel_DBHT_type
	//metDependencyHashs *generated.Subscribe_ByChannel_Hash_type
	//chainReveals       *generated.Subscribe_ByChannel_Hash_type
	//commits            *generated.Subscribe_ByChannel_Hash_type

}

func NewDependentHolding(parent common.NamedObject, instance int) *DependentHolding {
	b := new(DependentHolding)
	b.NameInit(parent, fmt.Sprintf("dependentHolding%d", instance), reflect.TypeOf(b).String())
	b.inMsgs = generated.Subscribe_ByChannel_IMsg(pubsub.SubFactory.Channel(100)) //.Subscribe("path?")
	// All dependent holdings in an fnode publish into one multiwrap
	path := pubsub.GetPath(b.Name.GetParentName(), "dependentholding", "msgout")
	b.outMsgs = generated.Publish_PubBase_IMsg(pubsub.PubFactory.Threaded(100).Publish(path, pubsub.PubMultiWrap()))
	return b
}

func (b *DependentHolding) Publish() {
	go b.outMsgs.Start()
}
func (b *DependentHolding) Subscribe() {
	// TODO: Find actual paths
	b.inMsgs.SubChannel.Subscribe(pubsub.GetPath(b.GetParentName(), "bmv", "rest"))
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
			data2 := data.(interfaces.IMsg)
			b.outMsgs.Write(data2)
		}
	}
}
