package subregistry

import (
	"github.com/FactomProject/factomd/modules/internalevents"
	"github.com/FactomProject/factomd/modules/pubsub"
)

type SubRegistry struct {
	factomNodeName string
}

func New(factomNodeName string) *SubRegistry {
	p := &SubRegistry{
		factomNodeName: factomNodeName,
	}
	return p
}

func (p *SubRegistry) mkPath(name string) string {
	return pubsub.GetPath(p.factomNodeName, name)
}

func (p *SubRegistry) newChannel(name string) *pubsub.SubChannel {
	return pubsub.SubFactory.Channel(4096).Subscribe(p.mkPath(name))
}

func (p *SubRegistry) BlkSeqChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.Seq)
}

func (p *SubRegistry) BankChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.Bank)
}

func (p *SubRegistry) DirectoryChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.Directory)
}

func (p *SubRegistry) LeaderConfigChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.LeaderConfig)
}

func (p *SubRegistry) CommitChainChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.CommitChain)
}

func (p *SubRegistry) CommitEntryChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.CommitEntry)
}

func (p *SubRegistry) RevealEntryChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.RevealEntry)
}

func (p *SubRegistry) CommitDBStateChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.CommitDBState)
}

func (p *SubRegistry) DBAnchoredChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.DBAnchored)
}

func (p *SubRegistry) NodeMessageChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.NodeMessage)
}

func (p *SubRegistry) EOMTickerChannel() *pubsub.SubChannel {
	return p.newChannel(internalevents.Path.EOM)
}
