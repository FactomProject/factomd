package state

import (
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
)

// NOTE: a temporary home for publishers
// in the future it is likely we wouldn't want these accessible under fnode.State struct

type Pub struct {
	BlkSeq       pubsub.IPublisher
	Bank         pubsub.IPublisher
	Directory    pubsub.IPublisher
	EOMTicker    pubsub.IPublisher
	LeaderConfig pubsub.IPublisher
	AuthoritySet pubsub.IPublisher
}

func (s *State) mkPath(name string) string {
	return pubsub.GetPath(s.GetFactomNodeName(), name)
}

func (s *State) newPublisher(name string) pubsub.IPublisher {
	p := pubsub.PubFactory.Threaded(100).Publish(s.mkPath(name))
	go p.Start()
	return p
}

func (s *State) BindPublishers() {
	s.Publish()

	// MoveStateToHeight
	s.Pub.BlkSeq = s.newPublisher(event.Path.Seq)
	s.Pub.Bank = s.newPublisher(event.Path.Bank)
	s.Pub.Directory = s.newPublisher(event.Path.Directory)
	s.Pub.LeaderConfig = s.newPublisher(event.Path.LeaderConfig)
	s.Pub.AuthoritySet = s.newPublisher(event.Path.AuthoritySet)
}
