package pubregistry

import (
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
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
	p.BlkSeq = p.newPublisher(event.Path.Seq)
	p.Bank = p.newPublisher(event.Path.Bank)
	p.Directory = p.newPublisher(event.Path.Directory)
	p.LeaderConfig = p.newPublisher(event.Path.LeaderConfig)
	p.CommitChain = p.newPublisher(event.Path.CommitChain)
	p.CommitEntry = p.newPublisher(event.Path.CommitEntry)
	p.RevealEntry = p.newPublisher(event.Path.RevealEntry)
	p.CommitDBState = p.newPublisher(event.Path.CommitDBState)
	p.DBAnchored = p.newPublisher(event.Path.DBAnchored)
	p.NodeMessage = p.newPublisher(event.Path.NodeMessage)
	p.ProcessListInfo = p.newPublisher(event.Path.ProcessListInfo)
	p.StateUpdate = p.newPublisher(event.Path.StateUpdate)
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
	return p.DBAnchored
}
