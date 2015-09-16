// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/btcd/blockchain"
	"github.com/FactomProject/btcd/wire"
)

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
func DirBlockLocatorFromHash(hash *wire.ShaHash) blockchain.BlockLocator {
	// The locator contains the requested hash at the very least.
	locator := make(blockchain.BlockLocator, 0, wire.MaxBlockLocatorsPerMsg)
	locator = append(locator, hash)

	h, _ := common.HexToHash(common.GENESIS_DIR_BLOCK_HASH)
	genesisHash := wire.FactomHashToShaHash(h)
	// Nothing more to do if a locator for the genesis hash was requested.
	if genesisHash.IsEqual(hash) {
		return locator
	}

	// Attempt to find the height of the block that corresponds to the
	// passed hash, and if it's on a side chain, also find the height at
	// which it forks from the main chain.
	blockHeight := int64(-1)

	// Generate the block locators according to the algorithm described in
	// in the BlockLocator comment and make sure to leave room for the
	// final genesis hash.

	dblock, _ := db.FetchDBlockByHash(hash.ToFactomHash())
	if dblock != nil {
		blockHeight = int64(dblock.Header.DBHeight)
	}
	increment := int64(1)
	for len(locator) < wire.MaxBlockLocatorsPerMsg-1 {
		// Once there are 10 locators, exponentially increase the
		// distance between each block locator.
		if len(locator) > 10 {
			increment *= 2
		}
		blockHeight -= increment
		if blockHeight < 1 {
			break
		}

		blk, _ := db.FetchDBlockByHeight(uint32(blockHeight))
		if blk == nil {
			continue
		} else if blk.DBHash == nil {
			blk.DBHash, _ = common.CreateHash(blk)
		}

		locator = append(locator, wire.FactomHashToShaHash(blk.DBHash))
	}

	// Append the appropriate genesis block.
	locator = append(locator, genesisHash)
	return locator
}

// LatestDirBlockLocator returns a block locator for the latest known tip of the
// main (best) chain.
func LatestDirBlockLocator() (blockchain.BlockLocator, error) {
	latestDirBlockHash, _, _ := db.FetchBlockHeightCache()

	if latestDirBlockHash == nil {
		latestDirBlockHash = &zeroHash
	}

	// The best chain is set, so use its hash.
	return DirBlockLocatorFromHash(latestDirBlockHash), nil
}
