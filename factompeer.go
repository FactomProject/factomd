// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"fmt"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/blockchain"
	"github.com/FactomProject/btcd/database"
	"github.com/FactomProject/btcd/wire"
	"github.com/davecgh/go-spew/spew"
)

// handleDirBlockMsg is invoked when a peer receives a dir block message.
func (p *peer) handleDirBlockMsg(msg *wire.MsgDirBlock, buf []byte) {
	util.Trace()
	// Convert the raw MsgBlock to a btcutil.Block which provides some
	// convenience methods and things such as hash caching.

	fmt.Printf("msgDirBlock=%v\n", spew.Sdump(msg.DBlk))

	binary, _ := msg.DBlk.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)

	iv := wire.NewInvVect(wire.InvTypeFactomDirBlock, hash)
	p.AddKnownInventory(iv)

	p.pushGetNonDirDataMsg(msg.DBlk)
	
	inMsgQueue <- msg

}

// handleABlockMsg is invoked when a peer receives a entry credit block message.
func (p *peer) handleABlockMsg(msg *wire.MsgABlock, buf []byte) {
	util.Trace()
	// Convert the raw MsgBlock to a btcutil.Block which provides some
	// convenience methods and things such as hash caching.

	fmt.Printf("msgABlock=%v\n", spew.Sdump(msg.ABlk))

	binary, _ := msg.ABlk.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)

	iv := wire.NewInvVect(wire.InvTypeFactomAdminBlock, hash)
	p.AddKnownInventory(iv)

	inMsgQueue <- msg
}

// handleECBlockMsg is invoked when a peer receives a entry credit block
// message.
func (p *peer) handleECBlockMsg(msg *wire.MsgECBlock, buf []byte) {
	util.Trace()
	// Convert the raw MsgBlock to a btcutil.Block which provides some
	// convenience methods and things such as hash caching.

	fmt.Printf("msgECBlock=%v\n", spew.Sdump(msg.ECBlock))

	binary, _ := msg.ECBlock.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)

	iv := wire.NewInvVect(wire.InvTypeFactomEntryCreditBlock, hash)
	p.AddKnownInventory(iv)

	inMsgQueue <- msg
}

// handleEBlockMsg is invoked when a peer receives an entry block bitcoin message.
func (p *peer) handleEBlockMsg(msg *wire.MsgEBlock, buf []byte) {
	util.Trace()
	// Convert the raw MsgBlock to a btcutil.Block which provides some
	// convenience methods and things such as hash caching.

	fmt.Printf("msgEBlock=%v\n", spew.Sdump(msg.EBlk))

	binary, _ := msg.EBlk.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)

	iv := wire.NewInvVect(wire.InvTypeFactomEntryBlock, hash)
	p.AddKnownInventory(iv)

	p.pushGetEntryDataMsg(msg.EBlk)
	
	inMsgQueue <- msg


}

// handleEntryMsg is invoked when a peer receives a EBlock Entry message.
func (p *peer) handleEntryMsg(msg *wire.MsgEntry, buf []byte) {
	util.Trace()
	// Convert the raw MsgBlock to a btcutil.Block which provides some
	// convenience methods and things such as hash caching.

	fmt.Printf("msgEntry=%v\n", spew.Sdump(msg.Entry))

	binary, _ := msg.Entry.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)

	iv := wire.NewInvVect(wire.InvTypeFactomEntry, hash)
	p.AddKnownInventory(iv)

	inMsgQueue <- msg
}

// handleGetEntryDataMsg is invoked when a peer receives a get entry data message and
// is used to deliver entry of EBlock information.
func (p *peer) handleGetEntryDataMsg(msg *wire.MsgGetEntryData) {
	util.Trace()
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
			peerLog.Tracef("Unable to fetch requested EBlock sha %v: %v",
				iv.Hash, err)

			if doneChan != nil {
				doneChan <- struct{}{}
			}
			return
		}

		fmt.Printf("commonHash=%s, entry block=%s\n", iv.Hash.ToFactomHash().String(), spew.Sdump(blk))

		for _, ebEntry := range blk.EBEntries {

			var err error
			err = p.pushEntryMsg(ebEntry.EntryHash, c, waitChan)
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
	util.Trace()
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

		fmt.Printf("commonHash=%s, directory block=%s\n", iv.Hash.ToFactomHash().String(), spew.Sdump(blk))

		for _, dbEntry := range blk.DBEntries {

			var err error
			switch dbEntry.ChainID.String() {
			case ecchain.ChainID.String():
				err = p.pushECBlockMsg(dbEntry.MerkleRoot, c, waitChan)

			case achain.ChainID.String():
				err = p.pushABlockMsg(dbEntry.MerkleRoot, c, waitChan)
				
			case wire.FChainID.String():
				err = p.pushBlockMsg(wire.FactomHashToShaHash(dbEntry.MerkleRoot), c, waitChan)

			default:
				err = p.pushEBlockMsg(dbEntry.MerkleRoot, c, waitChan)
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
	util.Trace()
	p.server.blockManager.QueueDirInv(msg, p)
}

// handleGetDirDataMsg is invoked when a peer receives a getdata bitcoin message and
// is used to deliver block and transaction information.
func (p *peer) handleGetDirDataMsg(msg *wire.MsgGetDirData) {
	util.Trace()
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
	util.Trace()
	endHeight := int64(len(dchain.Blocks)) - 1
	endIdx := database.AllShas //factom db
	if endIdx >= 500 {
		endIdx = 500
	}
	if endIdx >= endHeight {
		endIdx = endHeight
	}

	if !msg.HashStop.IsEqual(&zeroHash) {

		//to be improved??
		commonhash := new(common.Hash)
		commonhash.SetBytes(msg.HashStop.Bytes())
		dblock, _ := db.FetchDBlockByHash(commonhash)
		if dblock != nil {
			height := int64(dblock.Header.BlockHeight)
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

		//to be improved??
		commonhash := new(common.Hash)
		commonhash.SetBytes(hash.Bytes())
		dblock, _ := db.FetchDBlockByHash(commonhash)
		if dblock != nil {
			height := int64(dblock.Header.BlockHeight)
			startIdx = height + 1
			break
		}

	}

	// Don't attempt to fetch more than we can put into a single message.
	autoContinue := false
	if endIdx-startIdx > wire.MaxBlocksPerMsg {
		endIdx = startIdx + wire.MaxBlocksPerMsg
		autoContinue = true
	}

	fmt.Printf("Newest height=%d, startIdx=%d, endIdx=%d, autoContinue=%v\n",
		endHeight, startIdx, endIdx, autoContinue)

	// Generate inventory message.
	//
	// The FetchBlockBySha call is limited to a maximum number of hashes
	// per invocation.  Since the maximum number of inventory per message
	// might be larger, call it multiple times with the appropriate indices
	// as needed.
	invMsg := wire.NewMsgDirInv()
	for start := startIdx; start < endIdx; {
		// Fetch the inventory from the block database.
		//hashList, err := db.FetchHeightRange(start, endIdx)
		// to be improved??
		hashList := make([]wire.ShaHash, 0, endIdx-startIdx)
		for i := int64(0); i < endIdx; i++ {
			newhash, _ := wire.NewShaHash(dchain.Blocks[i].DBHash.Bytes)
			hashList = append(hashList, *newhash)
			fmt.Printf("appended hash=%s\n", newhash.String())
		}

		/*		if err != nil {
					peerLog.Warnf("Block lookup failed: %v", err)
					return
				}
		*/
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
		util.Trace()
		invListLen := len(invMsg.InvList)
		if autoContinue && invListLen == wire.MaxBlocksPerMsg {
			// Intentionally use a copy of the final hash so there
			// is not a reference into the inventory slice which
			// would prevent the entire slice from being eligible
			// for GC as soon as it's sent.
			util.Trace()
			continueHash := invMsg.InvList[invListLen-1].Hash
			p.continueHash = &continueHash
		}
		p.QueueMessage(invMsg, nil)
	}
}

// pushDirBlockMsg sends a dir block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushDirBlockMsg(sha *wire.ShaHash, doneChan, waitChan chan struct{}) error {
	util.Trace()

	//to be improved??
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

	fmt.Printf("commonHash=%s, dir block=%s\n", commonhash.String(), spew.Sdump(blk))

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	// We only send the channel for this message if we aren't sending
	// an inv straight after.
	var dc chan struct{}
	sendInv := p.continueHash != nil && p.continueHash.IsEqual(sha)
	if !sendInv {
		dc = doneChan
	}
	msg := wire.NewMsgDirBlock()
	msg.DBlk = blk
	fmt.Printf("dblock=%s\n", spew.Sdump(blk))
	p.QueueMessage(msg, dc) //blk.MsgBlock(), dc)

	// When the peer requests the final block that was advertised in
	// response to a getblocks message which requested more blocks than
	// would fit into a single message, send it a new inventory message
	// to trigger it to issue another getblocks message for the next
	// batch of inventory.
	if p.continueHash != nil && p.continueHash.IsEqual(sha) {
		util.Trace()
		hash, _ := wire.NewShaHash(dchain.Blocks[dchain.NextBlockHeight-1].DBHash.Bytes) // to be improved??
		if err == nil {
			util.Trace()
			invMsg := wire.NewMsgDirInvSizeHint(1)
			iv := wire.NewInvVect(wire.InvTypeFactomDirBlock, hash)
			invMsg.AddInvVect(iv)
			p.QueueMessage(invMsg, doneChan)
			p.continueHash = nil
		} else if doneChan != nil {
			doneChan <- struct{}{}
		}
	}
	return nil
}

// PushGetDirBlocksMsg sends a getdirblocks message for the provided block locator
// and stop hash.  It will ignore back-to-back duplicate requests.
func (p *peer) PushGetDirBlocksMsg(locator blockchain.BlockLocator, stopHash *wire.ShaHash) error {
	util.Trace()

	// Extract the begin hash from the block locator, if one was specified,
	// to use for filtering duplicate getblocks requests.
	// request.
	var beginHash *wire.ShaHash
	if len(locator) > 0 {
		beginHash = locator[0]
	}

	fmt.Printf("beginHash=%s, stopHash=%s\n", beginHash.String(), stopHash.String())

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
		fmt.Printf("add dir block hash=%s\n", hash.String())
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
	util.Trace()

	binary, _ := dblock.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)

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
	util.Trace()

	binary, _ := eblock.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)

	iv := wire.NewInvVect(wire.InvTypeFactomEntry, hash)
	gdmsg := wire.NewMsgGetEntryData()
	gdmsg.AddInvVect(iv)
	if len(gdmsg.InvList) > 0 {
		p.QueueMessage(gdmsg, nil)
	}
}

// pushABlockMsg sends an admin block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushABlockMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	util.Trace()

	blk, err := db.FetchABlockByHash(commonhash)

	if err != nil {
		peerLog.Tracef("Unable to fetch requested admin block sha %v: %v",
			commonhash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	fmt.Printf("commonHash=%s, admin block=%s\n", commonhash.String(), spew.Sdump(blk))

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgABlock()
	msg.ABlk = blk
	fmt.Printf("ablock=%s\n", spew.Sdump(blk))
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// pushECBlockMsg sends a entry credit block message for the provided block
// hash to the connected peer.  An error is returned if the block hash is not
// known.
func (p *peer) pushECBlockMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {

	blk, err := db.FetchECBlockByHash(commonhash)

	if err != nil {
		peerLog.Tracef("Unable to fetch requested entry credit block sha %v: %v",
			commonhash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	fmt.Printf("commonHash=%s, entry credit block=%s\n", commonhash.String(), spew.Sdump(blk))

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgECBlock()
	msg.ECBlock = blk
	fmt.Printf("cblock=%s\n", spew.Sdump(blk))
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// pushEBlockMsg sends a entry block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushEBlockMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	util.Trace()

	blk, err := db.FetchEBlockByMR(commonhash)

	if err != nil {
		peerLog.Tracef("Unable to fetch requested entry block sha %v: %v",
			commonhash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	fmt.Printf("commonHash=%s, entry block=%s\n", commonhash.String(), spew.Sdump(blk))

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgEBlock()
	msg.EBlk = blk
	fmt.Printf("eblock=%s\n", spew.Sdump(blk))
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}

// pushEntryMsg sends a EBlock entry message for the provided ebentry hash to the
// connected peer.  An error is returned if the block hash is not known.
func (p *peer) pushEntryMsg(commonhash *common.Hash, doneChan, waitChan chan struct{}) error {
	util.Trace()

	entry, err := db.FetchEntryByHash(commonhash)

	if err != nil {
		peerLog.Tracef("Unable to fetch requested eblock entry sha %v: %v",
			commonhash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	fmt.Printf("commonHash=%s, entry=%s\n", commonhash.String(), spew.Sdump(entry))

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	msg := wire.NewMsgEntry()
	msg.Entry = entry
	fmt.Printf("Entry=%s\n", spew.Sdump(entry))
	p.QueueMessage(msg, doneChan) //blk.MsgBlock(), dc)
	return nil
}
