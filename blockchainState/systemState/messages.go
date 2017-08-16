// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (ss *SystemState) ProcessMessage(msg interfaces.IMsg) error {
	var err error
	switch msg.Type() {
	case constants.EOM_MSG:
		err = ss.ProcessEOMMessage(msg)
		break
	case constants.ACK_MSG:
		err = ss.ProcessAckMessage(msg)
		break
	case constants.FED_SERVER_FAULT_MSG:
		err = ss.ProcessFedServerFaultMessage(msg)
		break
	case constants.AUDIT_SERVER_FAULT_MSG:
		err = ss.ProcessAuditServerFaultMessage(msg)
		break
	case constants.FULL_SERVER_FAULT_MSG:
		err = ss.ProcessFullServerFaultMessage(msg)
		break
	case constants.COMMIT_CHAIN_MSG:
		err = ss.ProcessCommitChainMessage(msg)
		break
	case constants.COMMIT_ENTRY_MSG:
		err = ss.ProcessCommitEntryMessage(msg)
		break
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		err = ss.ProcessDirectoryBlockSignatureMessage(msg)
		break
	case constants.EOM_TIMEOUT_MSG:
		err = ss.ProcessEOMTimeoutMessage(msg)
		break
	case constants.FACTOID_TRANSACTION_MSG:
		err = ss.ProcessFactoidTransactionMessage(msg)
		break
	case constants.HEARTBEAT_MSG:
		err = ss.ProcessHeartbeatMessage(msg)
		break
	case constants.INVALID_ACK_MSG:
		err = ss.ProcessInvalidAckMessage(msg)
		break
	case constants.INVALID_DIRECTORY_BLOCK_MSG:
		err = ss.ProcessInvalidDirectoryBlockMessage(msg)
		break
	case constants.REVEAL_ENTRY_MSG:
		err = ss.ProcessRevealEntryMessage(msg)
		break
	case constants.REQUEST_BLOCK_MSG:
		err = ss.ProcessRequestBlockMessage(msg)
		break
	case constants.SIGNATURE_TIMEOUT_MSG:
		err = ss.ProcessSignatureTimeoutMessage(msg)
		break
	case constants.MISSING_MSG:
		err = ss.ProcessMissingMessage(msg)
		break
	case constants.MISSING_DATA:
		err = ss.ProcessMissingDataMessage(msg)
		break
	case constants.DATA_RESPONSE:
		err = ss.ProcessDataResponseMessage(msg)
		break
	case constants.MISSING_MSG_RESPONSE:
		err = ss.ProcessMissingMessage(msg)
		break
	case constants.DBSTATE_MSG:
		err = ss.ProcessDBStateMessage(msg)
		break
	case constants.DBSTATE_MISSING_MSG:
		err = ss.ProcessDBStateMissingMessage(msg)
		break
	case constants.ADDSERVER_MSG:
		err = ss.ProcessAddServerMessage(msg)
		break
	case constants.CHANGESERVER_KEY_MSG:
		err = ss.ProcessChangeServerKeyMessage(msg)
		break
	case constants.REMOVESERVER_MSG:
		err = ss.ProcessRemoveServerMessage(msg)
		break
	case constants.BOUNCE_MSG:
		err = ss.ProcessBounceMessage(msg)
		break
	case constants.BOUNCEREPLY_MSG:
		err = ss.ProcessBounceReplyMessage(msg)
		break
	case constants.MISSING_ENTRY_BLOCKS:
		err = ss.ProcessMissingEntryBlocksMessage(msg)
		break
	case constants.ENTRY_BLOCK_RESPONSE:
		err = ss.ProcessEntryBlockResponseMessage(msg)
		break
	}

	if err != nil {
		return err
	}

	return nil
}

func (ss *SystemState) ProcessAckMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.ACK_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}

	if ss.MessageHoldingQueue.IsAcked(msg.GetHash()) == true {
		//Nothing to do, the message is already acked
		//TODO: double-check
		return nil
	}

	ss.MessageHoldingQueue.AddAck(msg)
	msg2 := ss.MessageHoldingQueue.GetMessage(msg.GetHash())
	if msg2 != nil {
		//If we have acked a message we know about, time to process it
		return ss.ProcessMessage(msg2)
	}

	return nil
}

func (ss *SystemState) ProcessAddServerMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.ADDSERVER_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessAuditServerFaultMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.AUDIT_SERVER_FAULT_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessBounceMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.BOUNCE_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessBounceReplyMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.BOUNCEREPLY_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessChangeServerKeyMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.CHANGESERVER_KEY_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessCommitChainMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.COMMIT_CHAIN_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}

	ss.MessageHoldingQueue.AddMessage(msg)

	if ss.MessageHoldingQueue.IsAcked(msg.GetHash()) {
		//TODO: process properly
	}

	return nil
}

func (ss *SystemState) ProcessCommitEntryMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.COMMIT_ENTRY_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}

	ss.MessageHoldingQueue.AddMessage(msg)

	if ss.MessageHoldingQueue.IsAcked(msg.GetHash()) {
		//TODO: process properly
	}

	return nil
}

func (ss *SystemState) ProcessDataResponseMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.DATA_RESPONSE {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessDBStateMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.DBSTATE_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return ss.BStateHandler.HandleDBStateMsg(msg)
}

func (ss *SystemState) ProcessDBStateMissingMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.DBSTATE_MISSING_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessDirectoryBlockSignatureMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessEntryBlockResponseMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.ENTRY_BLOCK_RESPONSE {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}

	ss.MessageHoldingQueue.AddMessage(msg)

	if ss.MessageHoldingQueue.IsAcked(msg.GetHash()) {
		//TODO: process properly
	}

	return nil
}

func (ss *SystemState) ProcessEOMMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.EOM_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}

	ss.MessageHoldingQueue.AddMessage(msg)

	if ss.MessageHoldingQueue.IsAcked(msg.GetHash()) {
		//TODO: process properly
	}

	return nil
}

func (ss *SystemState) ProcessEOMTimeoutMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.EOM_TIMEOUT_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessFactoidTransactionMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.FACTOID_TRANSACTION_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}

	ss.MessageHoldingQueue.AddMessage(msg)

	if ss.MessageHoldingQueue.IsAcked(msg.GetHash()) {
		//TODO: process properly
	}

	return nil
}

func (ss *SystemState) ProcessFedServerFaultMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.FED_SERVER_FAULT_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessFullServerFaultMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.FULL_SERVER_FAULT_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessHeartbeatMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.HEARTBEAT_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessInvalidAckMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.INVALID_ACK_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessMissingMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.MISSING_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessMissingDataMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.MISSING_DATA {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessMissingEntryBlocksMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.MISSING_ENTRY_BLOCKS {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessMissingMessageResponseMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.MISSING_MSG_RESPONSE {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessRemoveServerMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.REMOVESERVER_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessRequestBlockMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.REQUEST_BLOCK_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}

func (ss *SystemState) ProcessRevealEntryMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.REVEAL_ENTRY_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}

	ss.MessageHoldingQueue.AddMessage(msg)

	if ss.MessageHoldingQueue.IsAcked(msg.GetHash()) {
		//TODO: process properly
	}

	return nil
}

func (ss *SystemState) ProcessSignatureTimeoutMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.SIGNATURE_TIMEOUT_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}
