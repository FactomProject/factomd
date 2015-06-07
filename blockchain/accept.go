// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"

	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcutil"

	"github.com/davecgh/go-spew/spew"
)

// maybeAcceptBlock potentially accepts a block into the memory block chain.
// It performs several validation checks which depend on its position within
// the block chain before adding it.  The block is expected to have already gone
// through ProcessBlock before calling this function with it.
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: The somewhat expensive BIP0034 validation is not performed.
//  - BFDryRun: The memory chain index will not be pruned and no accept
//    notification will be sent since the block is not being accepted.
func (b *BlockChain) maybeAcceptBlock(block *btcutil.Block, flags BehaviorFlags) error {
	util.Trace()
	fmt.Println("maybeAcceptBlock=", spew.Sdump(block))

	fastAdd := flags&BFFastAdd == BFFastAdd
	dryRun := flags&BFDryRun == BFDryRun

	// Get a block node for the block previous to this one.  Will be nil
	// if this is the genesis block.
	prevNode, err := b.getPrevNodeFromBlock(block)
	if err != nil {
		log.Errorf("getPrevNodeFromBlock: %v", err)
		return err
	}

	util.Trace()
	// The height of this block is one more than the referenced previous
	// block.
	blockHeight := int64(0)
	if prevNode != nil {
		blockHeight = prevNode.height + 1
	}
	block.SetHeight(blockHeight)

	blockHeader := &block.MsgBlock().Header
	if !fastAdd {
		/*
			// Ensure the difficulty specified in the block header matches
			// the calculated difficulty based on the previous block and
			// difficulty retarget rules.
			expectedDifficulty, err := b.calcNextRequiredDifficulty(prevNode,
				block.MsgBlock().Header.Timestamp)
			if err != nil {
				return err
			}
			blockDifficulty := blockHeader.Bits
			if blockDifficulty != expectedDifficulty {
				str := "block difficulty of %d is not the expected value of %d"
				str = fmt.Sprintf(str, blockDifficulty, expectedDifficulty)
				return ruleError(ErrUnexpectedDifficulty, str)
			}
		*/

		// FIXME TODO : determine if this is needed for Factoids, probably a good idea
		/*
			// Ensure the timestamp for the block header is after the
			// median time of the last several blocks (medianTimeBlocks).
			medianTime, err := b.calcPastMedianTime(prevNode)
			if err != nil {
				log.Errorf("calcPastMedianTime: %v", err)
				return err
			}
			util.Trace()

			if !blockHeader.Timestamp.After(medianTime) {
				str := "block timestamp of %v is not after expected %v"
				str = fmt.Sprintf(str, blockHeader.Timestamp,
					medianTime)
				return ruleError(ErrTimeTooOld, str)
			}
			util.Trace()
		*/

		// Ensure all transactions in the block are finalized.
		for _, tx := range block.Transactions() {
			if !IsFinalizedTransaction(tx, blockHeight) {
				str := fmt.Sprintf("block contains "+
					"unfinalized transaction %v", tx.Sha())
				return ruleError(ErrUnfinalizedTx, str)
			}
		}

		util.Trace()
	}

	// Ensure chain matches up to predetermined checkpoints.
	// It's safe to ignore the error on Sha since it's already cached.
	blockHash, _ := block.Sha()
	/*
			if !b.verifyCheckpoint(blockHeight, blockHash) {
				str := fmt.Sprintf("block at height %d does not match "+
					"checkpoint hash", blockHeight)
				return ruleError(ErrBadCheckpoint, str)
			}
			util.Trace()

		// Find the previous checkpoint and prevent blocks which fork the main
		// chain before it.  This prevents storage of new, otherwise valid,
		// blocks which build off of old blocks that are likely at a much easier
		// difficulty and therefore could be used to waste cache and disk space.
		checkpointBlock, err := b.findPreviousCheckpoint()
		if err != nil {
			return err
		}
		util.Trace()
		if checkpointBlock != nil && blockHeight < checkpointBlock.Height() {
			str := fmt.Sprintf("block at height %d forks the main chain "+
				"before the previous checkpoint at height %d",
				blockHeight, checkpointBlock.Height())
			return ruleError(ErrForkTooOld, str)
		}
		util.Trace()
	*/

	/*********************************************
	    Once there was code here for rejecting blocks.
	    FACTOM doesn't do this sort of thing.  Blocks
	    simply are not rejected.

	    Rejecting messages or transactins based on
	    versions will be done above the coin.  So
	    we don't need that logic here.
	**********************************************/
	// Prune block nodes which are no longer needed before creating
	// a new node.
	if !dryRun {
		err = b.pruneBlockNodes()
		if err != nil {
			return err
		}
	}

	util.Trace()

	// Create a new block node for the block and add it to the in-memory
	// block chain (could be either a side chain or the main chain).
	newNode := newBlockNode(blockHeader, blockHash, blockHeight)
	if prevNode != nil {
		newNode.parent = prevNode
		newNode.height = blockHeight
		//		newNode.workSum.Add(prevNode.workSum, newNode.workSum)
	}
	util.Trace()

	// Connect the passed block to the chain while respecting proper chain
	// selection according to the chain with the most proof of work.  This
	// also handles validation of the transaction scripts.
	err = b.connectBestChain(newNode, block, flags)
	if err != nil {
		return err
	}
	util.Trace()

	// Notify the caller that the new block was accepted into the block
	// chain.  The caller would typically want to react by relaying the
	// inventory to other peers.
	if !dryRun {
		b.sendNotification(NTBlockAccepted, block)
	}

	return nil
}
