// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package process

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/FactomProject/factomd/btcd/wire"
	fct "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/factoid/state"
	"github.com/FactomProject/factomd/consensus"
	cp "github.com/FactomProject/factomd/controlpanel"
	"github.com/FactomProject/factomd/logger"
	"github.com/FactomProject/factomd/util"
	"github.com/davecgh/go-spew/spew"
	"runtime/debug"
	"sort"
	"strconv"

	. "github.com/FactomProject/factomd/common"
	. "github.com/FactomProject/factomd/common/AdminBlock"
	. "github.com/FactomProject/factomd/common/DirectoryBlock"
	. "github.com/FactomProject/factomd/common/EntryBlock"
	. "github.com/FactomProject/factomd/common/EntryCreditBlock"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

var _ = debug.PrintStack

// Initialize Directory Block Chain from database
func initDChain() {
	dchain = new(DChain)

	//Initialize the Directory Block Chain ID
	dchain.ChainID = new(Hash)
	barray := D_CHAINID
	dchain.ChainID.SetBytes(barray)

	// get all dBlocks from db
	dBlocks, _ := db.FetchAllDBlocks()
	sort.Sort(util.ByDBlockIDAccending(dBlocks))

	dchain.Blocks = make([]*DirectoryBlock, len(dBlocks), len(dBlocks)+1)

	for i := 0; i < len(dBlocks); i = i + 1 {
		if dBlocks[i].Header.DBHeight != uint32(i) {
			panic("Error in initializing dChain:" + dchain.ChainID.String())
		}
		dBlocks[i].Chain = dchain
		dBlocks[i].IsSealed = true
		dBlocks[i].IsSavedInDB = true
		dchain.Blocks[i] = &dBlocks[i]
	}

	// double check the block ids
	for i := 0; i < len(dchain.Blocks); i = i + 1 {
		if uint32(i) != dchain.Blocks[i].Header.DBHeight {
			panic(errors.New("BlockID does not equal index for chain:" + dchain.ChainID.String() + " block:" + fmt.Sprintf("%v", dchain.Blocks[i].Header.DBHeight)))
		}
	}

	//Create an empty block and append to the chain
	if len(dchain.Blocks) == 0 {
		dchain.NextDBHeight = 0
		dchain.NextBlock, _ = CreateDBlock(dchain, nil, 10)
	} else {
		dchain.NextDBHeight = uint32(len(dchain.Blocks))
		dchain.NextBlock, _ = CreateDBlock(dchain, dchain.Blocks[len(dchain.Blocks)-1], 10)
		// Update dir block height cache in db
		db.UpdateBlockHeightCache(dchain.NextDBHeight-1, dchain.NextBlock.Header.PrevLedgerKeyMR)
	}

	exportDChain(dchain)

	//Double check the sealed flag
	if dchain.NextBlock.IsSealed == true {
		panic("dchain.Blocks[dchain.NextBlockID].IsSealed for chain:" + dchain.ChainID.String())
	}

}

// Initialize Entry Credit Block Chain from database
func initECChain() {

	eCreditMap = make(map[string]int32)

	//Initialize the Entry Credit Chain ID
	ecchain = NewECChain()

	// get all ecBlocks from db
	ecBlocks, _ := db.FetchAllECBlocks()
	sort.Sort(util.ByECBlockIDAccending(ecBlocks))

	for i, v := range ecBlocks {
		if v.Header.DBHeight != uint32(i) {
			panic("Error in initializing dChain:" + ecchain.ChainID.String() + " DBHeight:" + strconv.Itoa(int(v.Header.DBHeight)) + " i:" + strconv.Itoa(i))
		}

		// Calculate the EC balance for each account
		initializeECreditMap(&v)
	}

	//Create an empty block and append to the chain
	if len(ecBlocks) == 0 || dchain.NextDBHeight == 0 {
		ecchain.NextBlockHeight = 0
		ecchain.NextBlock = NewECBlock()
		ecchain.NextBlock.AddEntry(serverIndex)
		for i := 0; i < 10; i++ {
			marker := NewMinuteNumber()
			marker.Number = uint8(i + 1)
			ecchain.NextBlock.AddEntry(marker)
		}
	} else {
		// Entry Credit Chain should have the same height as the dir chain
		ecchain.NextBlockHeight = dchain.NextDBHeight
		var err error
		ecchain.NextBlock, err = NextECBlock(&ecBlocks[ecchain.NextBlockHeight-1])
		if err != nil {
			panic(err)
		}
	}

	// create a backup copy before processing entries
	copyCreditMap(eCreditMap, eCreditMapBackup)
	exportECChain(ecchain)

	// ONly for debugging
	if procLog.Level() > logger.Info {
		printCreditMap()
	}

}

// Initialize Admin Block Chain from database
func initAChain() {

	//Initialize the Admin Chain ID
	achain = new(AdminChain)
	achain.ChainID = new(Hash)
	achain.ChainID.SetBytes(ADMIN_CHAINID)

	// get all aBlocks from db
	aBlocks, _ := db.FetchAllABlocks()
	sort.Sort(util.ByABlockIDAccending(aBlocks))

	// double check the block ids
	for i := 0; i < len(aBlocks); i = i + 1 {
		if uint32(i) != aBlocks[i].Header.DBHeight {
			panic(errors.New("BlockID does not equal index for chain:" + achain.ChainID.String() + " block:" + fmt.Sprintf("%v", aBlocks[i].Header.DBHeight)))
		}
		if !validateDBSignature(&aBlocks[i], dchain) {
			panic(errors.New("No valid signature found in Admin Block = " + fmt.Sprintf("%s\n", spew.Sdump(aBlocks[i]))))
		}
	}

	//Create an empty block and append to the chain
	if len(aBlocks) == 0 || dchain.NextDBHeight == 0 {
		achain.NextBlockHeight = 0
		achain.NextBlock, _ = CreateAdminBlock(achain, nil, 10)

	} else {
		// Entry Credit Chain should have the same height as the dir chain
		achain.NextBlockHeight = dchain.NextDBHeight
		achain.NextBlock, _ = CreateAdminBlock(achain, &aBlocks[achain.NextBlockHeight-1], 10)
	}

	exportAChain(achain)

}

// Initialize Factoid Block Chain from database
func initFctChain() {

	//Initialize the Admin Chain ID
	fchain = new(FctChain)
	fchain.ChainID = new(Hash)
	fchain.ChainID.SetBytes(FACTOID_CHAINID)

	// get all aBlocks from db
	fBlocks, _ := db.FetchAllFBlocks()
	sort.Sort(util.ByFBlockIDAccending(fBlocks))

	// double check the block ids
	for i := 0; i < len(fBlocks); i = i + 1 {
		if uint32(i) != fBlocks[i].GetDBHeight() {
			panic(errors.New("BlockID does not equal index for chain:" +
				fchain.ChainID.String() + " block:" +
				fmt.Sprintf("%v", fBlocks[i].GetDBHeight())))
		} else {
			FactoshisPerCredit = fBlocks[i].GetExchRate()
			state.FactoidStateGlobal.SetFactoshisPerEC(FactoshisPerCredit)
			// initialize the FactoidState in sequence
			err := state.FactoidStateGlobal.AddTransactionBlock(fBlocks[i])
			if err != nil {
				panic("Failed to rebuild factoid state: " + err.Error())
			}
		}
	}

	//Create an empty block and append to the chain
	if len(fBlocks) == 0 || dchain.NextDBHeight == 0 {
		state.FactoidStateGlobal.SetFactoshisPerEC(FactoshisPerCredit)
		fchain.NextBlockHeight = 0
		// func GetGenesisFBlock(ftime uint64, ExRate uint64, addressCnt int, Factoids uint64 ) IFBlock {
		//fchain.NextBlock = block.GetGenesisFBlock(0, FactoshisPerCredit, 10, 200000000000)
		fchain.NextBlock = block.GetGenesisFBlock()
		gb := fchain.NextBlock

		// If a client, this block is going to get downloaded and added.  Don't do it twice.
		if nodeMode == SERVER_NODE {
			err := state.FactoidStateGlobal.AddTransactionBlock(gb)
			if err != nil {
				panic(err)
			}
		}

	} else {
		fchain.NextBlockHeight = dchain.NextDBHeight
		state.FactoidStateGlobal.ProcessEndOfBlock2(dchain.NextDBHeight)
		fchain.NextBlock = state.FactoidStateGlobal.GetCurrentBlock()
	}

	exportFctChain(fchain)

}

// Initialize Entry Block Chains from database
func initEChains() {

	chainIDMap = make(map[string]*EChain)

	chains, err := db.FetchAllChains()

	if err != nil {
		panic(err)
	}

	for _, chain := range chains {
		var newChain = chain
		chainIDMap[newChain.ChainID.String()] = newChain
		exportEChain(chain)
	}

}

// Re-calculate Entry Credit Balance Map with a new Entry Credit Block
func initializeECreditMap(block *ECBlock) {
	for _, entry := range block.Body.Entries {
		// Only process: ECIDChainCommit, ECIDEntryCommit, ECIDBalanceIncrease
		switch entry.ECID() {
		case ECIDChainCommit:
			e := entry.(*CommitChain)
			eCreditMap[string(e.ECPubKey[:])] -= int32(e.Credits)
			state.FactoidStateGlobal.UpdateECBalance(fct.NewAddress(e.ECPubKey[:]), int64(e.Credits))
		case ECIDEntryCommit:
			e := entry.(*CommitEntry)
			eCreditMap[string(e.ECPubKey[:])] -= int32(e.Credits)
			state.FactoidStateGlobal.UpdateECBalance(fct.NewAddress(e.ECPubKey[:]), int64(e.Credits))
		case ECIDBalanceIncrease:
			e := entry.(*IncreaseBalance)
			eCreditMap[string(e.ECPubKey[:])] += int32(e.NumEC)
			// Don't add the Increases to Factoid state, the Factoid processing will do that.
		case ECIDServerIndexNumber:
		case ECIDMinuteNumber:
		default:
			panic("Unknow entry type:" + string(entry.ECID()) + " for ECBlock:" + strconv.FormatUint(uint64(block.Header.DBHeight), 10))
		}
	}
}

// Initialize server private key and server public key for milestone 1
func initServerKeys() {
	if nodeMode == SERVER_NODE {
		var err error
		serverPrivKey, err = NewPrivateKeyFromHex(serverPrivKeyHex)
		if err != nil {
			panic("Cannot parse Server Private Key from configuration file: " + err.Error())
		}
	}

	serverPubKey = PubKeyFromString(SERVER_PUB_KEY)

}

// Initialize the process list manager with the proper dir block height
func initProcessListMgr() {
	plMgr = consensus.NewProcessListMgr(dchain.NextDBHeight, 1, 10, serverPrivKey)

}

// Initialize the entry chains in memory from db
func initEChainFromDB(chain *EChain) {

	eBlocks, _ := db.FetchAllEBlocksByChain(chain.ChainID)
	sort.Sort(util.ByEBlockIDAccending(*eBlocks))

	for i := 0; i < len(*eBlocks); i = i + 1 {
		if uint32(i) != (*eBlocks)[i].Header.EBSequence {
			panic(errors.New("BlockID does not equal index for chain:" + chain.ChainID.String() + " block:" + fmt.Sprintf("%v", (*eBlocks)[i].Header.EBSequence)))
		}
	}

	var err error
	if len(*eBlocks) == 0 {
		chain.NextBlockHeight = 0
		chain.NextBlock, err = MakeEBlock(chain, nil)
		if err != nil {
			panic(err)
		}
	} else {
		chain.NextBlockHeight = uint32(len(*eBlocks))
		chain.NextBlock, err = MakeEBlock(chain, &(*eBlocks)[len(*eBlocks)-1])
		if err != nil {
			panic(err)
		}
	}

	// Initialize chain with the first entry (Name and rules) for non-server mode
	if nodeMode != SERVER_NODE && chain.FirstEntry == nil && len(*eBlocks) > 0 {
		chain.FirstEntry, _ = db.FetchEntryByHash((*eBlocks)[0].Body.EBEntries[0])
		if chain.FirstEntry != nil {
			db.InsertChain(chain)
		}
	}

}

// Validate dir chain from genesis block
func validateDChain(c *DChain) error {

	if nodeMode != SERVER_NODE && len(c.Blocks) == 0 {
		return nil
	}

	if uint32(len(c.Blocks)) != c.NextDBHeight {
		return errors.New("Dir chain has an un-expected Next Block ID: " + strconv.Itoa(int(c.NextDBHeight)))
	}

	//prevMR and prevBlkHash are used to validate against the block next in the chain
	prevMR, prevBlkHash, err := validateDBlock(c, c.Blocks[0])
	if err != nil {
		return err
	}

	//validate the genesis block
	//prevBlkHash is the block hash for c.Blocks[0]
	if prevBlkHash == nil || prevBlkHash.String() != GENESIS_DIR_BLOCK_HASH {

		str := fmt.Sprintf("<pre>" +
			"Expected: " + GENESIS_DIR_BLOCK_HASH + "<br>" +
			"Found:    " + prevBlkHash.String() + "</pre><br><br>")
		cp.CP.AddUpdate(
			"GenHash",                    // tag
			"warning",                    // Category
			"Genesis Hash doesn't match", // Title
			str, // Message
			0)
		// panic for Milestone 1
		panic("Genesis Block wasn't as expected:\n" +
			"    Expected: " + GENESIS_DIR_BLOCK_HASH + "\n" +
			"    Found:    " + prevBlkHash.String())

	}

	for i := 1; i < len(c.Blocks); i++ {
		if !prevBlkHash.IsSameAs(c.Blocks[i].Header.PrevLedgerKeyMR) {
			return errors.New("Previous block hash not matching for Dir block: " + strconv.Itoa(i))
		}
		if !prevMR.IsSameAs(c.Blocks[i].Header.PrevKeyMR) {
			return errors.New("Previous merkle root not matching for Dir block: " + strconv.Itoa(i))
		}
		mr, dblkHash, err := validateDBlock(c, c.Blocks[i])
		if err != nil {
			c.Blocks[i].IsValidated = false
			return err
		}

		prevMR = mr
		prevBlkHash = dblkHash
		c.Blocks[i].IsValidated = true
	}

	return nil
}

// Validate a dir block
func validateDBlock(c *DChain, b *DirectoryBlock) (merkleRoot IHash, dbHash IHash, err error) {

	bodyMR, err := b.BuildBodyMR()
	if err != nil {
		return nil, nil, err
	}

	if !b.Header.BodyMR.IsSameAs(bodyMR) {
		return nil, nil, errors.New("Invalid body MR for dir block: " + string(b.Header.DBHeight))
	}

	for _, dbEntry := range b.DBEntries {
		switch dbEntry.ChainID.String() {
		case ecchain.ChainID.String():
			err := validateCBlockByMR(dbEntry.KeyMR)
			if err != nil {
				return nil, nil, err
			}
		case achain.ChainID.String():
			err := validateABlockByMR(dbEntry.KeyMR)
			if err != nil {
				return nil, nil, err
			}
		case wire.FChainID.String():
			err := validateFBlockByMR(dbEntry.KeyMR)
			if err != nil {
				return nil, nil, err
			}
		default:
			err := validateEBlockByMR(dbEntry.ChainID, dbEntry.KeyMR)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	b.DBHash, _ = CreateHash(b)
	b.BuildKeyMerkleRoot()

	return b.KeyMR, b.DBHash, nil
}

// Validate Entry Credit Block by merkle root
func validateCBlockByMR(mr IHash) error {
	cb, _ := db.FetchECBlockByHash(mr)

	if cb == nil {
		return errors.New("Entry Credit block not found in db for merkle root: " + mr.String())
	}

	return nil
}

// Validate Admin Block by merkle root
func validateABlockByMR(mr IHash) error {
	b, _ := db.FetchABlockByHash(mr)

	if b == nil {
		return errors.New("Admin block not found in db for merkle root: " + mr.String())
	}

	return nil
}

// Validate FBlock by merkle root
func validateFBlockByMR(mr IHash) error {
	b, _ := db.FetchFBlockByHash(mr)

	if b == nil {
		return errors.New("Factoid block not found in db for merkle root: \n" + mr.String())
	}

	// check that we used the KeyMR to store the block...
	if !bytes.Equal(b.GetKeyMR().Bytes(), mr.Bytes()) {
		return fmt.Errorf("Factoid block match failure: block %d \n%s\n%s",
			b.GetDBHeight(),
			"Key in the database:   "+mr.String(),
			"Hash of the blk found: "+b.GetKeyMR().String())
	}

	return nil
}

// Validate Entry Block by merkle root
func validateEBlockByMR(cid IHash, mr IHash) error {

	eb, err := db.FetchEBlockByMR(mr)
	if err != nil {
		return err
	}

	if eb == nil {
		return errors.New("Entry block not found in db for merkle root: " + mr.String())
	}
	keyMR, err := eb.KeyMR()
	if err != nil {
		return err
	}
	if !mr.IsSameAs(keyMR) {
		return errors.New("Entry block's merkle root does not match with: " + mr.String())
	}

	for _, ebEntry := range eb.Body.EBEntries {
		if !bytes.Equal(ebEntry.Bytes()[:31], ZERO_HASH[:31]) {
			entry, _ := db.FetchEntryByHash(ebEntry)
			if entry == nil {
				return errors.New("Entry not found in db for entry hash: " + ebEntry.String())
			}
		} // Else ... we could do a bit more validation of the minute markers.
	}

	return nil
}
