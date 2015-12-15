// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	. "github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	//"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

// BlockLocator is used to help locate a specific block.  The algorithm for
// building the block locator is to add the hashes in reverse order until
// the genesis block is reached.  In order to keep the list of locator hashes
// to a reasonable number of entries, first the most recent previous 10 block
// hashes are added, then the step is doubled each loop iteration to
// exponentially decrease the number of hashes as a function of the distance
// from the block being located.
//
// For example, assume you have a block chain with a side chain as depicted
// below:
// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
// 	                              \-> 16a -> 17a
//
// The block locator for block 17a would be the hashes of blocks:
// [17a 16a 15 14 13 12 11 10 9 8 6 2 genesis]
type BlockLocator []interfaces.IHash

// DirBlockLocatorFromHash returns a block locator for the passed block hash.
// See BlockLocator for details on the algotirhm used to create a block locator.
//
// In addition to the general algorithm referenced above, there are a couple of
// special cases which are handled:
//
//  - If the genesis hash is passed, there are no previous hashes to add and
//    therefore the block locator will only consist of the genesis hash
//  - If the passed hash is not currently known, the block locator will only
//    consist of the passed hash
func DirBlockLocatorFromHash(hash interfaces.IHash, state interfaces.IState) BlockLocator {
	// The locator contains the requested hash at the very least.
	locator := make(BlockLocator, 0, messages.MaxBlockLocatorsPerMsg)
	locator = append(locator, hash)

	genesisHash, _ := HexToHash(GENESIS_DIR_BLOCK_HASH)
	// Nothing more to do if a locator for the genesis hash was requested.
	if genesisHash.IsSameAs(hash) {
		return locator
	}

	// Attempt to find the height of the block that corresponds to the
	// passed hash, and if it's on a side chain, also find the height at
	// which it forks from the main chain.
	blockHeight := int64(-1)

	// Generate the block locators according to the algorithm described in
	// in the BlockLocator comment and make sure to leave room for the
	// final genesis hash.

	dblock, _ := state.GetDB().FetchDBlockByHash(hash)
	//dblock := dblock0.(directoryBlock.DirectoryBlock)
	if dblock != nil {
		blockHeight = int64(dblock.GetHeader().GetDBHeight())
	}
	increment := int64(1)
	for len(locator) < messages.MaxBlockLocatorsPerMsg-1 {
		// Once there are 10 locators, exponentially increase the
		// distance between each block locator.
		if len(locator) > 10 {
			increment *= 2
		}
		blockHeight -= increment
		if blockHeight < 1 {
			break
		}

		blk, _ := state.GetDB().FetchDBlockByHeight(uint32(blockHeight))
		if blk == nil || blk.GetHash() == nil {
			//blk.DBHash, _ = CreateHash(blk)
			continue
		}

		locator = append(locator, blk.GetHash())
	}

	// Append the appropriate genesis block.
	locator = append(locator, genesisHash)
	return locator
}

// LatestDirBlockLocator returns a block locator for the latest known tip of the
// main (best) chain.
func LatestDirBlockLocator(state interfaces.IState) (BlockLocator, error) {
	latestDirBlockHash := state.GetCurrentDirectoryBlock().GetHash() //, _, _ := db.FetchBlockHeightCache()

	if latestDirBlockHash == nil {
		latestDirBlockHash = NewZeroHash() //zeroHash
	}

	// The best chain is set, so use its hash.
	return DirBlockLocatorFromHash(latestDirBlockHash, state), nil
}
