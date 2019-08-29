package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
)

func EmitRegistrationEvent(msg interfaces.IMsg, state *State) {
	if state.EventsService != nil {
		switch msg.(type) { // Do not fill the channel with message we don't want anyway (like EOM's)
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg:
			event := events.NewRegistrationEvent(GetStreamSource(state), msg)
			state.EventsService.Send(event)
		}
	}
}

func EmitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState, state *State) {
	if state.EventsService != nil {
		switch msg.(type) {
		case *messages.CommitChainMsg, *messages.CommitEntryMsg, *messages.RevealEntryMsg, *messages.DBStateMsg:
			event := events.NewStateChangeEvent(GetStreamSource(state), entityState, msg)
			state.EventsService.Send(event)
		}
	}
}

func GetStreamSource(state *State) eventmessages.StreamSource {
	if state.IsRunLeader() { // FIXME this is only true during startup
		return eventmessages.StreamSource_LIVE
	} else {
		return eventmessages.StreamSource_REPLAY
	}
}
