package events

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventInput_RegistrationEvent(t *testing.T) {
	payload := new(messages.CommitChainMsg)
	registrationEvent := NewRegistrationEvent(eventmessages.EventSource_LIVE, payload)

	assert.NotNil(t, registrationEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, registrationEvent.GetStreamSource())
	assert.Equal(t, payload, registrationEvent.GetPayload())
}

func TestEventInput_StateChangeEventMsg(t *testing.T) {
	payload := new(messages.CommitChainMsg)
	stateChangeEvent := NewStateChangeEventFromMsg(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, payload)

	assert.NotNil(t, stateChangeEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, stateChangeEvent.GetStreamSource())
	assert.Equal(t, eventmessages.EntityState_ACCEPTED, stateChangeEvent.GetEntityState())
	assert.Equal(t, payload, stateChangeEvent.GetPayload())
}

func TestEventInput_StateChangeEvent(t *testing.T) {
	dbState := new(mockDBState)
	stateChangeEvent := NewStateChangeEvent(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, dbState)

	assert.NotNil(t, stateChangeEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, stateChangeEvent.GetStreamSource())
	assert.Equal(t, eventmessages.EntityState_ACCEPTED, stateChangeEvent.GetEntityState())
	assert.Equal(t, dbState, stateChangeEvent.GetPayload())
}

func TestEventInput_ProcessInfoMessage(t *testing.T) {
	processInfoEvent := ProcessInfoEventF(eventmessages.EventSource_LIVE, eventmessages.ProcessCode_NEW_MINUTE, "test: %s", "the process info")

	assert.NotNil(t, processInfoEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, processInfoEvent.GetStreamSource())
	if assert.NotNil(t, processInfoEvent.GetProcessMessage()) {
		assert.Equal(t, eventmessages.Level_INFO, processInfoEvent.GetProcessMessage().Level)
		assert.Equal(t, eventmessages.ProcessCode_NEW_MINUTE, processInfoEvent.GetProcessMessage().ProcessCode)
		assert.Equal(t, "test: the process info", processInfoEvent.GetProcessMessage().MessageText)
	}
}

func TestEventInput_NodeInfoMessage(t *testing.T) {
	nodeInfoEvent := NodeInfoMessageF(eventmessages.NodeMessageCode_STARTED, "test: %s", "the node info")

	assert.NotNil(t, nodeInfoEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, nodeInfoEvent.GetStreamSource())
	if assert.NotNil(t, nodeInfoEvent.GetNodeMessage()) {
		assert.Equal(t, eventmessages.Level_INFO, nodeInfoEvent.GetNodeMessage().Level)
		assert.Equal(t, eventmessages.NodeMessageCode_STARTED, nodeInfoEvent.GetNodeMessage().MessageCode)
		assert.Equal(t, "test: the node info", nodeInfoEvent.GetNodeMessage().MessageText)
	}
}

func TestEventInput_NodeErrorMessage(t *testing.T) {
	nodeInfoEvent := NodeErrorMessage(eventmessages.NodeMessageCode_SHUTDOWN, "test: %s", "the node error")

	assert.NotNil(t, nodeInfoEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, nodeInfoEvent.GetStreamSource())
	if assert.NotNil(t, nodeInfoEvent.GetNodeMessage()) {
		assert.Equal(t, eventmessages.Level_ERROR, nodeInfoEvent.GetNodeMessage().Level)
		assert.Equal(t, eventmessages.NodeMessageCode_SHUTDOWN, nodeInfoEvent.GetNodeMessage().MessageCode)
		assert.Equal(t, "test: the node error", nodeInfoEvent.GetNodeMessage().MessageText)
	}
}

type mockDBState struct{}

func (*mockDBState) GetDirectoryBlock() interfaces.IDirectoryBlock {
	return &directoryBlock.DirectoryBlock{}
}
func (*mockDBState) GetAdminBlock() interfaces.IAdminBlock {
	return &adminBlock.AdminBlock{}
}
func (*mockDBState) GetFactoidBlock() interfaces.IFBlock {
	return &factoid.FBlock{}
}
func (*mockDBState) GetEntryCreditBlock() interfaces.IEntryCreditBlock {
	return &entryCreditBlock.ECBlock{}
}

func (*mockDBState) GetEntryBlocks() []interfaces.IEntryBlock {
	return []interfaces.IEntryBlock{}
}
func (*mockDBState) GetEntries() []interfaces.IEBEntry {
	return []interfaces.IEBEntry{}
}
