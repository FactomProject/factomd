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

type LiveFeedService interface {
	Start(state StateEventServices, config *util.FactomdConfig, factomParams *globals.FactomParams)
	ConfigSender(state StateEventServices, sender eventservices.EventSender)
}

type liveFeedService struct {
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

func NewLiveFeedService() LiveFeedService {
	liveFeedService := new(liveFeedService)
	liveFeedService.factomEventPublisher = pubsub.PubFactory.Threaded(p2p.StandardChannelSize).Publish("/live-feed", pubsub.PubMultiWrap())
	pubsub.SubFactory.PrometheusCounter("factomd_livefeed_total_events_published", "Number of events published by the factomd backend")
	return liveFeedService
}

func (liveFeedService *liveFeedService) Start(serviceState StateEventServices, config *util.FactomdConfig, factomParams *globals.FactomParams) {
	liveFeedService.parentState = serviceState
	liveFeedService.eventSender = eventservices.NewEventSender(config, factomParams)

	subRegistry := subregistry.New(serviceState.GetFactomNodeName())
	liveFeedService.subNodeMessage = subRegistry.NodeMessageChannel()
	liveFeedService.subCommitDBState = subRegistry.CommitDBStateChannel()
	liveFeedService.subDBAnchored = subRegistry.DBAnchoredChannel()
	liveFeedService.subCommitChain = subRegistry.CommitChainChannel()
	liveFeedService.subCommitEntry = subRegistry.CommitEntryChannel()
	liveFeedService.subRevealEntry = subRegistry.RevealEntryChannel()
	liveFeedService.subBlkSeq = subRegistry.BlkSeqChannel()
	go liveFeedService.processSubChannels()
	go liveFeedService.factomEventPublisher.Start()
}

func (liveFeedService *liveFeedService) ConfigSender(state StateEventServices, eventSender eventservices.EventSender) {
	liveFeedService.parentState = state
	liveFeedService.eventSender = eventSender
}

func (liveFeedService *liveFeedService) processSubChannels() {
	broadcastContent := liveFeedService.eventSender.GetBroadcastContent()
	sendStateChangeEvents := liveFeedService.eventSender.IsSendStateChangeEvents()

	for {
		select {
		case v := <-liveFeedService.subBlkSeq.Updates:
			liveFeedService.Send(eventservices.MapDBHT(v.(*event.DBHT), liveFeedService.GetStreamSource()))
		case v := <-liveFeedService.subCommitChain.Updates:
			commitChainEvent := v.(*event.CommitChain)
			if !sendStateChangeEvents || commitChainEvent.RequestState == event.RequestState_HOLDING {
				liveFeedService.Send(eventservices.MapCommitChain(commitChainEvent, liveFeedService.GetStreamSource()))
			} else {
				liveFeedService.Send(eventservices.MapCommitChainState(commitChainEvent, liveFeedService.GetStreamSource()))
			}
		case v := <-liveFeedService.subCommitEntry.Updates:
			commitEntryEvent := v.(*event.CommitEntry)
			if !sendStateChangeEvents || commitEntryEvent.RequestState == event.RequestState_HOLDING {
				liveFeedService.Send(eventservices.MapCommitEntry(commitEntryEvent, liveFeedService.GetStreamSource()))
			} else {
				liveFeedService.Send(eventservices.MapCommitEntryState(commitEntryEvent, liveFeedService.GetStreamSource()))
			}
		case v := <-liveFeedService.subRevealEntry.Updates:
			revealEntryEvent := v.(*event.RevealEntry)
			if !sendStateChangeEvents || revealEntryEvent.RequestState == event.RequestState_HOLDING {
				liveFeedService.Send(eventservices.MapRevealEntry(revealEntryEvent, liveFeedService.GetStreamSource(), broadcastContent))
			} else {
				liveFeedService.Send(eventservices.MapRevealEntryState(revealEntryEvent, liveFeedService.GetStreamSource()))
			}
		case v := <-liveFeedService.subCommitDBState.Updates:
			liveFeedService.Send(eventservices.MapCommitDBState(v.(*event.DBStateCommit), liveFeedService.GetStreamSource(), broadcastContent))
		case v := <-liveFeedService.subDBAnchored.Updates:
			liveFeedService.Send(eventservices.MapCommitDBAnchored(v.(*event.DBAnchored), liveFeedService.GetStreamSource()))
		case v := <-liveFeedService.subNodeMessage.Updates:
			nodeMessageEvent := v.(*event.NodeMessage)
			liveFeedService.Send(eventservices.MapNodeMessage(nodeMessageEvent, liveFeedService.GetStreamSource()))
			if nodeMessageEvent.MessageCode == event.NodeMessageCode_SHUTDOWN {
				break
			}
		}
	}
}

func (liveFeedService *liveFeedService) Send(factomEvent *eventmessages.FactomEvent) error {
	if factomEvent == nil {
		return nil
	}

	factomEvent.IdentityChainID = liveFeedService.parentState.GetIdentityChainID().Bytes()
	factomEvent.FactomNodeName = liveFeedService.parentState.GetFactomNodeName()
	liveFeedService.factomEventPublisher.Write(factomEvent)
	return nil
}

func (liveFeedService *liveFeedService) GetStreamSource() eventmessages.EventSource {
	if liveFeedService.parentState == nil {
		return -1
	}

	if liveFeedService.parentState.GetRunLeader() {
		return eventmessages.EventSource_LIVE
	} else {
		return eventmessages.EventSource_REPLAY_BOOT
	}
}
