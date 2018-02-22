// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

//https://docs.google.com/spreadsheets/d/1wy9JDEqyM2uRYhZ6Y1e9C3hIDm2prIILebztQ5BGlr8/edit#gid=1997221100

import (
	"errors"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"

	log "github.com/sirupsen/logrus"
)

// packageLogger is the general logger for all message related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{"package": "messages"})

func UnmarshalMessage(data []byte) (interfaces.IMsg, error) {
	_, msg, err := UnmarshalMessageData(data)
	return msg, err
}

func UnmarshalMessageData(data []byte) (newdata []byte, msg interfaces.IMsg, err error) {
	if data == nil {
		return nil, nil, fmt.Errorf("No data provided")
	}
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("No data provided")
	}
	messageType := data[0]

	switch messageType {
	case constants.EOM_MSG:
		msg = new(EOM)
	case constants.ACK_MSG:
		msg = new(Ack)
	case constants.AUDIT_SERVER_FAULT_MSG:
		msg = new(AuditServerFault)
	case constants.FED_SERVER_FAULT_MSG:
		msg = new(ServerFault)
	case constants.FULL_SERVER_FAULT_MSG:
		msg = new(FullServerFault)
	case constants.COMMIT_CHAIN_MSG:
		msg = new(CommitChainMsg)
	case constants.COMMIT_ENTRY_MSG:
		msg = new(CommitEntryMsg)
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		msg = new(DirectoryBlockSignature)
	case constants.EOM_TIMEOUT_MSG:
		msg = new(EOMTimeout)
	case constants.FACTOID_TRANSACTION_MSG:
		msg = new(FactoidTransaction)
	case constants.HEARTBEAT_MSG:
		msg = new(Heartbeat)
	case constants.INVALID_DIRECTORY_BLOCK_MSG:
		msg = new(InvalidDirectoryBlock)
	case constants.MISSING_MSG:
		msg = new(MissingMsg)
	case constants.MISSING_MSG_RESPONSE:
		msg = new(MissingMsgResponse)
	case constants.MISSING_DATA:
		msg = new(MissingData)
	case constants.DATA_RESPONSE:
		msg = new(DataResponse)
	case constants.REVEAL_ENTRY_MSG:
		msg = new(RevealEntryMsg)
	case constants.REQUEST_BLOCK_MSG:
		msg = new(RequestBlock)
	case constants.SIGNATURE_TIMEOUT_MSG:
		msg = new(SignatureTimeout)
	case constants.DBSTATE_MISSING_MSG:
		msg = new(DBStateMissing)
	case constants.DBSTATE_MSG:
		msg = new(DBStateMsg)
	case constants.ADDSERVER_MSG:
		msg = new(AddServerMsg)
	case constants.CHANGESERVER_KEY_MSG:
		msg = new(ChangeServerKeyMsg)
	case constants.REMOVESERVER_MSG:
		msg = new(RemoveServerMsg)
	case constants.BOUNCE_MSG:
		msg = new(Bounce)
	case constants.BOUNCEREPLY_MSG:
		msg = new(BounceReply)
	default:
		fmt.Sprintf("Transaction Failed to Validate %x", data[0])
		return data, nil, fmt.Errorf("Unknown message type %d %x", messageType, data[0])
	}

	newdata, err = msg.UnmarshalBinaryData(data[:])
	if err != nil {
		fmt.Sprintf("Transaction Failed to Unmarshal %x", data[0])
		return data, nil, err
	}

	return newdata, msg, nil

}

func MessageName(Type byte) string {
	switch Type {
	case constants.EOM_MSG:
		return "EOM"
	case constants.ACK_MSG:
		return "Ack"
	case constants.AUDIT_SERVER_FAULT_MSG:
		return "Audit Server Fault"
	case constants.FED_SERVER_FAULT_MSG:
		return "Fed Server Fault"
	case constants.FULL_SERVER_FAULT_MSG:
		return "Full Server Fault"
	case constants.COMMIT_CHAIN_MSG:
		return "Commit Chain"
	case constants.COMMIT_ENTRY_MSG:
		return "Commit Entry"
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		return "Directory Block Signature"
	case constants.EOM_TIMEOUT_MSG:
		return "EOM Timeout"
	case constants.FACTOID_TRANSACTION_MSG:
		return "Factoid Transaction"
	case constants.HEARTBEAT_MSG:
		return "HeartBeat"
	case constants.INVALID_ACK_MSG:
		return "Invalid Ack"
	case constants.INVALID_DIRECTORY_BLOCK_MSG:
		return "Invalid Directory Block"
	case constants.MISSING_MSG:
		return "Missing Msg"
	case constants.MISSING_MSG_RESPONSE:
		return "Missing Msg Response"
	case constants.MISSING_DATA:
		return "Missing Data"
	case constants.DATA_RESPONSE:
		return "Data Response"
	case constants.REVEAL_ENTRY_MSG:
		return "Reveal Entry"
	case constants.REQUEST_BLOCK_MSG:
		return "Request Block"
	case constants.SIGNATURE_TIMEOUT_MSG:
		return "Signature Timeout"
	case constants.DBSTATE_MISSING_MSG:
		return "DBState Missing"
	case constants.DBSTATE_MSG:
		return "DBState"
	case constants.BOUNCE_MSG:
		return "Bounce Message"
	case constants.BOUNCEREPLY_MSG:
		return "Bounce Reply Message"

	case constants.ADDSERVER_MSG:
		return "ADDSERVER_MSG" // 22
	case constants.CHANGESERVER_KEY_MSG:
		return "CHANGESERVER_KEY_MS" // 23
	case constants.REMOVESERVER_MSG:
		return "REMOVESERVER_MSG" // 24

	default:
		return "Unknown:" + fmt.Sprintf(" %d", Type)
	}
}

type Signable interface {
	Sign(interfaces.Signer) error
	MarshalForSignature() ([]byte, error)
	GetSignature() interfaces.IFullSignature
	IsValid() bool // Signature already checked
	SetValid()     // Mark as validated so we don't have to repeat.
}

func SignSignable(s Signable, key interfaces.Signer) (interfaces.IFullSignature, error) {
	toSign, err := s.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := key.Sign(toSign)
	return sig, nil
}

func VerifyMessage(s Signable) (bool, error) {
	if s.IsValid() {
		return true, nil
	}
	toSign, err := s.MarshalForSignature()
	if err != nil {
		return false, err
	}
	sig := s.GetSignature()
	if sig == nil {
		return false, fmt.Errorf("%s", "Message signature is nil")
	}
	if sig.Verify(toSign) {
		s.SetValid()
		return true, nil
	}
	return false, errors.New("Signarue is invalid")
}
