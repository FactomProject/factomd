// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

func UnmarshalMessage(data []byte) (interfaces.IMsg, error) {
	if data == nil {
		return nil, fmt.Errorf("No data provided")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("No data provided")
	}
	messageType := int(data[0])
	switch messageType {
	case constants.EOM_MSG:
		msg := new(EOM)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.AUDIT_SERVER_FAULT_MSG:
		msg := new(AuditServerFault)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.COMMIT_CHAIN_MSG:
		msg := new(CommitChainMsg)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.COMMIT_CHAIN_ACK_MSG:
		msg := new(CommitChainAck)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.COMMIT_ENTRY_MSG:
		msg := new(CommitEntryMsg)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.COMMIT_ENTRY_ACK_MSG:
		msg := new(CommitEntryAck)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		msg := new(DirectoryBlockSignature)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.DUPLICATE_HEIGHT_ACK_MSG:
		msg := new(DuplicateHeightAck)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.EOM_TIMEOUT_MSG:
		msg := new(EOMTimeout)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.FACTOID_TRANSACTION_MSG:
		msg := new(FactoidTransaction)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.HEARTBEAT_MSG:
		msg := new(Heartbeat)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.INVALID_ACK_MSG:
		msg := new(InvalidAck)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.INVALID_DIRECTORY_BLOCK_MSG:
		msg := new(InvalidDirectoryBlock)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.MISSING_ACK_MSG:
		msg := new(MissingAck)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.PROMOTION_DEMOTION_MSG:
		msg := new(PromotionDemotion)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.REVEAL_ENTRY_MSG:
		msg := new(RevealEntry)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.REQUEST_BLOCK_MSG:
		msg := new(RequestBlock)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.REVEAL_ENTRY_ACK_MSG:
		msg := new(RevealEntryAck)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.SIGNATURE_TIMEOUT_MSG:
		msg := new(SignatureTimeout)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case constants.TRANSACTION_ACK_MSG:
		msg := new(TransactionAck)
		err := msg.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return msg, nil
	default:
		return nil, fmt.Errorf("Unknown message type")
	}
	return nil, fmt.Errorf("Unknown message type")
}
