package livefeed

import (
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/modules/livefeed/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/modules/livefeed/eventservices"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/pubsub/subregistry"
	"github.com/FactomProject/factomd/util"
)

type EventService interface {
	Start(state StateEventServices, config *util.FactomdConfig, factomParams *globals.FactomParams)
	ConfigSender(state StateEventServices, sender eventservices.EventSender)
}

type eventHub struct {
	parentState          StateEventServices
	eventSender          eventservices.EventSender
	factomEventPublisher pubsub.IPublisher
	subCommitChain       *pubsub.SubChannel
	subCommitEntry       *pubsub.SubChannel
	subRevealEntry       *pubsub.SubChannel
	subCommitDBState     *pubsub.SubChannel
	subDBAnchored        *pubsub.SubChannel
	subBlkSeq            *pubsub.SubChannel
	subNodeMessage       *pubsub.SubChannel
}

func NewEventService() EventService {
	eventEmitter := new(eventHub)
	eventEmitter.factomEventPublisher = pubsub.PubFactory.Threaded(p2p.StandardChannelSize).Publish("/live-feed", pubsub.PubMultiWrap())
	pubsub.SubFactory.PrometheusCounter("factomd_livefeed_total_events_published", "Number of events published by the factomd backend")
	return eventEmitter
}

func (eventEmitter *eventHub) Start(serviceState StateEventServices, config *util.FactomdConfig, factomParams *globals.FactomParams) {
	eventEmitter.parentState = serviceState
	eventEmitter.eventSender = eventservices.NewEventSender(config, factomParams)

	subRegistry := subregistry.New(serviceState.GetFactomNodeName())
	eventEmitter.subNodeMessage = subRegistry.NodeMessageChannel()
	eventEmitter.subCommitDBState = subRegistry.CommitDBStateChannel()
	eventEmitter.subDBAnchored = subRegistry.DBAnchoredChannel()
	eventEmitter.subCommitChain = subRegistry.CommitChainChannel()
	eventEmitter.subCommitEntry = subRegistry.CommitEntryChannel()
	eventEmitter.subRevealEntry = subRegistry.RevealEntryChannel()
	eventEmitter.subBlkSeq = subRegistry.BlkSeqChannel()
	go eventEmitter.processSubChannels()
}

func (eventEmitter *eventHub) ConfigSender(state StateEventServices, eventSender eventservices.EventSender) {
	eventEmitter.parentState = state
	eventEmitter.eventSender = eventSender
}

func (eventEmitter *eventHub) processSubChannels() {
	broadcastContent := eventEmitter.eventSender.GetBroadcastContent()
	sendStateChangeEvents := eventEmitter.eventSender.IsSendStateChangeEvents()

	for {
		select {
		case v := <-eventEmitter.subBlkSeq.Updates:
			eventEmitter.Send(eventservices.MapDBHT(v.(*event.DBHT), eventEmitter.GetStreamSource()))
		case v := <-eventEmitter.subCommitChain.Updates:
			commitChainEvent := v.(*event.CommitChain)
			if !sendStateChangeEvents || commitChainEvent.RequestState == event.RequestState_HOLDING {
				eventEmitter.Send(eventservices.MapCommitChain(commitChainEvent, eventEmitter.GetStreamSource()))
			} else {
				eventEmitter.Send(eventservices.MapCommitChainState(commitChainEvent, eventEmitter.GetStreamSource()))
			}
		case v := <-eventEmitter.subCommitEntry.Updates:
			commitEntryEvent := v.(*event.CommitEntry)
			if !sendStateChangeEvents || commitEntryEvent.RequestState == event.RequestState_HOLDING {
				eventEmitter.Send(eventservices.MapCommitEntry(commitEntryEvent, eventEmitter.GetStreamSource()))
			} else {
				eventEmitter.Send(eventservices.MapCommitEntryState(commitEntryEvent, eventEmitter.GetStreamSource()))
			}
		case v := <-eventEmitter.subRevealEntry.Updates:
			revealEntryEvent := v.(*event.RevealEntry)
			if !sendStateChangeEvents || revealEntryEvent.RequestState == event.RequestState_HOLDING {
				eventEmitter.Send(eventservices.MapRevealEntry(revealEntryEvent, eventEmitter.GetStreamSource(), broadcastContent))
			} else {
				eventEmitter.Send(eventservices.MapRevealEntryState(revealEntryEvent, eventEmitter.GetStreamSource()))
			}
		case v := <-eventEmitter.subCommitDBState.Updates:
			eventEmitter.Send(eventservices.MapCommitDBState(v.(*event.DBStateCommit), eventEmitter.GetStreamSource(), broadcastContent))
		case v := <-eventEmitter.subDBAnchored.Updates:
			eventEmitter.Send(eventservices.MapCommitDBAnchored(v.(*event.DBAnchored), eventEmitter.GetStreamSource()))
		case v := <-eventEmitter.subNodeMessage.Updates:
			nodeMessageEvent := v.(*event.NodeMessage)
			eventEmitter.Send(eventservices.MapNodeMessage(nodeMessageEvent, eventEmitter.GetStreamSource()))
			if nodeMessageEvent.MessageCode == event.NodeMessageCode_SHUTDOWN {
				break
			}
		}
	}
}

func (eventEmitter *eventHub) Send(factomEvent *eventmessages.FactomEvent) error {
	if factomEvent == nil {
		return nil
	}

	factomEvent.IdentityChainID = eventEmitter.parentState.GetIdentityChainID().Bytes()
	factomEvent.FactomNodeName = eventEmitter.parentState.GetFactomNodeName()
	eventEmitter.factomEventPublisher.Write(factomEvent)
	return nil
}

func (eventEmitter *eventHub) GetStreamSource() eventmessages.EventSource {
	if eventEmitter.parentState == nil {
		return -1
	}

	if eventEmitter.parentState.IsRunLeader() {
		return eventmessages.EventSource_LIVE
	} else {
		return eventmessages.EventSource_REPLAY_BOOT
	}
}
