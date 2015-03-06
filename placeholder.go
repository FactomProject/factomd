//
// placeholder stuff, while merging
//

package main

import (
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/btcutil"
	"math"
)

// IsCoinBase determines whether or not a transaction is a coinbase.  A coinbase
// is a special transaction created by miners that has no inputs.  This is
// represented in the block chain by a transaction with a single input that has
// a previous output transaction index set to the maximum value along with a
// zero hash.
func IsCoinBase(tx *btcutil.Tx) bool {
	msgTx := tx.MsgTx()

	// A coin base must only have one transaction input.
	if len(msgTx.TxIn) != 1 {
		return false
	}

	zero_Hash := &wire.ShaHash{}

	// The previous output of a coin base must have a max value index and
	// a zero hash.
	prevOut := msgTx.TxIn[0].PreviousOutPoint
	if prevOut.Index != math.MaxUint32 || !prevOut.Hash.IsEqual(zero_Hash) {
		return false
	}

	return true
}
