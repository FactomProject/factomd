package eventservices

import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/modules/livefeed/eventconfig"
	"github.com/FactomProject/factomd/modules/livefeed/eventmessages/generated/eventmessages"
	"github.com/gogo/protobuf/types"
	"time"
)

func MapCommitDBState(dbStateCommitEvent *events.DBStateCommit, eventSource eventmessages.EventSource,
	broadcastContent eventconfig.BroadcastContent) *eventmessages.FactomEvent {
	dbState := dbStateCommitEvent.DBState
	shouldIncludeContent := broadcastContent > eventconfig.BroadcastNever
	factomEvent := &eventmessages.FactomEvent_DirectoryBlockCommit{DirectoryBlockCommit: &eventmessages.DirectoryBlockCommit{
		DirectoryBlock:    mapDirectoryBlock(dbState.GetDirectoryBlock()),
		AdminBlock:        MapAdminBlock(dbState.GetAdminBlock()),
		FactoidBlock:      mapFactoidBlock(dbState.GetFactoidBlock()),
		EntryCreditBlock:  mapEntryCreditBlock(dbState.GetEntryCreditBlock()),
		EntryBlocks:       mapEntryBlocks(dbState.GetEntryBlocks()),
		EntryBlockEntries: mapEntryBlockEntries(dbState.GetEntries(), shouldIncludeContent),
	}}

	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}

func MapCommitDBAnchored(dbAnchoredEvent *events.DBAnchored, eventSource eventmessages.EventSource) *eventmessages.FactomEvent {
	dirBlockInfo := dbAnchoredEvent.DirBlockInfo
	factomEvent := &eventmessages.FactomEvent_DirectoryBlockAnchor{
		DirectoryBlockAnchor: &eventmessages.DirectoryBlockAnchor{
			DirectoryBlockHash:            dirBlockInfo.DatabaseSecondaryIndex().Bytes(),
			DirectoryBlockMerkleRoot:      dirBlockInfo.GetDBMerkleRoot().Bytes(),
			BlockHeight:                   dirBlockInfo.GetDatabaseHeight(),
			Timestamp:                     ConvertTimeToTimestamp(dirBlockInfo.GetTimestamp().GetTime()),
			BtcTxHash:                     dirBlockInfo.GetBTCTxHash().Bytes(),
			BtcTxOffset:                   uint32(dirBlockInfo.GetBTCTxOffset()),
			BtcBlockHeight:                uint32(dirBlockInfo.GetBTCBlockHeight()),
			BtcBlockHash:                  dirBlockInfo.GetBTCBlockHash().Bytes(),
			BtcConfirmed:                  dirBlockInfo.GetBTCConfirmed(),
			EthereumAnchorRecordEntryHash: dirBlockInfo.GetEthereumAnchorRecordEntryHash().Bytes(),
			EthereumConfirmed:             dirBlockInfo.GetEthereumConfirmed(),
		},
	}

	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}

func MapDBHT(dbht *events.DBHT, eventSource eventmessages.EventSource) *eventmessages.FactomEvent {
	factomEvent := &eventmessages.FactomEvent_ProcessListEvent{}
	if dbht.Minute == 0 {
		factomEvent.ProcessListEvent = &eventmessages.ProcessListEvent{
			ProcessListEvent: &eventmessages.ProcessListEvent_NewBlockEvent{
				NewBlockEvent: &eventmessages.NewBlockEvent{
					NewBlockHeight: dbht.DBHeight,
				},
			},
		}
	} else {
		factomEvent.ProcessListEvent = &eventmessages.ProcessListEvent{
			ProcessListEvent: &eventmessages.ProcessListEvent_NewMinuteEvent{
				NewMinuteEvent: &eventmessages.NewMinuteEvent{
					BlockHeight: dbht.DBHeight,
					NewMinute:   uint32(dbht.Minute),
				},
			},
		}
	}

	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}

func mapRequestState(state events.RequestState) eventmessages.EntityState {
	switch state {
	case events.RequestState_HOLDING:
		return eventmessages.EntityState_REQUESTED
	case events.RequestState_ACCEPTED:
		return eventmessages.EntityState_ACCEPTED
	case events.RequestState_REJECTED:
		return eventmessages.EntityState_REJECTED
	}
	panic(fmt.Sprintf("Unknown request state %v", state))
}

func MapNodeMessage(nodeMessageEvent *events.NodeMessage, eventSource eventmessages.EventSource) *eventmessages.FactomEvent {
	factomEvent := &eventmessages.FactomEvent_NodeMessage{
		NodeMessage: &eventmessages.NodeMessage{
			MessageCode: mapNodeEventMessageCode(nodeMessageEvent.MessageCode),
			Level:       mapNodeEventLevel(nodeMessageEvent.Level),
			MessageText: nodeMessageEvent.MessageText,
		}}
	return &eventmessages.FactomEvent{
		EventSource: eventSource,
		Event:       factomEvent,
	}
}

func mapNodeEventLevel(level events.Level) eventmessages.Level {
	switch level {
	case events.Level_INFO:
		return eventmessages.Level_INFO
	case events.Level_ERROR:
		return eventmessages.Level_ERROR
	case events.Level_WARNING:
		return eventmessages.Level_WARNING
	}
	panic(fmt.Sprintf("Unknown level %v", level))
}

func mapNodeEventMessageCode(code events.NodeMessageCode) eventmessages.NodeMessageCode {
	switch code {
	case events.NodeMessageCode_STARTED:
		return eventmessages.NodeMessageCode_STARTED
	case events.NodeMessageCode_SYNCED:
		return eventmessages.NodeMessageCode_SYNCED
	case events.NodeMessageCode_GENERAL:
		return eventmessages.NodeMessageCode_GENERAL
	case events.NodeMessageCode_SHUTDOWN:
		return eventmessages.NodeMessageCode_SHUTDOWN
	}
	panic(fmt.Sprintf("Unknown NodeMessageCode %v", code))
}

func convertByteSlice6ToTimestamp(milliTime *primitives.ByteSlice6) *types.Timestamp {
	// TODO Is there an easier way to do this?
	slice8 := make([]byte, 8)
	copy(slice8[2:], milliTime[:])
	millis := int64(binary.BigEndian.Uint64(slice8))
	t := time.Unix(0, millis*1000000)
	return ConvertTimeToTimestamp(t)
}

func ConvertTimeToTimestamp(t time.Time) *types.Timestamp {
	return &types.Timestamp{Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}
}
