// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package process

import (
	"github.com/FactomProject/factomd/util"
	"github.com/davecgh/go-spew/spew"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

var _ = util.Trace
var _ = spew.Sdump

func GetEntryCreditBalance(pubKey *[32]byte) (int32, error) {

	return eCreditMap[string(pubKey[:])], nil
}

//--------------------------------------

func getPrePaidChainKey(entryHash *Hash, chainIDHash *Hash) string {
	return chainIDHash.String() + entryHash.String()
}

func copyCreditMap(
	originalMap map[string]int32,
	newMap map[string]int32) {
	newMap = make(map[string]int32)

	// copy every element from the original map
	for k, v := range originalMap {
		newMap[k] = v
	}

}

func printCreditMap() {
	procLog.Debug("eCreditMap:")
	for key := range eCreditMap {
		procLog.Debugf("Entry credit Key: %x Value %d", key, eCreditMap[key])
	}
}

// HaveBlockInDB returns whether or not the chain instance has the block represented
// by the passed hash.  This includes checking the various places a block can
// be like part of the main chain, on a side chain, or in the orphan pool.
//
// This function is NOT safe for concurrent access.
func HaveBlockInDB(hash interfaces.IHash) (bool, error) {
	//util.Trace(spew.Sdump(hash))

	if hash == nil || dchain.Blocks == nil || len(dchain.Blocks) == 0 {
		return false, nil
	}

	// double check the block ids
	for i := 0; i < len(dchain.Blocks); i = i + 1 {
		if dchain.Blocks[i] == nil {
			continue
		}
		if dchain.Blocks[i].DBHash == nil {
			dchain.Blocks[i].DBHash, _ = CreateHash(dchain.Blocks[i])
		}
		if dchain.Blocks[i].DBHash.IsSameAs(hash) {
			return true, nil
		}
	}

	return false, nil
}
