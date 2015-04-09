// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/btcd/blockchain"
	"github.com/FactomProject/btcd/wire"
)

// BlockLocatorFromHash returns a block locator for the passed block hash.
// See BlockLocator for details on the algotirhm used to create a block locator.
//
// In addition to the general algorithm referenced above, there are a couple of
// special cases which are handled:
//
//  - If the genesis hash is passed, there are no previous hashes to add and
//    therefore the block locator will only consist of the genesis hash
//  - If the passed hash is not currently known, the block locator will only
//    consist of the passed hash
func DirBlockLocatorFromHash(hash *wire.ShaHash, dChain *common.DChain) blockchain.BlockLocator {
	// The locator contains the requested hash at the very least.
	locator := make(blockchain.BlockLocator, 0, wire.MaxBlockLocatorsPerMsg)
	locator = append(locator, hash)

	genesisHash, _ := wire.NewShaHash(dchain.Blocks[0].DBHash.Bytes)
	// Nothing more to do if a locator for the genesis hash was requested.
	if hash.IsEqual(genesisHash) {
		return locator
	}

	// Attempt to find the height of the block that corresponds to the
	// passed hash, and if it's on a side chain, also find the height at
	// which it forks from the main chain.
	blockHeight := int64(-1)
	/*	node, exists := b.index[*hash]
		if !exists {
			// Try to look up the height for passed block hash.  Assume an
			// error means it doesn't exist and just return the locator for
			// the block itself.
			block, err := b.db.FetchBlockBySha(hash)
			if err != nil {
				return locator
			}
			blockHeight = block.Height()

		} else {
			blockHeight = node.height

			// Find the height at which this node forks from the main chain
			// if the node is on a side chain.
			if !node.inMainChain {
				for n := node; n.parent != nil; n = n.parent {
					if n.inMainChain {
						forkHeight = n.height
						break
					}
				}
			}
		}
	*/
	// Generate the block locators according to the algorithm described in
	// in the BlockLocator comment and make sure to leave room for the
	// final genesis hash.

	dblock, _ := db.FetchDBlockByHash(hash.ToFactomHash())
	if dblock != nil {
		blockHeight = int64(dblock.Header.BlockID)
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

		locator = append(locator, wire.FactomHashToShaHash(dchain.Blocks[blockHeight].DBHash))
	}

	// Append the appropriate genesis block.
	locator = append(locator, genesisHash)
	return locator
}

// LatestBlockLocator returns a block locator for the latest known tip of the
// main (best) chain.
func LatestDirBlockLocator(dChain *common.DChain) (blockchain.BlockLocator, error) {

	latestDirBlockHash, _ := wire.NewShaHash(dChain.Blocks[dChain.NextBlockID-1].DBHash.Bytes)
	// The best chain is set, so use its hash.
	return DirBlockLocatorFromHash(latestDirBlockHash, dChain), nil
}


// LatestDirBlockSha returns newest shahash of dchain and current height
func LatestDirBlockSha(dChain *common.DChain) (sha *wire.ShaHash, height int64, err error) {

	sha, _ = wire.NewShaHash(dChain.Blocks[dChain.NextBlockID-1].DBHash.Bytes)
		
	height = int64(dChain.Blocks[dChain.NextBlockID-1].Header.BlockID)

	// The best chain is set, so use its hash.
	return sha, height, nil
}