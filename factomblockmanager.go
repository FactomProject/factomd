// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"container/list"
	"fmt"
	"github.com/FactomProject/FactomCode/common"
	cp "github.com/FactomProject/FactomCode/controlpanel"
	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/wire"
	"github.com/davecgh/go-spew/spew"
	"sync/atomic"
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
	bmgrLog.Debug("handleDirInvMsg: ", spew.Sdump(imsg))

	// Ignore invs from peers that aren't the sync if we are not current.
	// Helps prevent fetching a mass of orphans.
	if imsg.peer != b.syncPeer && !b.current() {
		return
	}

	// Attempt to find the final block in the inventory list.  There may
	// not be one.
	lastBlock := -1
	invVects := imsg.inv.InvList
	bmgrLog.Debugf("len(InvVects)=%d", len(invVects))
	for i := len(invVects) - 1; i >= 0; i-- {
		if invVects[i].Type == wire.InvTypeFactomDirBlock {
			lastBlock = i
			bmgrLog.Debugf("lastBlock=%d", lastBlock)
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

			// We already have the final block advertised by this
			// inventory message, so force a request for more.  This
			// should only happen if we're on a really long side
			// chain.
			if i == lastBlock {
				// Request blocks after this one up to the
				// final one the remote peer knows about (zero
				// stop hash).
				bmgrLog.Debug("push for more dir blocks: PushGetDirBlocksMsg")
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
	// Don't accept more blocks if we're shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		p.blockProcessed <- struct{}{}
		return
	}

	b.msgChan <- &dirBlockMsg{block: msg.DBlk, peer: p}
}

// QueueDirInv adds the passed inv message and peer to the block handling queue.
func (b *blockManager) QueueDirInv(inv *wire.MsgDirInv, p *peer) {
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
		locator, err := LatestDirBlockLocator()
		if err != nil {
			bmgrLog.Errorf("Failed to get block locator for the "+
				"latest block: %v", err)
			return
		}

		bmgrLog.Infof("LatestDirBlockLocator: %s", spew.Sdump(locator))

		str := fmt.Sprintf("At %d: syncing to block height %d from peer %v",
			height, bestPeer.lastBlock, bestPeer.addr)
		bmgrLog.Infof(str)

		cp.CP.AddUpdate(
			"Syncing", // tag
			"status",  // Category
			"Client is Syncing with Federated Server(s)", // Title
			str, // Message
			60)

		bestPeer.PushGetDirBlocksMsg(locator, &zeroHash)
		b.syncPeer = bestPeer
	} else {
		bmgrLog.Warnf("No sync peer candidates available")
	}
}

// isSyncCandidateFactom returns whether or not the peer is a candidate to consider
// syncing from.
func (b *blockManager) isSyncCandidateFactom(p *peer) bool {
	// Typically a peer is not a candidate for sync if it's not a Factom SERVER node,
	if common.SERVER_NODE == util.ReadConfig().App.NodeMode {
		return true
	}
	return true
}
