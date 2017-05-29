// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/blockchainState"
)

type BlockMaker struct {
	PendingEBEntries   []interfaces.IEntry
	ProcessedEBEntries []interfaces.IEntry

	PendingFBEntries   []interfaces.ITransaction
	ProcessedFBEntries []interfaces.ITransaction

	PendingABEntries   []interfaces.IABEntry
	ProcessedABEntries []interfaces.IABEntry

	PendingECBEntries   []interfaces.IECBlockEntry
	ProcessedECBEntries []interfaces.IECBlockEntry

	BState blockchainState.BlockchainState
}
