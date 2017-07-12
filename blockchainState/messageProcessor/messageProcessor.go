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
		break
	case constants.ACK_MSG:
		break
	case constants.FED_SERVER_FAULT_MSG:
		break
	case constants.AUDIT_SERVER_FAULT_MSG:
		break
	case constants.FULL_SERVER_FAULT_MSG:
		break
	case constants.COMMIT_CHAIN_MSG:
		break
	case constants.COMMIT_ENTRY_MSG:
		break
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		break
	case constants.EOM_TIMEOUT_MSG:
		break
	case constants.FACTOID_TRANSACTION_MSG:
		break
	case constants.HEARTBEAT_MSG:
		break
	case constants.INVALID_ACK_MSG:
		break
	case constants.INVALID_DIRECTORY_BLOCK_MSG:
		break
	case constants.REVEAL_ENTRY_MSG:
		break
	case constants.REQUEST_BLOCK_MSG:
		break
	case constants.SIGNATURE_TIMEOUT_MSG:
		break
	case constants.MISSING_MSG:
		break
	case constants.MISSING_DATA:
		break
	case constants.DATA_RESPONSE:
		break
	case constants.MISSING_MSG_RESPONSE:
		break
	case constants.DBSTATE_MSG:
		break
	case constants.DBSTATE_MISSING_MSG:
		break
	case constants.ADDSERVER_MSG:
		break
	case constants.CHANGESERVER_KEY_MSG:
		break
	case constants.REMOVESERVER_MSG:
		break
	case constants.BOUNCE_MSG:
		break
	case constants.BOUNCEREPLY_MSG:
		break
	case constants.MISSING_ENTRY_BLOCKS:
		break
	case constants.ENTRY_BLOCK_RESPONSE:
		break
	}

	if err != nil {
		return err
	}

	return nil
}
