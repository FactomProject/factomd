package events

import (
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventinput"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/events/events_config"
)

type Events interface {
	EmitRegistrationEvent(msg interfaces.IMsg)
	EmitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState)
	EmitDBStateEvent(dbState interfaces.IDBState, entityState eventmessages.EntityState)
	EmitDBAnchorEvent(dirBlockInfo interfaces.IDirBlockInfo)
	EmitReplayStateChangeEvent(msg interfaces.IMsg, state eventmessages.EntityState)
	EmitProcessListEventNewBlock(newBlockHeight uint32)
	EmitProcessListEventNewMinute(newMinute int, blockHeight uint32)
	EmitNodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string)
	EmitNodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{})
	EmitNodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{})
}

type StateEventServices interface {
	GetRunState() runstate.RunState
	GetIdentityChainID() interfaces.IHash
	IsRunLeader() bool
	GetEvents() Events
}

type EventService interface {
	Send(event eventinput.EventInput) error
}

type EventServiceControl interface {
	GetBroadcastContent() events_config.BroadcastContent
	Shutdown()
	IsSendStateChangeEvents() bool
}
