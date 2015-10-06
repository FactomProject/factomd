// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"encoding/hex"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/database"
	"github.com/FactomProject/btcd/blockchain"
	"github.com/FactomProject/btcd/wire"
	"github.com/davecgh/go-spew/spew"
	"time"
)

// handleFBlockMsg is invoked when a peer receives a factoid block message.
func (p *peer) handleFBlockMsg(msg *wire.MsgFBlock, buf []byte) {
	binary, _ := msg.SC.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeFactomFBlock, hash)
	p.AddKnownInventory(iv)
	inMsgQueue <- msg
}

// handleDirBlockMsg is invoked when a peer receives a dir block message.
func (p *peer) handleDirBlockMsg(msg *wire.MsgDirBlock, buf []byte) {
	binary, _ := msg.DBlk.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeFactomDirBlock, hash)
	p.AddKnownInventory(iv)

	p.pushGetNonDirDataMsg(msg.DBlk)

	inMsgQueue <- msg

	delete(p.requestedBlocks, *hash)
	delete(p.server.blockManager.requestedBlocks, *hash)
}

// handleABlockMsg is invoked when a peer receives a entry credit block message.
func (p *peer) handleABlockMsg(msg *wire.MsgABlock, buf []byte) {
	binary, _ := msg.ABlk.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeFactomAdminBlock, hash)
	p.AddKnownInventory(iv)
	inMsgQueue <- msg
}

// handleECBlockMsg is invoked when a peer receives a entry credit block
// message.
func (p *peer) handleECBlockMsg(msg *wire.MsgECBlock, buf []byte) {
	headerHash, err := msg.ECBlock.HeaderHash()
	if err != nil {
		panic(err)
	}
	hash := wire.FactomHashToShaHash(headerHash)

	iv := wire.NewInvVect(wire.InvTypeFactomEntryCreditBlock, hash)
	p.AddKnownInventory(iv)

	inMsgQueue <- msg
}

// handleEBlockMsg is invoked when a peer receives an entry block bitcoin message.
func (p *peer) handleEBlockMsg(msg *wire.MsgEBlock, buf []byte) {
	binary, _ := msg.EBlk.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeFactomEntryBlock, hash)
	p.AddKnownInventory(iv)

	p.pushGetEntryDataMsg(msg.EBlk)

	inMsgQueue <- msg

}

// handleEntryMsg is invoked when a peer receives a EBlock Entry message.
func (p *peer) handleEntryMsg(msg *wire.MsgEntry, buf []byte) {
	binary, _ := msg.Entry.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeFactomEntry, hash)
	p.AddKnownInventory(iv)

	inMsgQueue <- msg
}

// handleGetEntryDataMsg is invoked when a peer receives a get entry data message and
// is used to deliver entry of EBlock information.
func (p *peer) handleGetEntryDataMsg(msg *wire.MsgGetEntryData) {
	numAdded := 0
	notFound := wire.NewMsgNotFound()

	// We wait on the this wait channel periodically to prevent queueing
	// far more data than we can send in a reasonable time, wasting memory.
	// The waiting occurs after the database fetch for the next one to
	// provide a little pipelining.

	var waitChan chan struct{}
	doneChan := make(chan struct{}, 1)
	for i, iv := range msg.InvList {

		var c chan struct{}
		// If this will be the last message we send.
		if i == len(msg.InvList)-1 && len(notFound.InvList) == 0 {
			c = doneChan
		} else { //if (i+1)%3 == 0 {
			// Buffered so as to not make the send goroutine block.
			c = make(chan struct{}, 1)
		}

		if iv.Type != wire.InvTypeFactomEntry {
			continue
		}

		// Is this right? what is iv.hash?
		blk, err := db.FetchEBlockByHash(iv.Hash.ToFactomHash())

		if err != nil {

			if doneChan != nil {
				doneChan <- struct{}{}
			}
			return
		}

		for _, ebEntry := range blk.Body.EBEntries {

			//Skip the minute markers
			if ebEntry.IsMinuteMarker() {
				continue
			}
			var err error
			err = p.pushEntryMsg(ebEntry, c, waitChan)
			if err != nil {
				notFound.AddInvVect(iv)
				// When there is a failure fetching the final entry
				// and the done channel was sent in due to there
				// being no outstanding not found inventory, consume
				// it here because there is now not found inventory
				// that will use the channel momentarily.
				if i == len(msg.InvList)-1 && c != nil {
					<-c
				}
			}
			numAdded++
			waitChan = c
		}

	}
	if len(notFound.InvList) != 0 {
		p.QueueMessage(notFound, doneChan)
	}

	// Wait for messages to be sent. We can send quite a lot of data at this
	// point and this will keep the peer busy for a decent amount of time.
	// We don't process anything else by them in this time so that we
	// have an idea of when we should hear back from them - else the idle
	// timeout could fire when we were only half done sending the blocks.
	if numAdded > 0 {
		<-doneChan
	}
}

// handleGetNonDirDataMsg is invoked when a peer receives a dir block message.
// It returns the corresponding data block like Factoid block,
// EC block, Entry block, and Entry based on directory block's ChainID
func (p *peer) handleGetNonDirDataMsg(msg *wire.MsgGetNonDirData) {
	numAdded := 0
	notFound := wire.NewMsgNotFound()

	// We wait on the this wait channel periodically to prevent queueing
	// far more data than we can send in a reasonable time, wasting memory.
	// The waiting occurs after the database fetch for the next one to
	// provide a little pipelining.

	var waitChan chan struct{}
	doneChan := make(chan struct{}, 1)
	for i, iv := range msg.InvList {
		var c chan struct{}
		// If this will be the last message we send.
		if i == len(msg.InvList)-1 && len(notFound.InvList) == 0 {
			c = doneChan
		} else { //if (i+1)%3 == 0 {
			// Buffered so as to not make the send goroutine block.
			c = make(chan struct{}, 1)
		}

		if iv.Type != wire.InvTypeFactomNonDirBlock {
			continue
		}

		// Is this right? what is iv.hash?
		blk, err := db.FetchDBlockByHash(iv.Hash.ToFactomHash())

		if err != nil {
			peerLog.Tracef("Unable to fetch requested EC block sha %v: %v",
				iv.Hash, err)

			if doneChan != nil {
				doneChan <- struct{}{}
			}
			return
		}

		for _, dbEntry := range blk.DBEntries {

			var err error
			switch dbEntry.ChainID.String() {
			case hex.EncodeToString(common.EC_CHAINID[:]):
				err = p.pushECBlockMsg(dbEntry.KeyMR, c, waitChan)

			case hex.EncodeToString(common.ADMIN_CHAINID[:]):
				err = p.pushABlockMsg(dbEntry.KeyMR, c, waitChan)

			case wire.FChainID.String():
				err = p.pushFBlockMsg(dbEntry.KeyMR, c, waitChan)

			default:
				err = p.pushEBlockMsg(dbEntry.KeyMR, c, waitChan)
				//continue
			}
			if err != nil {
				notFound.AddInvVect(iv)
				// When there is a failure fetching the final entry
				// and the done channel was sent in due to there
				// being no outstanding not found inventory, consume
				// it here because there is now not found inventory
				// that will use the channel momentarily.
				if i == len(msg.InvList)-1 && c != nil {
					<-c
				}
			}
			numAdded++
			waitChan = c
		}

	}
	if len(notFound.InvList) != 0 {
		p.QueueMessage(notFound, doneChan)
	}

	// Wait for messages to be sent. We can send quite a lot of data at this
	// point and this will keep the peer busy for a decent amount of time.
	// We don't process anything else by them in this time so that we
	// have an idea of when we should hear back from them - else the idle
	// timeout could fire when we were only half done sending the blocks.
	if numAdded > 0 {
		<-doneChan
	}
}

// handleDirInvMsg is invoked when a peer receives an inv bitcoin message and is
// used to examine the inventory being advertised by the remote peer and react
// accordingly.  We pass the message down to blockmanager which will call
// QueueMessage with any appropriate responses.
func (p *peer) handleDirInvMsg(msg *wire.MsgDirInv) {
	p.server.blockManager.QueueDirInv(msg, p)
}

// handleGetDirDataMsg is invoked when a peer receives a getdata bitcoin message and
// is used to deliver block and transaction information.
func (p *peer) handleGetDirDataMsg(msg *wire.MsgGetDirData) {
	numAdded := 0
	notFound := wire.NewMsgNotFound()

	// We wait on the this wait channel periodically to prevent queueing
	// far more data than we can send in a reasonable time, wasting memory.
	// The waiting occurs after the database fetch for the next one to
	// provide a little pipelining.
	var waitChan chan struct{}
	doneChan := make(chan struct{}, 1)

	for i, iv := range msg.InvList {
		var c chan struct{}
		// If this will be the last message we send.
		if i == len(msg.InvList)-1 && len(notFound.InvList) == 0 {
			c = doneChan
		} else if (i+1)%3 == 0 {
			// Buffered so as to not make the send goroutine block.
			c = make(chan struct{}, 1)
		}
		var err error
		switch iv.Type {
		//case wire.InvTypeTx:
		//err = p.pushTxMsg(&iv.Hash, c, waitChan)
		case wire.InvTypeFactomDirBlock:
			err = p.pushDirBlockMsg(&iv.Hash, c, waitChan)
			/*
				case wire.InvTypeFilteredBlock:
					err = p.pushMerkleBlockMsg(&iv.Hash, c, waitChan)
			*/
		default:
			peerLog.Warnf("Unknown type in inventory request %d",
				iv.Type)
			continue
		}
		if err != nil {
			notFound.AddInvVect(iv)

			// When there is a failure fetching the final entry
			// and the done channel was sent in due to there
			// being no outstanding not found inventory, consume
			// it here because there is now not found inventory
			// that will use the channel momentarily.
			if i == len(msg.InvList)-1 && c != nil {
				<-c
			}
		}
		numAdded++
		waitChan = c
	}
	if len(notFound.InvList) != 0 {
		p.QueueMessage(notFound, doneChan)
	}

	// Wait for messages to be sent. We can send quite a lot of data at this
	// point and this will keep the peer busy for a decent amount of time.
	// We don't process anything else by them in this time so that we
	// have an idea of when we should hear back from them - else the idle
	// timeout could fire when we were only half done sending the blocks.
	if numAdded > 0 {
		<-doneChan
	}
}

// handleGetDirBlocksMsg is invoked when a peer receives a getdirblocks factom message.
func (p *peer) handleGetDirBlocksMsg(msg *wire.MsgGetDirBlocks) {
	// Return all block hashes to the latest one (up to max per message) if
	// no stop hash was specified.
	// Attempt to find the ending index of the stop hash if specified.
	endIdx := database.AllShas //factom db
	if !msg.HashStop.IsEqual(&zeroHash) {
		height, err := db.FetchBlockHeightBySha(&msg.HashStop)
		if err == nil {
			endIdx = height + 1
		}
	}

	// Find the most recent known block based on the block locator.
	// Use the block after the genesis block if no other blocks in the
	// provided locator are known.  This does mean the client will start
	// over with the genesis block if unknown block locators are provided.
	// This mirrors the behavior in the reference implementation.
	startIdx := int64(1)
	for _, hash := range msg.BlockLocatorHashes {
		height, err := db.FetchBlockHeightBySha(hash)
		if err == nil {
			// Start with the next hash since we know this one.
			startIdx = height + 1
			break
		}

	}

	peerLog.Info("startIdx=", startIdx, ", endIdx=", endIdx)

	// Don't attempt to fetch more than we can put into a single message.
	autoContinue := false
	if endIdx-startIdx > wire.MaxBlocksPerMsg {
		endIdx = startIdx + wire.MaxBlocksPerMsg
		autoContinue = true
	}

	// Generate inventory message.
	//
	// The FetchBlockBySha call is limited to a maximum number of hashes
	// per invocation.  Since the maximum number of inventory per message
	// might be larger, call it multiple times with the appropriate indices
	// as needed.
	invMsg := wire.NewMsgDirInv()
	for start := startIdx; start < endIdx; {
		// Fetch the inventory from the block database.
		hashList, err := db.FetchHeightRange(start, endIdx)
		if err != nil {
			peerLog.Warnf("Dir Block lookup failed: %v", err)
			return
		}

		// The database did not return any further hashes.  Break out of
		// the loop now.
		if len(hashList) == 0 {
			break
		}

		// Add dir block inventory to the message.
		for _, hash := range hashList {
			hashCopy := hash
			iv := wire.NewInvVect(wire.InvTypeFactomDirBlock, &hashCopy)
			invMsg.AddInvVect(iv)
		}
		start += int64(len(hashList))
	}

	// Send the inventory message if there is anything to send.
	if len(invMsg.InvList) > 0 {
		invListLen := len(invMsg.InvList)
		if autoContinue && invListLen == wire.MaxBlocksPerMsg {
			// Intentionally use a copy of the final hash so there
			// is not a reference into the inventory slice which
			// would prevent the entire slice from being eligible
			// for GC as soon as it's sent.
			continueHash := invMsg.InvList[invListLen-1].Hash
			p.continueHash = &continueHash
		}
		p.QueueMessage(invMsg, nil)
	}
}

// pushDirBlockMsg sends a dir block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushDirBlockMsg(sha *wire.ShaHash, doneChan, waitChan chan struct{}) error {
	commonhash := new(common.Hash)
	commonhash.SetBytes(sha.Bytes())
	blk, err := db.FetchDBlockByHash(commonhash)

	if err != nil {
		peerLog.Tracef("Unable to fetch requested dir block sha %v: %v",
			sha, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	// We only send the channel for this message if we aren't sending(sha)
	// an inv straight after.
	var dc chan struct{}
	sendInv := p.continueHash != nil && p.continueHash.IsEqual(sha)
	if !sendInv {
		dc = doneChan
	}
	msg := wire.NewMsgDirBlock()
	msg.DBlk = blk
	p.QueueMessage(msg, dc) //blk.MsgBlock(), dc)

	// When the peer requests the final block that was advertised in
	// response to a getblocks message which requested more blocks than
	// would fit into a single message, send it a new inventory message
	// to trigger it to issue another getblocks message for the next
	// batch of inventory.
	if p.continueHash != nil && p.continueHash.IsEqual(sha) {
		peerLog.Debug("continueHash: " + spew.Sdump(sha))
		// Sleep for 5 seconds for the peer to catch up
		time.Sleep(5 * time.Second)

		//
		// Note: Rather than the latest block height, we should pass
		// the last block height of this batch of wire.MaxBlockLocatorsPerMsg
		// to signal this is the end of the batch and
		// to trigger a client to send a new GetDirBlocks message
		//
		//hash, _, err := db.FetchBlockHeightCache()
		//if err == nil {
		invMsg := wire.NewMsgDirInvSizeHint(1)
		iv := wire.NewInvVect(wire.InvTypeFactomDirBlock, sha) //hash)
		invMsg.AddInvVect(iv)
		p.QueueMessage(invMsg, doneChan)
		p.continueHash = nil
		//} else if doneChan != nil {
		if doneChan != nil {
			doneChan <- struct{}{}
		}
	}
	return nil
}

// PushGetDirBlocksMsg sends a getdirblocks message for the provided block locator
// and stop hash.  It will ignore back-to-back duplicate requests.
func (p *peer) PushGetDirBlocksMsg(locator blockchain.BlockLocator, stopHash *wire.ShaHash) error {

	// Extract the begin hash from the block locator, if one was specified,
	// to use for filtering duplicate getblocks requests.
	// request.
	var beginHash *wire.ShaHash
	if len(locator) > 0 {
		beginHash = locator[0]
	}

	// Filter duplicate getdirblocks requests.
	if p.prevGetBlocksStop != nil && p.prevGetBlocksBegin != nil &&
		beginHash != nil && stopHash.IsEqual(p.prevGetBlocksStop) &&
		beginHash.IsEqual(p.prevGetBlocksBegin) {

		peerLog.Tracef("Filtering duplicate [getdirblocks] with begin "+
			"hash %v, stop hash %v", beginHash, stopHash)
		return nil
	}

	// Construct the getblocks request and queue it to be sent.
	msg := wire.NewMsgGetDirBlocks(stopHash)
	for _, hash := range locator {
		err := msg.AddBlockLocatorHash(hash)
		if err != nil {
			return err
		}
	}
	p.QueueMessage(msg, nil)

	// Update the previous getblocks request information for filtering
	// duplicates.
	p.prevGetBlocksBegin = beginHash
	p.prevGetBlocksStop = stopHash
	return nil
}

// pushGetNonDirDataMsg takes the passed DBlock
// and return corresponding data block like Factoid block,
// EC block, Entry block, and Entry
func (p *peer) pushGetNonDirDataMsg(dblock *common.DirectoryBlock) {
	binary, _ := dblock.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeFactomNonDirBlock, hash)
	gdmsg := wire.NewMsgGetNonDirData()
	gdmsg.AddInvVect(iv)
	if len(gdmsg.InvList) > 0 {
		p.QueueMessage(gdmsg, nil)
	}
}

// pushGetEntryDataMsg takes the passed EBlock
// and return all the corresponding EBEntries
func (p *peer) pushGetEntryDataMsg(eblock *common.EBlock) {
	binary, _ := eblock.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeFactomEntry, hash)
	gdmsg := wire.NewMsgGetEntryData()
	gdmsg.AddInvVect(iv)
	if len(gdmsg.InvList) > 0 {
		p.QueueMessage(gdmsg, nil)
	}
}

// pushFBlockMsg sends an factoid block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushFBlockMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	blk, err := db.FetchFBlockByHash(commonhash)

	if err != nil || blk == nil {
		peerLog.Tracef("Unable to fetch requested SC block sha %v: %v",
			commonhash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgFBlock()
	msg.SC = blk
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// pushABlockMsg sends an admin block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushABlockMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	blk, err := db.FetchABlockByHash(commonhash)

	if err != nil || blk == nil {
		peerLog.Tracef("Unable to fetch requested Admin block sha %v: %v",
			commonhash, err)
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgABlock()
	msg.ABlk = blk
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// pushECBlockMsg sends a entry credit block message for the provided block
// hash to the connected peer.  An error is returned if the block hash is not
// known.
func (p *peer) pushECBlockMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	blk, err := db.FetchECBlockByHash(commonhash)
	if err != nil || blk == nil {
		peerLog.Tracef("Unable to fetch requested Entry Credit block sha %v: %v",
			commonhash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgECBlock()
	msg.ECBlock = blk
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// pushEBlockMsg sends a entry block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushEBlockMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	blk, err := db.FetchEBlockByMR(commonhash)
	if err != nil {
		if doneChan != nil || blk == nil {
			peerLog.Tracef("Unable to fetch requested Entry block sha %v: %v",
				commonhash, err)
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgEBlock()
	msg.EBlk = blk
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// pushEntryMsg sends a EBlock entry message for the provided ebentry hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushEntryMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	entry, err := db.FetchEntryByHash(commonhash)
	if err != nil || entry == nil {
		peerLog.Tracef("Unable to fetch requested Entry sha %v: %v",
			commonhash, err)
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgEntry()
	msg.Entry = entry
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// handleFactoidMsg
func (p *peer) handleFactoidMsg(msg *wire.MsgFactoidTX, buf []byte) {
	binary, _ := msg.Transaction.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes())

	iv := wire.NewInvVect(wire.InvTypeTx, hash)
	p.AddKnownInventory(iv)

	inMsgQueue <- msg
}
