// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package msgsupport

//https://docs.google.com/spreadsheets/d/1wy9JDEqyM2uRYhZ6Y1e9C3hIDm2prIILebztQ5BGlr8/edit#gid=1997221100

import (
	"fmt"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"

	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/messages/electionMsgs"
)

func init() {
	// KLUDGE: a way to allow logs from p2p/connection.go
	messages.FP = UnmarshalMessage
}

func UnmarshalMessage(data []byte) (interfaces.IMsg, error) {
	_, msg, err := UnmarshalMessageData(data)
	return msg, err
}

func CreateMsg(messageType byte) interfaces.IMsg {
	switch messageType {
	case constants.EOM_MSG:
		return new(messages.EOB)
	case constants.ACK_MSG:
		return new(messages.Ack)
	case constants.COMMIT_CHAIN_MSG:
		return new(messages.CommitChainMsg)
	case constants.COMMIT_ENTRY_MSG:
		return new(messages.CommitEntryMsg)
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		return new(messages.DirectoryBlockSignature)
	case constants.FACTOID_TRANSACTION_MSG:
		return new(messages.FactoidTransaction)
	case constants.HEARTBEAT_MSG:
		return new(messages.Heartbeat)
	case constants.MISSING_MSG:
		return new(messages.MissingMsg)
	case constants.MISSING_MSG_RESPONSE:
		return new(messages.MissingMsgResponse)
	case constants.MISSING_DATA:
		return new(messages.MissingData)
	case constants.DATA_RESPONSE:
		return new(messages.DataResponse)
	case constants.REVEAL_ENTRY_MSG:
		return new(messages.RevealEntryMsg)
	case constants.REQUEST_BLOCK_MSG:
		return new(messages.RequestBlock)
	case constants.DBSTATE_MISSING_MSG:
		return new(messages.DBStateMissing)
	case constants.DBSTATE_MSG:
		return new(messages.DBStateMsg)
	case constants.ADDSERVER_MSG:
		return new(messages.AddServerMsg)
	case constants.CHANGESERVER_KEY_MSG:
		return new(messages.ChangeServerKeyMsg)
	case constants.REMOVESERVER_MSG:
		return new(messages.RemoveServerMsg)
	case constants.BOUNCE_MSG:
		return new(messages.Bounce)
	case constants.BOUNCEREPLY_MSG:
		return new(messages.BounceReply)
	case constants.SYNC_MSG:
		return new(electionMsgs.SyncMsg)
	case constants.VOLUNTEERAUDIT:
		msg := new(electionMsgs.FedVoteVolunteerMsg)
		msg.SetFullBroadcast(true)
		return msg
	case constants.VOLUNTEERPROPOSAL:
		msg := new(electionMsgs.FedVoteProposalMsg)
		msg.SetFullBroadcast(true)
		return msg
	case constants.VOLUNTEERLEVELVOTE:
		msg := new(electionMsgs.FedVoteLevelMsg)
		msg.SetFullBroadcast(true)
		return msg
	default:
		return nil
	}
}

func UnmarshalMessageData(data []byte) (newdata []byte, msg interfaces.IMsg, err error) {
	if data == nil {
		return nil, nil, fmt.Errorf("No data provided")
	}
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("No data provided")
	}
	messageType := data[0]
	msg = CreateMsg(messageType)
	if msg == nil {
		fmt.Printf("***** Marshal Failed to create message for %d %s", messageType, constants.MessageName(messageType))
		return data, nil, fmt.Errorf("Unknown message type %d %x", messageType, data[0])
	}

	newdata, err = msg.UnmarshalBinaryData(data[:])
	if err != nil {
		fmt.Printf("***** Marshal Failed to unmarshal %d %s %x\n",
			messageType,
			constants.MessageName(messageType),
			data)
		return data, nil, err
	}

	return newdata, msg, nil

}

// GeneralFactory is used to get around package import loops.
type GeneralFactory struct {
}

var _ interfaces.IGeneralMsg = (*GeneralFactory)(nil)

func (GeneralFactory) CreateMsg(messageType byte) interfaces.IMsg {
	return CreateMsg(messageType)
}

func (GeneralFactory) MessageName(Type byte) string {
	return constants.MessageName(Type)
}

func (GeneralFactory) UnmarshalMessageData(data []byte) (newdata []byte, msg interfaces.IMsg, err error) {
	return UnmarshalMessageData(data)
}

func (GeneralFactory) UnmarshalMessage(data []byte) (interfaces.IMsg, error) {
	return UnmarshalMessage(data)
}
