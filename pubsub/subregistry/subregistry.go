package subregistry

import (
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
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
	return p.newChannel(event.Path.Seq)
}

func (p *SubRegistry) BankChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.Bank)
}

func (p *SubRegistry) DirectoryChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.Directory)
}

func (p *SubRegistry) LeaderConfigChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.LeaderConfig)
}

func (p *SubRegistry) CommitChainChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.CommitChain)
}

func (p *SubRegistry) CommitEntryChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.CommitEntry)
}

func (p *SubRegistry) RevealEntryChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.RevealEntry)
}

func (p *SubRegistry) CommitDBStateChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.CommitDBState)
}

func (p *SubRegistry) DBAnchoredChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.DBAnchored)
}

func (p *SubRegistry) NodeMessageChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.NodeMessage)
}

func (p *SubRegistry) EOMTickerChannel() *pubsub.SubChannel {
	return p.newChannel(event.Path.EOM)
}
