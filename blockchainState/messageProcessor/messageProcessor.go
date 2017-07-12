// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messageProcessor

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

type MessageProcessor struct {
}

func (mp *MessageProcessor) ProcessMessage(msg interfaces.IMsg) error {
	var err error
	switch msg.Type() {
	case constants.EOM_MSG:
		err = mp.ProcessEOMMessage(msg)
		break
	case constants.ACK_MSG:
		err = mp.ProcessAckMessage(msg)
		break
	case constants.FED_SERVER_FAULT_MSG:
		err = mp.ProcessFedServerFaultMessage(msg)
		break
	case constants.AUDIT_SERVER_FAULT_MSG:
		err = mp.ProcessAuditServerFaultMessage(msg)
		break
	case constants.FULL_SERVER_FAULT_MSG:
		err = mp.ProcessFullServerFaultMessage(msg)
		break
	case constants.COMMIT_CHAIN_MSG:
		err = mp.ProcessCommitChainMessage(msg)
		break
	case constants.COMMIT_ENTRY_MSG:
		err = mp.ProcessCommitEntryMessage(msg)
		break
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		err = mp.ProcessDirectoryBlockSignatureMessage(msg)
		break
	case constants.EOM_TIMEOUT_MSG:
		err = mp.ProcessEOMTimeoutMessage(msg)
		break
	case constants.FACTOID_TRANSACTION_MSG:
		err = mp.ProcessFactoidTransactionMessage(msg)
		break
	case constants.HEARTBEAT_MSG:
		err = mp.ProcessHeartbeatMessage(msg)
		break
	case constants.INVALID_ACK_MSG:
		err = mp.ProcessInvalidAckMessage(msg)
		break
	case constants.INVALID_DIRECTORY_BLOCK_MSG:
		err = mp.ProcessInvalidDirectoryBlockMessage(msg)
		break
	case constants.REVEAL_ENTRY_MSG:
		err = mp.ProcessRevealEntryMessage(msg)
		break
	case constants.REQUEST_BLOCK_MSG:
		err = mp.ProcessRequestBlockMessage(msg)
		break
	case constants.SIGNATURE_TIMEOUT_MSG:
		err = mp.ProcessSignatureTimeoutMessage(msg)
		break
	case constants.MISSING_MSG:
		err = mp.ProcessMissingMessage(msg)
		break
	case constants.MISSING_DATA:
		err = mp.ProcessMissingDataMessage(msg)
		break
	case constants.DATA_RESPONSE:
		err = mp.ProcessDataResponseMessage(msg)
		break
	case constants.MISSING_MSG_RESPONSE:
		err = mp.ProcessMissingMessage(msg)
		break
	case constants.DBSTATE_MSG:
		err = mp.ProcessDBStateMessage(msg)
		break
	case constants.DBSTATE_MISSING_MSG:
		err = mp.ProcessDBStateMissingMessage(msg)
		break
	case constants.ADDSERVER_MSG:
		err = mp.ProcessAddServerMessage(msg)
		break
	case constants.CHANGESERVER_KEY_MSG:
		err = mp.ProcessChangeServerKeyMessage(msg)
		break
	case constants.REMOVESERVER_MSG:
		err = mp.ProcessRemoveServerMessage(msg)
		break
	case constants.BOUNCE_MSG:
		err = mp.ProcessBounceMessage(msg)
		break
	case constants.BOUNCEREPLY_MSG:
		err = mp.ProcessBounceReplyMessage(msg)
		break
	case constants.MISSING_ENTRY_BLOCKS:
		err = mp.ProcessMissingEntryBlocksMessage(msg)
		break
	case constants.ENTRY_BLOCK_RESPONSE:
		err = mp.ProcessEntryBlockResponseMessage(msg)
		break
	}

	if err != nil {
		return err
	}

	return nil
}
