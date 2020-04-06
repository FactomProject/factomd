package eventservices

import (
	"github.com/FactomProject/factomd/modules/events"
	"github.com/FactomProject/factomd/modules/livefeed/eventmessages/generated/eventmessages"
)

func MapCommitChain(commitChainEvent *events.CommitChain, eventSource eventmessages.EventSource) *eventmessages.FactomEvent {
	commitChain := commitChainEvent.CommitChain
	ecPubKey := commitChain.GetECPubKey().Fixed()
	sig := commitChain.GetSig()

	factomEvent := &eventmessages.FactomEvent_ChainCommit{
		ChainCommit: &eventmessages.ChainCommit{
			EntityState:          mapRequestState(commitChainEvent.RequestState),
			ChainIDHash:          commitChain.GetChainIDHash().Bytes(),
			EntryHash:            commitChain.GetEntryHash().Bytes(),
			Timestamp:            convertByteSlice6ToTimestamp(commitChain.GetMilliTime()),
			Credits:              uint32(commitChain.GetCredits()),
			EntryCreditPublicKey: ecPubKey[:],
			Signature:            sig[:],
			Version:              uint32(commitChain.GetVersion()),
			Weld:                 commitChain.GetWeld().Bytes(),
		},
	}
	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}

func MapCommitChainState(commitChainEvent *events.CommitChain, eventSource eventmessages.EventSource) *eventmessages.FactomEvent {
	commitChain := commitChainEvent.CommitChain
	factomEvent := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash:  commitChain.GetChainIDHash().Bytes(),
			EntityState: mapRequestState(commitChainEvent.RequestState),
		},
	}
	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}

func MapCommitEntryState(commitEntryEvent *events.CommitEntry, eventSource eventmessages.EventSource) *eventmessages.FactomEvent {
	commitEntry := commitEntryEvent.CommitEntry
	factomEvent := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash:  commitEntry.GetEntryHash().Bytes(),
			EntityState: mapRequestState(commitEntryEvent.RequestState),
		},
	}
	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}

func MapRevealEntryState(revealEntryEvent *events.RevealEntry, eventSource eventmessages.EventSource) *eventmessages.FactomEvent {
	revealEntry := revealEntryEvent.RevealEntry
	factomEvent := &eventmessages.FactomEvent_StateChange{
		StateChange: &eventmessages.StateChange{
			EntityHash:  revealEntry.GetHash().Bytes(),
			EntityState: mapRequestState(revealEntryEvent.RequestState),
		},
	}
	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}
