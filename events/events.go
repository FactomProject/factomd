package events

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
)

type Events interface {
	EmitRegistrationEvent(msg interfaces.IMsg)
	EmitStateChangeEvent(msg interfaces.IMsg, entityState eventmessages.EntityState)
	EmitDBStateEvent(dbState interfaces.IDBState, entityState eventmessages.EntityState)
	EmitDBAnchorEvent(dirBlockInfo interfaces.IDirBlockInfo)
	EmitProcessListEventNewBlock(newBlockHeight uint32)
	EmitProcessListEventNewMinute(newMinute int, blockHeight uint32)
	EmitNodeInfoMessage(messageCode eventmessages.NodeMessageCode, message string)
	EmitNodeInfoMessageF(messageCode eventmessages.NodeMessageCode, format string, values ...interface{})
	EmitNodeErrorMessage(messageCode eventmessages.NodeMessageCode, message string, values interface{})
}
