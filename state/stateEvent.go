package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events"
	"github.com/FactomProject/factomd/events/eventmessages"
)

func emitEvent(eventSource eventmessages.EventSource, msg interfaces.IMsg, state *State) {
	if state.EventsService != nil {
		event := events.EventFromNetworkMessage(eventSource, msg)
		state.EventsService.Send(event)
	}
}
