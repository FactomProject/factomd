package eventinput_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/adminBlock"
	"github.com/PaulSnow/factom2d/common/directoryBlock"
	"github.com/PaulSnow/factom2d/common/directoryBlock/dbInfo"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/events/eventinput"
	"github.com/PaulSnow/factom2d/events/eventmessages/generated/eventmessages"
	"github.com/PaulSnow/factom2d/testHelper"
	"github.com/stretchr/testify/assert"
)

func TestEventInput_RegistrationEvent(t *testing.T) {
	payload := new(messages.CommitChainMsg)
	registrationEvent := eventinput.NewRegistrationEvent(eventmessages.EventSource_LIVE, payload)

	assert.NotNil(t, registrationEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, registrationEvent.GetStreamSource())
	assert.Equal(t, payload, registrationEvent.GetPayload())
}

func TestEventInput_StateChangeEventMsg(t *testing.T) {
	payload := new(messages.CommitChainMsg)
	stateChangeEvent := eventinput.NewStateChangeEvent(eventmessages.EventSource_LIVE, eventmessages.EntityState_ACCEPTED, payload)

	assert.NotNil(t, stateChangeEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, stateChangeEvent.GetStreamSource())
	assert.Equal(t, eventmessages.EntityState_ACCEPTED, stateChangeEvent.GetEntityState())
	assert.Equal(t, payload, stateChangeEvent.GetPayload())
}

func TestEventInput_DirectoryBlockEvent(t *testing.T) {
	dbState := new(mockDBState)
	directoryBlockEvent := eventinput.NewDirectoryBlockEvent(eventmessages.EventSource_LIVE, dbState)

	assert.NotNil(t, directoryBlockEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, directoryBlockEvent.GetStreamSource())
	assert.Equal(t, dbState, directoryBlockEvent.GetPayload())
}

func TestEventInput_ReplayDirectoryBlockEvent(t *testing.T) {
	payload := new(messages.DBStateMsg)
	directoryBlockEvent := eventinput.NewReplayDirectoryBlockEvent(eventmessages.EventSource_LIVE, payload)

	assert.NotNil(t, directoryBlockEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, directoryBlockEvent.GetStreamSource())
	assert.Equal(t, payload, directoryBlockEvent.GetPayload())
}

func TestEventInput_AnchorEvent(t *testing.T) {
	dirBlockInfo := testHelper.CreateTestDirBlockInfo(&dbInfo.DirBlockInfo{DBHeight: 910})
	anchorEvent := eventinput.NewAnchorEvent(eventmessages.EventSource_LIVE, dirBlockInfo)
	assert.NotNil(t, anchorEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, anchorEvent.GetStreamSource())
	assert.Equal(t, dirBlockInfo, anchorEvent.GetPayload())
}

func TestEventInput_ProcessListEventNewBlock(t *testing.T) {
	processListEvent := eventinput.ProcessListEventNewBlock(eventmessages.EventSource_LIVE, 2)

	assert.NotNil(t, processListEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, processListEvent.GetStreamSource())
	if assert.NotNil(t, processListEvent.GetProcessListEvent()) && assert.NotNil(t, processListEvent.GetProcessListEvent().GetNewBlockEvent()) {
		assert.Equal(t, uint32(2), processListEvent.GetProcessListEvent().GetNewBlockEvent().NewBlockHeight)
	}
}

func TestEventInput_ProcessListEventNewMinute(t *testing.T) {
	processListEvent := eventinput.ProcessListEventNewMinute(eventmessages.EventSource_LIVE, 2, 3)

	assert.NotNil(t, processListEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, processListEvent.GetStreamSource())
	if assert.NotNil(t, processListEvent.GetProcessListEvent()) && assert.NotNil(t, processListEvent.GetProcessListEvent().GetNewMinuteEvent()) {
		assert.Equal(t, uint32(2), processListEvent.GetProcessListEvent().GetNewMinuteEvent().NewMinute)
		assert.Equal(t, uint32(3), processListEvent.GetProcessListEvent().GetNewMinuteEvent().BlockHeight)
	}
}

func TestEventInput_NodeInfoMessage(t *testing.T) {
	nodeInfoEvent := eventinput.NodeInfoMessageF(eventmessages.NodeMessageCode_STARTED, "test: %s", "the node info")

	assert.NotNil(t, nodeInfoEvent)
	assert.Equal(t, eventmessages.EventSource_LIVE, nodeInfoEvent.GetStreamSource())
	if assert.NotNil(t, nodeInfoEvent.GetNodeMessage()) {
		assert.Equal(t, eventmessages.Level_INFO, nodeInfoEvent.GetNodeMessage().Level)
		assert.Equal(t, eventmessages.NodeMessageCode_STARTED, nodeInfoEvent.GetNodeMessage().MessageCode)
		assert.Equal(t, "test: the node info", nodeInfoEvent.GetNodeMessage().MessageText)
	}
}

func TestEventInput_NodeErrorMessage(t *testing.T) {
	nodeInfoEvent := eventinput.NodeErrorMessage(eventmessages.NodeMessageCode_SHUTDOWN, "test: %s", "the node error")

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
