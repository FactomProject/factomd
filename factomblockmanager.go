// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"container/list"
	"sync/atomic"

	"github.com/FactomProject/btcd/wire"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/util"
	"github.com/davecgh/go-spew/spew"
)

// dirBlockMsg packages a directory block message and the peer it came from together
// so the block handler has access to that information.
type dirBlockMsg struct {
	block *common.DirectoryBlock
	peer  *peer
}

// dirInvMsg packages a dir block inv message and the peer it came from together
// so the block handler has access to that information.
type dirInvMsg struct {
	inv  *wire.MsgDirInv
	peer *peer
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
			/*
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
				}*/

			// We already have the final block advertised by this
			// inventory message, so force a request for more.  This
			// should only happen if we're on a really long side
			// chain.
			if i == lastBlock {
				// Request blocks after this one up to the
				// final one the remote peer knows about (zero
				// stop hash).
				locator := DirBlockLocatorFromHash(&iv.Hash)
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
	util.Trace()
	// Don't accept more blocks if we're shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		p.blockProcessed <- struct{}{}
		return
	}

	b.msgChan <- &dirBlockMsg{block: msg.DBlk, peer: p}
}

// QueueDirInv adds the passed inv message and peer to the block handling queue.
func (b *blockManager) QueueDirInv(inv *wire.MsgDirInv, p *peer) {
	util.Trace()
	// No channel handling here because peers do not need to block on inv
	// messages.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}

	b.msgChan <- &dirInvMsg{inv: inv, peer: p}
}

// startSyncFactom will choose the best peer among the available candidate peers to
// download/sync the blockchain from.  When syncing is already running, it
// simply returns.  It also examines the candidates for any which are no longer
// candidates and removes them as needed.
func (b *blockManager) startSyncFactom(peers *list.List) {
	util.Trace()
	// Return now if we're already syncing.
	if b.syncPeer != nil {
		return
	}

	// Find the height of the current known best block.
	_, height, err := db.FetchBlockHeightCache()
	if err != nil {
		bmgrLog.Errorf("%v", err)
		return
	}

	bmgrLog.Infof("Latest DirBlock Height: %d", height)

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
		util.Trace()
		locator, err := LatestDirBlockLocator()
		if err != nil {
			bmgrLog.Errorf("Failed to get block locator for the "+
				"latest block: %v", err)
			return
		}

		bmgrLog.Infof("LatestDirBlockLocator: %s", spew.Sdump(locator))

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
		/*
			if b.nextCheckpoint != nil && height < b.nextCheckpoint.Height &&
				!cfg.RegressionTest && !cfg.DisableCheckpoints {

				bestPeer.PushGetHeadersMsg(locator, b.nextCheckpoint.Hash)
				b.headersFirstMode = true
				bmgrLog.Infof("Downloading headers for blocks %d to "+
					"%d from peer %s", height+1,
					b.nextCheckpoint.Height, bestPeer.addr)
			} else {*/
		bestPeer.PushGetDirBlocksMsg(locator, &zeroHash)
		//}
		b.syncPeer = bestPeer
	} else {
		bmgrLog.Warnf("No sync peer candidates available")
	}
}

// isSyncCandidateFactom returns whether or not the peer is a candidate to consider
// syncing from.
func (b *blockManager) isSyncCandidateFactom(p *peer) bool {
	/*
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
	*/
	util.Trace()
	return true
}

// HaveBlockInDB returns whether or not the chain instance has the block represented
// by the passed hash.  This includes checking the various places a block can
// be like part of the main chain, on a side chain, or in the orphan pool.
//
// This function is NOT safe for concurrent access.
func HaveBlockInDB(hash *wire.ShaHash) (bool, error) {
	util.Trace()
	dblock, _ := db.FetchDBlockByHash(hash.ToFactomHash())
	if dblock != nil {
		return true, nil
	}
	return false, nil
}

/*
// currentDChain returns true if we believe we are synced with our peers, false if we
// still have blocks to check
func (b *blockManager) currentDChain() bool {

	if !b.dirChain.IsCurrent(b.server.timeSource) {
		return false
	}

	// if blockChain thinks we are current and we have no syncPeer it
	// is probably right.
	if b.syncPeer == nil {
		return true
	}

	_, height, err := db.NewestSha()
	// No matter what chain thinks, if we are below the block we are
	// syncing to we are not current.
	// TODO(oga) we can get chain to return the height of each block when we
	// parse an orphan, which would allow us to update the height of peers
	// from what it was at initial handshake.
	if err != nil || height < int64(b.syncPeer.lastBlock) {
		return false
	}

	return true
}
*/
