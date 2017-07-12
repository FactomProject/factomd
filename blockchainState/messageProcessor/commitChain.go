// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messageProcessor

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (mp *MessageProcessor) ProcessCommitChainMessage(msg interfaces.IMsg) error {
	if msg.Type() != constants.COMMIT_CHAIN_MSG {
		return fmt.Errorf("Invalid message type forwarded for processing")
	}
	return nil
}
