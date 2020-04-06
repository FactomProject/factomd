package events

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/modules/pubsub"
)

type RequestState int32

const (
	RequestState_HOLDING  RequestState = 0
	RequestState_ACCEPTED RequestState = 1
	RequestState_REJECTED RequestState = 2
)

func EmitEventFromMessage(pubState pubsub.IPubState, msg interfaces.IMsg, requestState RequestState) {
	pubRegistry := pubState.GetPubRegistry()
	switch msg.(type) {
	case *messages.CommitChainMsg:
		commitChainMsg := msg.(*messages.CommitChainMsg)
		commitChainEvent := &CommitChain{
			DBHeight:     pubState.GetLeaderHeight(),
			CommitChain:  commitChainMsg.CommitChain,
			RequestState: requestState,
		}
		pubRegistry.GetCommitChain().Write(commitChainEvent)
	case *messages.CommitEntryMsg:
		commitEntryMsg := msg.(*messages.CommitEntryMsg)
		commitEntryEvent := &CommitEntry{
			DBHeight:     pubState.GetLeaderHeight(),
			CommitEntry:  commitEntryMsg.CommitEntry,
			RequestState: requestState,
		}
		pubRegistry.GetCommitEntry().Write(commitEntryEvent)
	case *messages.RevealEntryMsg:
		revealEntryMsg := msg.(*messages.RevealEntryMsg)
		revealEntryEvent := &RevealEntry{
			DBHeight:     revealEntryMsg.Entry.GetDatabaseHeight(),
			RevealEntry:  revealEntryMsg.Entry,
			RequestState: requestState,
			MsgTimestamp: revealEntryMsg.Timestamp,
		}
		pubRegistry.GetRevealEntry().Write(revealEntryEvent)
	}
}

func EmitNodeMessageF(pubState pubsub.IPubState, messageCode NodeMessageCode, level Level, format string, values ...interface{}) {
	EmitNodeMessage(pubState, messageCode, level, fmt.Sprintf(format, values...))
}

func EmitNodeMessage(pubState pubsub.IPubState, messageCode NodeMessageCode, level Level, message string) {
	pubRegistry := pubState.GetPubRegistry()
	nodeMessage := &NodeMessage{
		MessageCode: messageCode,
		Level:       level,
		MessageText: message,
	}
	pubRegistry.GetNodeMessage().Write(nodeMessage)
}
