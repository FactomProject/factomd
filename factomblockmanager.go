// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"sync/atomic"

	//"github.com/FactomProject/btcd/blockchain"
	"github.com/FactomProject/btcd/wire"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/util"
)

// dirBlockMsg packages a directory block message and the peer it came from together
// so the block handler has access to that information.
type dirBlockMsg struct {
	block *common.DBlock
	peer  *peer
}

// dirInvMsg packages a dir block inv message and the peer it came from together
// so the block handler has access to that information.
type dirInvMsg struct {
	inv  *wire.MsgDirInv
	peer *peer
}

// handleDirBlockMsg handles dir block messages from all peers.
func (b *blockManager) handleDirBlockMsg(bmsg *dirBlockMsg) {
	util.Trace()
	// If we didn't ask for this block then the peer is misbehaving.
	binary, _ := bmsg.block.MarshalBinary()
	commonHash := common.Sha(binary)
	blockSha, _ := wire.NewShaHash(commonHash.Bytes)

	if _, ok := bmsg.peer.requestedBlocks[*blockSha]; !ok {
		// The regression test intentionally sends some blocks twice
		// to test duplicate block insertion fails.  Don't disconnect
		// the peer or ignore the block when we're in regression test
		// mode in this case so the chain code is actually fed the
		// duplicate blocks.
		if !cfg.RegressionTest {
			bmgrLog.Warnf("Got unrequested block %v from %s -- "+
				"disconnecting", blockSha, bmsg.peer.addr)
			bmsg.peer.Disconnect()
			return
		}
	}

	/*
		// When in headers-first mode, if the block matches the hash of the
		// first header in the list of headers that are being fetched, it's
		// eligible for less validation since the headers have already been
		// verified to link together and are valid up to the next checkpoint.
		// Also, remove the list entry for all blocks except the checkpoint
		// since it is needed to verify the next round of headers links
		// properly.
		isCheckpointBlock := false
		behaviorFlags := blockchain.BFNone
		if b.headersFirstMode {
			firstNodeEl := b.headerList.Front()
			if firstNodeEl != nil {
				firstNode := firstNodeEl.Value.(*headerNode)
				if blockSha.IsEqual(firstNode.sha) {
					behaviorFlags |= blockchain.BFFastAdd
					if firstNode.sha.IsEqual(b.nextCheckpoint.Hash) {
						isCheckpointBlock = true
					} else {
						b.headerList.Remove(firstNodeEl)
					}
				}
			}
		}
	*/

	// Remove block from request maps. Either chain will know about it and
	// so we shouldn't have any more instances of trying to fetch it, or we
	// will fail the insert and thus we'll retry next time we get an inv.
	delete(bmsg.peer.requestedBlocks, *blockSha)
	delete(b.requestedBlocks, *blockSha)

	util.Trace("just before BC_ProcessBlock")

	// Process the block to include validation, best chain selection, orphan
	// handling, etc.
	/*	isOrphan, err := b.blockChain.BC_ProcessBlock(bmsg.block,
		//		b.server.timeSource, behaviorFlags)
		//		b.server.timeSource, 0)
		b.server.timeSource, blockchain.BFFactomFlag1)
	*/
	isOrphan := false // to be improved

	util.Trace("BC_ProcessBlock error checking")
	/*	if err != nil {
			// When the error is a rule error, it means the block was simply
			// rejected as opposed to something actually going wrong, so log
			// it as such.  Otherwise, something really did go wrong, so log
			// it as an actual error.
			if _, ok := err.(blockchain.RuleError); ok {
				bmgrLog.Infof("Rejected block %v from %s: %v", blockSha,
					bmsg.peer, err)
			} else {
				bmgrLog.Errorf("Failed to process block %v: %v",
					blockSha, err)
			}

			// Convert the error into an appropriate reject message and
			// send it.
			code, reason := errToRejectErr(err)
			bmsg.peer.PushRejectMsg(wire.CmdDirBlock, code, reason,
				blockSha, false)
			return
		}
	*/
	util.Trace("just before block orphan checking")

	// Request the parents for the orphan block from the peer that sent it.
	if isOrphan {
		orphanRoot := b.blockChain.GetOrphanRoot(blockSha)
		locator, err := b.blockChain.LatestBlockLocator()
		if err != nil {
			bmgrLog.Warnf("Failed to get block locator for the "+
				"latest block: %v", err)
		} else {
			bmsg.peer.PushGetDirBlocksMsg(locator, orphanRoot)
		}
	} else {
		// When the block is not an orphan, log information about it and
		// update the chain state.

		util.Trace()

		//b.progressLogger.LogBlockHeight(bmsg.block)

		// Query the db for the latest best block since the block
		// that was processed could be on a side chain or have caused
		// a reorg.
		//		newestSha, newestHeight, _ := btcd.db.NewestSha()
		//		b.updateChainState(newestSha, newestHeight)

		// Allow any clients performing long polling via the
		// getblocktemplate RPC to be notified when the new block causes
		// their old block template to become stale.

		//rpcServer := b.server.rpcServer
		//if rpcServer != nil {
		//rpcServer.gbtWorkState.NotifyBlockConnected(blockSha)
		//}

	}

	// Sync the db to disk.
	db.Sync() // factom db

	/*
		// Nothing more to do if we aren't in headers-first mode.
		if !b.headersFirstMode {
			return
		}

		// This is headers-first mode, so if the block is not a checkpoint
		// request more blocks using the header list when the request queue is
		// getting short.
		if !isCheckpointBlock {
			if b.startHeader != nil &&
				len(bmsg.peer.requestedBlocks) < minInFlightBlocks {
				b.fetchHeaderBlocks()
			}
			return
		}

		// This is headers-first mode and the block is a checkpoint.  When
		// there is a next checkpoint, get the next round of headers by asking
		// for headers starting from the block after this one up to the next
		// checkpoint.
		prevHeight := b.nextCheckpoint.Height
		prevHash := b.nextCheckpoint.Hash
		b.nextCheckpoint = b.findNextHeaderCheckpoint(prevHeight)
		if b.nextCheckpoint != nil {
			locator := blockchain.BlockLocator([]*wire.ShaHash{prevHash})
			err := bmsg.peer.PushGetHeadersMsg(locator, b.nextCheckpoint.Hash)
			if err != nil {
				bmgrLog.Warnf("Failed to send getheaders message to "+
					"peer %s: %v", bmsg.peer.addr, err)
				return
			}
			bmgrLog.Infof("Downloading headers for blocks %d to %d from "+
				"peer %s", prevHeight+1, b.nextCheckpoint.Height,
				b.syncPeer.addr)
			return
		}

		// This is headers-first mode, the block is a checkpoint, and there are
		// no more checkpoints, so switch to normal mode by requesting blocks
		// from the block after this one up to the end of the chain (zero hash).
		b.headersFirstMode = false
		b.headerList.Init()
		bmgrLog.Infof("Reached the final checkpoint -- switching to normal mode")
		locator := blockchain.BlockLocator([]*wire.ShaHash{blockSha})
		err = bmsg.peer.PushGetBlocksMsg(locator, &zeroHash)
		if err != nil {
			bmgrLog.Warnf("Failed to send getblocks message to peer %s: %v",
				bmsg.peer.addr, err)
			return
		}
	*/
}

// handleDirInvMsg handles dir inv messages from all peers.
// We examine the inventory advertised by the remote peer and act accordingly.
func (b *blockManager) handleDirInvMsg(imsg *dirInvMsg) {
	util.Trace()
	/*
		// Ignore invs from peers that aren't the sync if we are not current.
		// Helps prevent fetching a mass of orphans.
		if imsg.peer != b.syncPeer && !b.current() {
			return
		}
	*/

	// Attempt to find the final block in the inventory list.  There may
	// not be one.
	lastBlock := -1
	invVects := imsg.inv.InvList
	for i := len(invVects) - 1; i >= 0; i-- {
		if invVects[i].Type == wire.InvTypeFactomDirBlock {
			lastBlock = i
			break
		}
	}

	// Request the advertised inventory if we don't already have it.  Also,
	// request parent blocks of orphans if we receive one we already have.
	// Finally, attempt to detect potential stalls due to long side chains
	// we already have and request more blocks to prevent them.
	chain := b.blockChain
	for i, iv := range invVects {
		// Ignore unsupported inventory types.
		if iv.Type != wire.InvTypeFactomDirBlock { //} && iv.Type != wire.InvTypeTx {
			continue
		}

		// Add the inventory to the cache of known inventory
		// for the peer.
		imsg.peer.AddKnownInventory(iv)

		// Ignore inventory when we're in headers-first mode.
		//if b.headersFirstMode {
		//continue
		//}

		// Request the inventory if we don't already have it.
		haveInv, err := b.haveInventory(iv)
		if err != nil {
			bmgrLog.Warnf("Unexpected failure when checking for "+
				"existing inventory during inv message "+
				"processing: %v", err)
			continue
		}
		if !haveInv {
			// Add it to the request queue.
			imsg.peer.requestQueue = append(imsg.peer.requestQueue, iv)
			continue
		}

		if iv.Type == wire.InvTypeFactomDirBlock {
			// The block is an orphan block that we already have.
			// When the existing orphan was processed, it requested
			// the missing parent blocks.  When this scenario
			// happens, it means there were more blocks missing
			// than are allowed into a single inventory message.  As
			// a result, once this peer requested the final
			// advertised block, the remote peer noticed and is now
			// resending the orphan block as an available block
			// to signal there are more missing blocks that need to
			// be requested.
			if chain.IsKnownOrphan(&iv.Hash) {
				// Request blocks starting at the latest known
				// up to the root of the orphan that just came
				// in.
				orphanRoot := chain.GetOrphanRoot(&iv.Hash)
				locator, err := chain.LatestBlockLocator()
				if err != nil {
					bmgrLog.Errorf("PEER: Failed to get block "+
						"locator for the latest block: "+
						"%v", err)
					continue
				}
				imsg.peer.PushGetDirBlocksMsg(locator, orphanRoot)
				continue
			}

			// We already have the final block advertised by this
			// inventory message, so force a request for more.  This
			// should only happen if we're on a really long side
			// chain.
			if i == lastBlock {
				// Request blocks after this one up to the
				// final one the remote peer knows about (zero
				// stop hash).
				locator := chain.BlockLocatorFromHash(&iv.Hash)
				imsg.peer.PushGetDirBlocksMsg(locator, &zeroHash)
			}
		}
	}

	// Request as much as possible at once.  Anything that won't fit into
	// the request will be requested on the next inv message.
	numRequested := 0
	gdmsg := wire.NewMsgGetDirData()
	requestQueue := imsg.peer.requestQueue
	for len(requestQueue) != 0 {
		iv := requestQueue[0]
		requestQueue[0] = nil
		requestQueue = requestQueue[1:]

		switch iv.Type {
		case wire.InvTypeFactomDirBlock:
			// Request the block if there is not already a pending
			// request.
			if _, exists := b.requestedBlocks[iv.Hash]; !exists {
				b.requestedBlocks[iv.Hash] = struct{}{}
				imsg.peer.requestedBlocks[iv.Hash] = struct{}{}
				gdmsg.AddInvVect(iv)
				numRequested++
			}

		case wire.InvTypeTx:
			// Request the transaction if there is not already a
			// pending request.
			if _, exists := b.requestedTxns[iv.Hash]; !exists {
				b.requestedTxns[iv.Hash] = struct{}{}
				imsg.peer.requestedTxns[iv.Hash] = struct{}{}
				gdmsg.AddInvVect(iv)
				numRequested++
			}
		}

		if numRequested >= wire.MaxInvPerMsg {
			break
		}
	}
	imsg.peer.requestQueue = requestQueue
	if len(gdmsg.InvList) > 0 {
		imsg.peer.QueueMessage(gdmsg, nil)
	}
}

// QueueDirBlock adds the passed GetDirBlocks message and peer to the block handling queue.
func (b *blockManager) QueueDirBlock(msg *wire.MsgDirBlock, p *peer) {
	// Don't accept more blocks if we're shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		p.blockProcessed <- struct{}{}
		return
	}

	b.msgChan <- &dirBlockMsg{block: msg.DBlk, peer: p}
}

// QueueDirInv adds the passed inv message and peer to the block handling queue.
func (b *blockManager) QueueDirInv(inv *wire.MsgDirInv, p *peer) {
	//	util.Trace()
	// No channel handling here because peers do not need to block on inv
	// messages.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}

	b.msgChan <- &dirInvMsg{inv: inv, peer: p}
}

/*
// ProcessBlock makes use of ProcessBlock on an internal instance of a block
// chain.  It is funneled through the block manager since btcchain is not safe
// for concurrent access.
func (b *blockManager) bm_ProcessBlock(block *btcutil.Block, flags blockchain.BehaviorFlags) (bool, error) {
	util.Trace()
	reply := make(chan processBlockResponse, 1)
	b.msgChan <- processBlockMsg{block: block, flags: flags, reply: reply}
	response := <-reply
	return response.isOrphan, response.err
}
*/

// IsCurrent returns whether or not the block manager believes it is synced with
// the connected peers.
//func (b *blockManager) IsCurrent() bool {
/*
	reply := make(chan bool)
	b.msgChan <- isCurrentMsg{reply: reply}
	return <-reply
*/
//return true
//}

/*
// startSync will choose the best peer among the available candidate peers to
// download/sync the blockchain from.  When syncing is already running, it
// simply returns.  It also examines the candidates for any which are no longer
// candidates and removes them as needed.
func (b *blockManager) startSync(peers *list.List) {
	// Return now if we're already syncing.
	if b.syncPeer != nil {
		return
	}

	// Find the height of the current known best block.
	_, height, err := btcd.db.NewestSha()
	if err != nil {
		bmgrLog.Errorf("%v", err)
		return
	}

	var bestPeer *peer
	var enext *list.Element
	for e := peers.Front(); e != nil; e = enext {
		enext = e.Next()
		p := e.Value.(*peer)

		// Remove sync candidate peers that are no longer candidates due
		// to passing their latest known block.  NOTE: The < is
		// intentional as opposed to <=.  While techcnically the peer
		// doesn't have a later block when it's equal, it will likely
		// have one soon so it is a reasonable choice.  It also allows
		// the case where both are at 0 such as during regression test.
		if p.lastBlock < int32(height) {
			peers.Remove(e)
			continue
		}

		// TODO(davec): Use a better algorithm to choose the best peer.
		// For now, just pick the first available candidate.
		bestPeer = p
	}

	// Start syncing from the best peer if one was selected.
	if bestPeer != nil {
		locator, err := b.blockChain.LatestBlockLocator()
		if err != nil {
			bmgrLog.Errorf("Failed to get block locator for the "+
				"latest block: %v", err)
			return
		}

		bmgrLog.Infof("Syncing to block height %d from peer %v",
			bestPeer.lastBlock, bestPeer.addr)

		// When the current height is less than a known checkpoint we
		// can use block headers to learn about which blocks comprise
		// the chain up to the checkpoint and perform less validation
		// for them.  This is possible since each header contains the
		// hash of the previous header and a merkle root.  Therefore if
		// we validate all of the received headers link together
		// properly and the checkpoint hashes match, we can be sure the
		// hashes for the blocks in between are accurate.  Further, once
		// the full blocks are downloaded, the merkle root is computed
		// and compared against the value in the header which proves the
		// full block hasn't been tampered with.
		//
		// Once we have passed the final checkpoint, or checkpoints are
		// disabled, use standard inv messages learn about the blocks
		// and fully validate them.  Finally, regression test mode does
		// not support the headers-first approach so do normal block
		// downloads when in regression test mode.
		if b.nextCheckpoint != nil && height < b.nextCheckpoint.Height &&
			!cfg.RegressionTest && !cfg.DisableCheckpoints {

			bestPeer.PushGetHeadersMsg(locator, b.nextCheckpoint.Hash)
			b.headersFirstMode = true
			bmgrLog.Infof("Downloading headers for blocks %d to "+
				"%d from peer %s", height+1,
				b.nextCheckpoint.Height, bestPeer.addr)
		} else {
			bestPeer.PushGetBlocksMsg(locator, &zeroHash)
		}
		b.syncPeer = bestPeer
	} else {
		bmgrLog.Warnf("No sync peer candidates available")
	}
}

// isSyncCandidate returns whether or not the peer is a candidate to consider
// syncing from.
func (b *blockManager) isSyncCandidate(p *peer) bool {
	// Typically a peer is not a candidate for sync if it's not a full node,
	// however regression test is special in that the regression tool is
	// not a full node and still needs to be considered a sync candidate.
	if cfg.RegressionTest {
		// The peer is not a candidate if it's not coming from localhost
		// or the hostname can't be determined for some reason.
		host, _, err := net.SplitHostPort(p.addr)
		if err != nil {
			return false
		}

		if host != "127.0.0.1" && host != "localhost" {
			return false
		}
	} else {
		// The peer is not a candidate for sync if it's not a full node.
		if p.services&wire.SFNodeNetwork != wire.SFNodeNetwork {
			return false
		}
	}

	// Candidate if all checks passed.
	return true
}
*/

// current returns true if we believe we are synced with our peers, false if we
// still have blocks to check
//func (b *blockManager) current() bool {
/*
	if !b.blockChain.IsCurrent(b.server.timeSource) {
		return false
	}

	// if blockChain thinks we are current and we have no syncPeer it
	// is probably right.
	if b.syncPeer == nil {
		return true
	}

	_, height, err := btcd.db.NewestSha()
	// No matter what chain thinks, if we are below the block we are
	// syncing to we are not current.
	// TODO(oga) we can get chain to return the height of each block when we
	// parse an orphan, which would allow us to update the height of peers
	// from what it was at initial handshake.
	if err != nil || height < int64(b.syncPeer.lastBlock) {
		return false
	}
*/
//return true
//}

// HaveBlock returns whether or not the chain instance has the block represented
// by the passed hash.  This includes checking the various places a block can
// be like part of the main chain, on a side chain, or in the orphan pool.
//
// This function is NOT safe for concurrent access.
func HaveBlockInDChain(dChain *common.DChain, hash *wire.ShaHash) (bool, error) {

	dblock, _ := db.FetchDBlockByHash(hash.ToFactomHash())
	if dblock != nil {
		return true, nil
	}
	return false, nil
}
