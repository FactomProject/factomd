package pubregistry

import (
	"github.com/FactomProject/factomd/modules/events"
	"github.com/FactomProject/factomd/modules/pubsub"
)

type PubRegistry struct {
	BlkSeq          pubsub.IPublisher
	Bank            pubsub.IPublisher
	Directory       pubsub.IPublisher
	EOMTicker       pubsub.IPublisher
	LeaderConfig    pubsub.IPublisher
	CommitChain     pubsub.IPublisher
	CommitEntry     pubsub.IPublisher
	RevealEntry     pubsub.IPublisher
	CommitDBState   pubsub.IPublisher
	DBAnchored      pubsub.IPublisher
	NodeMessage     pubsub.IPublisher
	ProcessListInfo pubsub.IPublisher
	StateUpdate     pubsub.IPublisher
	AuthoritySet    pubsub.IPublisher
	Blocktime       pubsub.IPublisher
	factomNodeName  string
}

func New(factomNodeName string) *PubRegistry {
	p := &PubRegistry{
		factomNodeName: factomNodeName,
	}
	p.bindPublishers()
	return p
}

func (p *PubRegistry) mkPath(name string) string {
	return pubsub.GetPath(p.factomNodeName, name)
}

func (p *PubRegistry) newPublisher(name string) pubsub.IPublisher {
	publisher := pubsub.PubFactory.Threaded(100).Publish(p.mkPath(name))
	go publisher.Start()
	return publisher
}

func (p *PubRegistry) bindPublishers() {
	p.BlkSeq = p.newPublisher(events.Path.Seq)
	p.Bank = p.newPublisher(events.Path.Bank)
	p.Directory = p.newPublisher(events.Path.Directory)
	p.LeaderConfig = p.newPublisher(events.Path.LeaderConfig)
	p.CommitChain = p.newPublisher(events.Path.CommitChain)
	p.CommitEntry = p.newPublisher(events.Path.CommitEntry)
	p.RevealEntry = p.newPublisher(events.Path.RevealEntry)
	p.CommitDBState = p.newPublisher(events.Path.CommitDBState)
	p.DBAnchored = p.newPublisher(events.Path.DBAnchored)
	p.NodeMessage = p.newPublisher(events.Path.NodeMessage)
	p.ProcessListInfo = p.newPublisher(events.Path.ProcessListInfo)
	p.StateUpdate = p.newPublisher(events.Path.StateUpdate)
	p.AuthoritySet = p.newPublisher(events.Path.AuthoritySet)
	p.Blocktime = p.newPublisher("Blocktime")
}

func (p PubRegistry) GetBlkSeq() pubsub.IPublisher {
	return p.BlkSeq
}

func (p PubRegistry) GetBank() pubsub.IPublisher {
	return p.Bank
}

func (p PubRegistry) GetDirectory() pubsub.IPublisher {
	return p.Directory
}

func (p PubRegistry) GetEOMTicker() pubsub.IPublisher {
	return p.EOMTicker
}

func (p PubRegistry) GetLeaderConfig() pubsub.IPublisher {
	return p.LeaderConfig
}

func (p PubRegistry) GetCommitChain() pubsub.IPublisher {
	return p.CommitChain
}

func (p PubRegistry) GetCommitEntry() pubsub.IPublisher {
	return p.CommitEntry
}

func (p PubRegistry) GetRevealEntry() pubsub.IPublisher {
	return p.RevealEntry
}

func (p PubRegistry) GetCommitDBState() pubsub.IPublisher {
	return p.CommitDBState
}

func (p PubRegistry) GetDBAnchored() pubsub.IPublisher {
	return p.DBAnchored
}

func (p PubRegistry) GetNodeMessage() pubsub.IPublisher {
	return p.NodeMessage
}

func (p PubRegistry) GetBlocktime() pubsub.IPublisher {
	return p.Blocktime
}
