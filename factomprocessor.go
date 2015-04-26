// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// factomlog is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package btcd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/consensus"
	"github.com/FactomProject/FactomCode/database"
	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/btcrpcclient"
	"github.com/FactomProject/btcutil"
	"github.com/davecgh/go-spew/spew"
)

const (
	//Server running mode
	FULL_NODE   = "FULL"
	SERVER_NODE = "SERVER"
	LIGHT_NODE  = "LIGHT"
)

var (
	wclient *btcrpcclient.Client //rpc client for btcwallet rpc server
	dclient *btcrpcclient.Client //rpc client for btcd rpc server

	currentAddr btcutil.Address
	tickers     [2]*time.Ticker
	db          database.Db    // database
	dchain      *common.DChain //Directory Block Chain
	cchain      *common.CChain //Entry Credit Chain
	fchainID    *common.Hash

	creditsPerChain   int32  = 10
	creditsPerFactoid uint64 = 1000

	// To be moved to ftmMemPool??
	chainIDMap      map[string]*common.EChain // ChainIDMap with chainID string([32]byte) as key
	eCreditMap      map[string]int32          // eCreditMap with public key string([32]byte) as key, credit balance as value
	prePaidEntryMap map[string]int32          // Paid but unrevealed entries string(Etnry Hash) as key, Number of payments as value

	chainIDMapBackup      map[string]*common.EChain //previous block bakcup - ChainIDMap with chainID string([32]byte) as key
	eCreditMapBackup      map[string]int32          // backup from previous block - eCreditMap with public key string([32]byte) as key, credit balance as value
	prePaidEntryMapBackup map[string]int32          // backup from previous block - Paid but unrevealed entries string(Etnry Hash) as key, Number of payments as value

	//Diretory Block meta data map
	dbInfoMap map[string]*common.DBInfo // dbInfoMap with dbHash string([32]byte) as key

	fMemPool *ftmMemPool
	plMgr    *consensus.ProcessListMgr
)

var (
	portNumber              int  = 8083
	sendToBTCinSeconds           = 600
	directoryBlockInSeconds      = 60
	dataStorePath                = "/tmp/store/seed/"
	ldbpath                      = "/tmp/ldb9"
	nodeMode                     = "FULL"
	devNet                  bool = false

	//BTC:
	//	addrStr = "movaFTARmsaTMk3j71MpX8HtMURpsKhdra"
	walletPassphrase          = "lindasilva"
	certHomePath              = "btcwallet"
	rpcClientHost             = "localhost:18332" //btcwallet rpcserver address
	rpcClientEndpoint         = "ws"
	rpcClientUser             = "testuser"
	rpcClientPass             = "notarychain"
	btcTransFee       float64 = 0.0001

	certHomePathBtcd = "btcd"
	rpcBtcdHost      = "localhost:18334" //btcd rpcserver address

)

func LoadConfigurations(cfg *util.FactomdConfig) {

	//setting the variables by the valued form the config file
	logLevel = cfg.Log.LogLevel
	portNumber = cfg.App.PortNumber
	dataStorePath = cfg.App.DataStorePath
	ldbpath = cfg.App.LdbPath
	directoryBlockInSeconds = cfg.App.DirectoryBlockInSeconds
	nodeMode = cfg.App.NodeMode

	//addrStr = cfg.Btc.BTCPubAddr
	sendToBTCinSeconds = cfg.Btc.SendToBTCinSeconds
	walletPassphrase = cfg.Btc.WalletPassphrase
	certHomePath = cfg.Btc.CertHomePath
	rpcClientHost = cfg.Btc.RpcClientHost
	rpcClientEndpoint = cfg.Btc.RpcClientEndpoint
	rpcClientUser = cfg.Btc.RpcClientUser
	rpcClientPass = cfg.Btc.RpcClientPass
	btcTransFee = cfg.Btc.BtcTransFee
	certHomePathBtcd = cfg.Btc.CertHomePathBtcd
	rpcBtcdHost = cfg.Btc.RpcBtcdHost //btcd rpcserver address

}

func watchError(err error) {
	panic(err)
}

func readError(err error) {
	fmt.Println("error: ", err)
}

// Initialize the entry chains in memory from db
func initEChainFromDB(chain *common.EChain) {

	eBlocks, _ := db.FetchAllEBlocksByChain(chain.ChainID)
	sort.Sort(util.ByEBlockIDAccending(*eBlocks))

	for i := 0; i < len(*eBlocks); i = i + 1 {
		if uint32(i) != (*eBlocks)[i].Header.EBHeight {
			panic(errors.New("BlockID does not equal index for chain:" + chain.ChainID.String() + " block:" + fmt.Sprintf("%v", (*eBlocks)[i].Header.EBHeight)))
		}
	}

	if len(*eBlocks) == 0 {
		chain.NextBlockHeight = 0
		chain.NextBlock, _ = common.CreateBlock(chain, nil, 10)
	} else {
		chain.NextBlockHeight = uint32(len(*eBlocks))
		chain.NextBlock, _ = common.CreateBlock(chain, &(*eBlocks)[len(*eBlocks)-1], 10)
	}

	//fmt.Println("Loaded", chain.NextBlockID, "blocks for chain: " + chain.ChainID.String())

	//Get the unprocessed entries in db for the past # of mins for the open block
	binaryTimestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(binaryTimestamp, uint64(0))
	if chain.NextBlock.IsSealed == true {
		panic("chain.NextBlock.IsSealed for chain:" + chain.ChainID.String())
	}
}

func init_processor() {
	util.Trace()

	// init mem pools
	fMemPool = new(ftmMemPool)
	fMemPool.init_ftmMemPool()

	// init fchainid
	barray := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F}
	fchainID = new(common.Hash)
	fchainID.SetBytes(barray)

	// init Directory Block Chain
	initDChain()
	fmt.Println("Loaded", dchain.NextBlockHeight, "Directory blocks for chain: "+dchain.ChainID.String())

	// init Entry Credit Chain
	initCChain()
	fmt.Println("Loaded", cchain.NextBlockHeight, "Entry Credit blocks for chain: "+cchain.ChainID.String())

	// build the Genesis blocks if the current height is 0
	if dchain.NextBlockHeight == 0 {
		buildGenesisBlocks()
	}

	// init process list manager
	initProcessListMgr()

	// init Entry Chains
	initEChains()
	for _, chain := range chainIDMap {
		initEChainFromDB(chain)

		fmt.Println("Loaded", chain.NextBlockHeight, "blocks for chain: "+chain.ChainID.String())

	}

	// create EBlocks and FBlock every 60 seconds
	tickers[0] = time.NewTicker(time.Second * time.Duration(directoryBlockInSeconds))

	// write 10 FBlock in a batch to BTC every 10 minutes
	tickers[1] = time.NewTicker(time.Second * time.Duration(sendToBTCinSeconds))

	util.Trace("NOT IMPLEMENTED! IMPORTANT: Anchoring code 1 !!!")

	/*
		go func() {
			for _ = range tickers[0].C {
				fmt.Println("in tickers[0]: newEntryBlock & newFactomBlock")

				eom10 := &wire.MsgInt_EOM{
					EOM_Type: wire.END_MINUTE_10,
				}

				inCtlMsgQueue <- eom10

				/*
					// Entry Chains
					for _, chain := range chainIDMap {
						eblock := newEntryBlock(chain)
						if eblock != nil {
							dchain.AddDBEntry(eblock)
						}
						save(chain)
					}

					// Entry Credit Chain
					cBlock := newEntryCreditBlock(cchain)
					if cBlock != nil {
						dchain.AddCBlockToDBEntry(cBlock)
					}
					saveCChain(cchain)

					util.Trace("NOT IMPLEMENTED: Factoid Chain init was here !!!!!!!!!!!")

					/*
						// Factoid Chain
						fBlock := newFBlock(fchain)
						if fBlock != nil {
							dchain.AddFBlockToDBEntry(factoid.NewDBEntryFromFBlock(fBlock))
						}
						saveFChain(fchain)
					*\

					// Directory Block chain
					dbBlock := newDirectoryBlock(dchain)
					saveDChain(dchain)

					// Only Servers can write the anchor to Bitcoin network
					if nodeMode == SERVER_NODE && dbBlock != nil {
						dbInfo := common.NewDBInfoFromDBlock(dbBlock)
						saveDBMerkleRoottoBTC(dbInfo)
					}


			}
		}()
	*/
}

func Start_Processor(ldb database.Db, inMsgQ chan wire.FtmInternalMsg, outMsgQ chan wire.FtmInternalMsg,
	inCtlMsgQ chan wire.FtmInternalMsg, outCtlMsgQ chan wire.FtmInternalMsg, doneFBlockQ chan wire.FtmInternalMsg) {
	db = ldb

	inMsgQueue = inMsgQ
	outMsgQueue = outMsgQ

	inCtlMsgQueue = inCtlMsgQ
	outCtlMsgQueue = outCtlMsgQ
	doneFBlockQueue = doneFBlockQ

	init_processor()
	/* for testing??
	if nodeMode == SERVER_NODE {
		err := initRPCClient()
		if err != nil {
			log.Fatalf("cannot init rpc client: %s", err)
		}

		if err := initWallet(); err != nil {
			log.Fatalf("cannot init wallet: %s", err)
		}
	}

	*/

	// Initialize timer for the open dblock before processing messages
	if nodeMode == SERVER_NODE {
		timer := &BlockTimer{
			nextDBlockHeight: dchain.NextBlockHeight,
			inCtlMsgQueue:    inCtlMsgQueue,
		}
		go timer.StartBlockTimer()
	}

	util.Trace("before range inMsgQ")
	// Process msg from the incoming queue one by one
	for {
		select {
		case msg := <-inMsgQ:
			fmt.Printf("PROCESSOR: in inMsgQ, msg:%+v\n", msg)

			err := serveMsgRequest(msg)
			if err != nil {
				log.Println(err)
			}

		case ctlMsg := <-inCtlMsgQueue:
			fmt.Printf("PROCESSOR: in ctlMsg, msg:%+v\n", ctlMsg)

			err := serveMsgRequest(ctlMsg)
			if err != nil {
				log.Println(err)
			}
		}

	}

	util.Trace()

	defer func() {
		//		shutdown()	// was defined in factombtc.go TODO: TBD
		tickers[0].Stop()
		tickers[1].Stop()
		//db.Close()
	}()

}

func fileNotExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return true
	}
	return err != nil
}

// Serve the "fast lane" incoming control msg from inCtlMsgQueue
func serveCtlMsgRequest(msg wire.FtmInternalMsg) error {

	util.Trace()

	switch msg.Command() {
	case wire.CmdCommitChain:

	default:
		return errors.New("Message type unsupported:" + fmt.Sprintf("%+v", msg))
	}
	return nil

}

// Serve incoming msg from inMsgQueue
func serveMsgRequest(msg wire.FtmInternalMsg) error {

	util.Trace()

	switch msg.Command() {
	case wire.CmdCommitChain:
		msgCommitChain, ok := msg.(*wire.MsgCommitChain)
		if ok && msgCommitChain.IsValid() {
			err := processCommitChain(msgCommitChain)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdRevealChain:
		msgRevealChain, ok := msg.(*wire.MsgRevealChain)
		if ok {
			err := processRevealChain(msgRevealChain)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdCommitEntry:
		msgCommitEntry, ok := msg.(*wire.MsgCommitEntry)
		if ok && msgCommitEntry.IsValid() {
			err := processCommitEntry(msgCommitEntry)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdRevealEntry:
		msgRevealEntry, ok := msg.(*wire.MsgRevealEntry)
		if ok {
			err := processRevealEntry(msgRevealEntry)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdInt_FactoidObj:
		factoidObj, ok := msg.(*wire.MsgInt_FactoidObj)
		if ok {
			err := processFactoidTx(factoidObj)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdInt_EOM:
		util.Trace("CmdInt_EOM")

		if nodeMode == SERVER_NODE {
			msgEom, ok := msg.(*wire.MsgInt_EOM)
			if !ok {
				return errors.New("Error in build blocks:" + fmt.Sprintf("%+v", msg))
			}
			fmt.Printf("PROCESSOR: End of minute msg - wire.CmdInt_EOM:%+v\n", msg)

			if msgEom.EOM_Type == wire.END_MINUTE_10 {
				// Process from Orphan pool before the end of process list
				processFromOrphanPool()

				plMgr.AddMyProcessListItem(msgEom, nil, wire.END_MINUTE_10)

				//Notify the factoid component to start building factoid block
				util.Trace("Notify the factoid component to start building factoid block")

				outCtlMsgQueue <- msgEom

				err := buildBlocks()
				if err != nil {
					return err
				}
			} else if msgEom.EOM_Type >= wire.END_MINUTE_1 && msgEom.EOM_Type < wire.END_MINUTE_10 {
				plMgr.AddMyProcessListItem(msgEom, nil, msgEom.EOM_Type)
			}
		}

	case wire.CmdInt_FactoidBlock:
		factoidBlock, ok := msg.(*wire.MsgInt_FactoidBlock)
		util.Trace("Factoid Block (GENERATED??)")
		fmt.Println("factoidBlock= ", factoidBlock, " ok= ", ok)

	case wire.CmdDirBlock:
		if nodeMode == SERVER_NODE {
			break
		}

		dirBlock, ok := msg.(*wire.MsgDirBlock)
		if ok {
			err := processDirBlock(dirBlock)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdCBlock:
		if nodeMode == SERVER_NODE {
			break
		}

		cblock, ok := msg.(*wire.MsgCBlock)
		if ok {
			err := processCBlock(cblock)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdEBlock:
		if nodeMode == SERVER_NODE {
			break
		}

		eblock, ok := msg.(*wire.MsgEBlock)
		if ok {
			err := processEBlock(eblock)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}

	case wire.CmdEntry:
		if nodeMode == SERVER_NODE {
			break
		}

		entry, ok := msg.(*wire.MsgEntry)
		if ok {
			err := processEntry(entry)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
		}
		/* this should be done on the btcd side
		case wire.CmdBlock: // Factoid block
			if nodeMode == SERVER_NODE { break }

			block, ok := msg.(*wire.MsgBlock)
			if ok {
				err := processFactoidBlock(block)
				if err != nil {
					return err
				}
			} else {
				return errors.New("Error in processing msg:" + fmt.Sprintf("%+v", msg))
			}
		*/
	default:
		return errors.New("Message type unsupported:" + fmt.Sprintf("%+v", msg))
	}
	return nil
}

// processDirBlock validates dir block and save it to factom db.
// similar to blockChain.BC_ProcessBlock
func processDirBlock(msg *wire.MsgDirBlock) error {
	util.Trace()

	blk, _ := db.FetchDBlockByHeight(uint64(msg.DBlk.Header.BlockHeight))
	if blk != nil {
		fmt.Println("DBlock already existing for height:" + string(msg.DBlk.Header.BlockHeight))
		return nil
	}

	dchain.AddDBlockToDChain(msg.DBlk)

	db.ProcessDBlockBatch(msg.DBlk) //?? to be removed later

	fmt.Printf("PROCESSOR: MsgDirBlock=%s\n", spew.Sdump(msg.DBlk))

	msg.DBlk = nil

	return nil
}

// processCBlock validates entry credit block and save it to factom db.
// similar to blockChain.BC_ProcessBlock
func processCBlock(msg *wire.MsgCBlock) error {
	util.Trace()

	//Need to validate against Dchain??

	db.ProcessCBlockBatch(msg.CBlk)

	fmt.Printf("PROCESSOR: MsgCBlock=%s\n", spew.Sdump(msg.CBlk))

	return nil
}

// processEBlock validates entry block and save it to factom db.
// similar to blockChain.BC_ProcessBlock
func processEBlock(msg *wire.MsgEBlock) error {
	util.Trace()
	if msg.EBlk.Header.DBHeight >= dchain.NextBlockHeight || msg.EBlk.Header.DBHeight < 0 {
		return errors.New("MsgEBlock has an invalid DBHeight:" + string(msg.EBlk.Header.DBHeight))
	}

	dblock := dchain.Blocks[msg.EBlk.Header.DBHeight]

	if dblock == nil {
		return errors.New("MsgEBlock has an invalid DBHeight:" + string(msg.EBlk.Header.DBHeight))
	}

	msg.EBlk.BuildMerkleRoot()

	validEblock := false
	for _, dbEntry := range dblock.DBEntries {
		if msg.EBlk.MerkleRoot.IsSameAs(dbEntry.MerkleRoot) && dbEntry.ChainID.IsSameAs(msg.EBlk.Header.ChainID) {
			validEblock = true
			break
		}
	}

	if !validEblock {
		return errors.New("Invalid MsgEBlock with height:" + string(msg.EBlk.Header.EBHeight))
	}

	// create a chain in db if it's not existing
	chain := chainIDMap[msg.EBlk.Header.ChainID.String()]
	if chain == nil {
		chain = new(common.EChain)
		chain.ChainID = msg.EBlk.Header.ChainID

		/******************************
         * TODO
         * 
         * A Chain needs an entry first... Not sure about handling here.
         * 
         ******************************/
		
		db.InsertChain(chain)
		chainIDMap[chain.ChainID.String()] = chain
	}

	db.ProcessEBlockBatch(msg.EBlk)

	fmt.Printf("PROCESSOR: MsgEBlock=%s\n", spew.Sdump(msg.EBlk))

	return nil
}

// processEntry validates entry and save it to factom db.
// similar to blockChain.BC_ProcessBlock
func processEntry(msg *wire.MsgEntry) error {
	util.Trace()

	// store the new entry in db
	entryBinary, _ := msg.Entry.MarshalBinary()
	entryHash := common.Sha(entryBinary)
	db.InsertEntry(entryHash, &entryBinary, msg.Entry, &msg.Entry.ChainID.Bytes)

	fmt.Printf("PROCESSOR: MsgEntry=%s\n", spew.Sdump(msg.Entry))

	return nil
}

/* this should be processed on btcd side
// processFactoidBlock validates factoid block and save it to factom db.
func processFactoidBlock(msg *wire.MsgBlock) error {
	util.Trace()
	fmt.Printf("PROCESSOR: MsgFactoidBlock=%s\n", spew.Sdump(msg))
	return nil
}
*/

// Process a factoid obj message and put it in the process list
func processFactoidTx(msg *wire.MsgInt_FactoidObj) error {

	// Update the credit balance in memory for each EC output
	for k, v := range msg.EntryCredits {
		pubKey := new(common.Hash)
		pubKey.SetBytes(k.Bytes())
		credits := int32(creditsPerFactoid * v / 100000000)
		// Update the credit balance in memory
		balance, _ := eCreditMap[pubKey.String()]
		eCreditMap[pubKey.String()] = balance + credits
	}

	// Add to MyPL if Server Node
	if nodeMode == SERVER_NODE {
		err := plMgr.AddMyProcessListItem(msg, msg.TxSha, wire.ACK_FACTOID_TX)
		if err != nil {
			return err
		}

	}

	return nil
}

// Process a reveal-entry message and put it in the mem pool and the process list
// Put the message in the orphan pool if the message is out of order
func processRevealEntry(msg *wire.MsgRevealEntry) error {

	// Calculate the hash
	entryBinary, _ := msg.Entry.MarshalBinary()
	entryHash := common.Sha(entryBinary)
	shaHash, _ := wire.NewShaHash(entryHash.Bytes)

	chain := chainIDMap[msg.Entry.ChainID.String()]
	if chain == nil {
		fMemPool.addOrphanMsg(msg, shaHash)
		return errors.New("This chain is not supported:" + msg.Entry.ChainID.String())
	}

	// Calculate the required credits
	credits := int32(binary.Size(entryBinary)/1000 + 1)

	// Precalculate the key for prePaidEntryMap
	key := entryHash.String()

	// Delete the entry in the prePaidEntryMap in memory
	prepayment, ok := prePaidEntryMap[key]
	if !ok || prepayment < credits {
		fMemPool.addOrphanMsg(msg, shaHash)
		return errors.New("Credit needs to paid first before an entry is revealed:" + entryHash.String())
	}

	delete(prePaidEntryMap, key) // Only revealed once for multiple prepayments??

	// Add the msg to the Mem pool
	fMemPool.addMsg(msg, shaHash)

	// Add to MyPL if Server Node
	if nodeMode == SERVER_NODE {
		err := plMgr.AddMyProcessListItem(msg, shaHash, wire.ACK_REVEAL_ENTRY)
		if err != nil {
			return err
		}
	}

	return nil
}

// Process a commint-entry message and put it in the mem pool and the process list
// Put the message in the orphan pool if the message is out of order
func processCommitEntry(msg *wire.MsgCommitEntry) error {

	shaHash, _ := msg.Sha()

	// Update the credit balance in memory
	creditBalance, _ := eCreditMap[msg.ECPubKey.String()]
	if creditBalance < int32(msg.Credits) {
		fMemPool.addOrphanMsg(msg, &shaHash)
		return errors.New("Not enough credit for public key:" + msg.ECPubKey.String() + " Balance:" + fmt.Sprint(creditBalance))
	}
	eCreditMap[msg.ECPubKey.String()] = creditBalance - int32(msg.Credits)
	// Update the prePaidEntryMapin memory
	payments, _ := prePaidEntryMap[msg.EntryHash.String()]
	prePaidEntryMap[msg.EntryHash.String()] = payments + int32(msg.Credits)

	// Add to MyPL if Server Node
	if nodeMode == SERVER_NODE {
		err := plMgr.AddMyProcessListItem(msg, &shaHash, wire.ACK_COMMIT_ENTRY)
		if err != nil {
			return err
		}

	}
	return nil
}

func processCommitChain(msg *wire.MsgCommitChain) error {

	shaHash, _ := msg.Sha()

	// Check if the chain id already exists
	_, existing := chainIDMap[msg.ChainID.String()]
	if !existing {
		if msg.ChainID.IsSameAs(dchain.ChainID) || msg.ChainID.IsSameAs(cchain.ChainID) {
			existing = true
		}
	}
	if existing {
		return errors.New("Already existing chain id:" + msg.ChainID.String())
	}

	// Precalculate the key and value pair for prePaidEntryMap
	key := getPrePaidChainKey(msg.EntryHash, msg.ChainID)

	// Update the credit balance in memory
	creditBalance, _ := eCreditMap[msg.ECPubKey.String()]
	if creditBalance < int32(msg.Credits) {
		return errors.New("Insufficient credits for public key:" + msg.ECPubKey.String() + " Balance:" + fmt.Sprint(creditBalance))
	}
	eCreditMap[msg.ECPubKey.String()] = creditBalance - int32(msg.Credits)

	// Update the prePaidEntryMap in memory
	payments, _ := prePaidEntryMap[key]
	prePaidEntryMap[key] = payments + int32(msg.Credits)

	// Add to MyPL if Server Node
	if nodeMode == SERVER_NODE {
		err := plMgr.AddMyProcessListItem(msg, &shaHash, wire.ACK_COMMIT_CHAIN)
		if err != nil {
			return err
		}

	}

	return nil
}

func processBuyEntryCredit(pubKey *common.Hash, credits int32, factoidTxHash *common.Hash) error {

	// Update the credit balance in memory
	balance, _ := eCreditMap[pubKey.String()]
	eCreditMap[pubKey.String()] = balance + credits

	return nil
}

func processRevealChain(msg *wire.MsgRevealChain) error {
	shaHash, _ := msg.Sha()
	newChain := msg.Chain

	// Check if the chain id already exists
	_, existing := chainIDMap[newChain.ChainID.String()]
	if !existing {
		if newChain.ChainID.IsSameAs(dchain.ChainID) || newChain.ChainID.IsSameAs(cchain.ChainID) {
			existing = true
		}
	}
	if existing {
		return errors.New("This chain is already existing:" + newChain.ChainID.String())
	}

	if newChain.FirstEntry == nil {
		return errors.New("The first entry is required to create a new chain:" + newChain.ChainID.String())
	}
	// Calculate the required credits
	binaryChain, _ := newChain.MarshalBinary()
	credits := int32(binary.Size(binaryChain)/1000+1) + creditsPerChain

	// Remove the entry for prePaidEntryMap
	binaryEntry, _ := newChain.FirstEntry.MarshalBinary()
	firstEntryHash := common.Sha(binaryEntry)
	key := getPrePaidChainKey(firstEntryHash, newChain.ChainID)
	prepayment, ok := prePaidEntryMap[key]
	if ok && prepayment >= credits {
		delete(prePaidEntryMap, key)
	} else {
		fMemPool.addOrphanMsg(msg, &shaHash)
		return errors.New("Enough credits need to paid first before creating a new chain:" + newChain.ChainID.String())
	}

	// Add the new chain in the chainIDMap
	chainIDMap[newChain.ChainID.String()] = newChain

	// Add to MyPL if Server Node
	if nodeMode == SERVER_NODE {
		err := plMgr.AddMyProcessListItem(msg, &shaHash, wire.ACK_REVEAL_CHAIN)
		if err != nil {
			return err
		}
	}

	return nil
}

// Process Orphan pool before the end of 10 min
func processFromOrphanPool() error {
	for k, msg := range fMemPool.orphans {
		switch msg.Command() {
		case wire.CmdCommitChain:
			msgCommitChain, _ := msg.(*wire.MsgCommitChain)
			err := processCommitChain(msgCommitChain)
			if err != nil {
				return err
			}
			delete(fMemPool.orphans, k)

		case wire.CmdRevealChain:
			msgRevealChain, _ := msg.(*wire.MsgRevealChain)
			err := processRevealChain(msgRevealChain)
			if err != nil {
				return err
			}
			delete(fMemPool.orphans, k)

		case wire.CmdCommitEntry:
			msgCommitEntry, _ := msg.(*wire.MsgCommitEntry)
			err := processCommitEntry(msgCommitEntry)
			if err != nil {
				return err
			}
			delete(fMemPool.orphans, k)

		case wire.CmdRevealEntry:
			msgRevealEntry, _ := msg.(*wire.MsgRevealEntry)
			err := processRevealEntry(msgRevealEntry)
			if err != nil {
				return err
			}
			delete(fMemPool.orphans, k)
		}
	}
	return nil
}

func buildRevealEntry(msg *wire.MsgRevealEntry) {

	chain := chainIDMap[msg.Entry.ChainID.String()]

	// store the new entry in db
	entryBinary, _ := msg.Entry.MarshalBinary()
	entryHash := common.Sha(entryBinary)
	db.InsertEntry(entryHash, &entryBinary, msg.Entry, &chain.ChainID.Bytes)

	err := chain.NextBlock.AddEBEntry(msg.Entry)

	if err != nil {
		panic("Error while adding Entity to Block:" + err.Error())
	}

}

func buildCommitEntry(msg *wire.MsgCommitEntry) {

	// Create PayEntryCBEntry
	cbEntry := common.NewPayEntryCBEntry(msg.ECPubKey, msg.EntryHash, int32(0-msg.Credits), int64(msg.Timestamp), msg.Sig)

	err := cchain.NextBlock.AddCBEntry(cbEntry)

	if err != nil {
		panic("Error while building Block:" + err.Error())
	}
}

func buildCommitChain(msg *wire.MsgCommitChain) {

	// Create PayChainCBEntry
	cbEntry := common.NewPayChainCBEntry(msg.ECPubKey, msg.EntryHash, int32(0-msg.Credits), msg.ChainID, msg.EntryChainIDHash, msg.Sig)

	err := cchain.NextBlock.AddCBEntry(cbEntry)

	if err != nil {
		panic("Error while building Block:" + err.Error())
	}
}

func buildFactoidObj(msg *wire.MsgInt_FactoidObj) {
	factoidTxHash := new(common.Hash)
	factoidTxHash.SetBytes(msg.TxSha.Bytes())

	for k, v := range msg.EntryCredits {
		pubKey := new(common.Hash)
		pubKey.SetBytes(k.Bytes())
		credits := int32(creditsPerFactoid * v / 100000000)
		cbEntry := common.NewBuyCBEntry(pubKey, factoidTxHash, credits)
		err := cchain.NextBlock.AddCBEntry(cbEntry)
		if err != nil {
			panic(fmt.Sprintf(`Error while adding the First Entry to Block: %s`, err.Error()))
		}
	}
}

func buildRevealChain(msg *wire.MsgRevealChain) {

	newChain := msg.Chain
	// Store the new chain in db
	db.InsertChain(newChain)

	// Chain initialization
	initEChainFromDB(newChain)

	// store the new entry in db
	entryBinary, _ := newChain.FirstEntry.MarshalBinary()
	entryHash := common.Sha(entryBinary)
	db.InsertEntry(entryHash, &entryBinary, newChain.FirstEntry, &newChain.ChainID.Bytes)

	err := newChain.NextBlock.AddEBEntry(newChain.FirstEntry)

	if err != nil {
		panic(fmt.Sprintf(`Error while adding the First Entry to Block: %s`, err.Error()))
	}
}

// Loop through the Process List items and get the touched chains
// Put End-Of-Minute marker in the entry chains
func buildEndOfMinute(pl *consensus.ProcessList, pli *consensus.ProcessListItem) {
	tempChainMap := make(map[string]*common.EChain)
	items := pl.GetPLItems()
	for i := pli.Ack.Index; i >= 0; i-- {
		if wire.END_MINUTE_1 <= items[i].Ack.Type && items[i].Ack.Type <= wire.END_MINUTE_10 {
			break
		} else if items[i].Ack.Type == wire.ACK_REVEAL_ENTRY && tempChainMap[items[i].Ack.ChainID.String()] == nil {

			chain := chainIDMap[items[i].Ack.ChainID.String()]
			chain.NextBlock.AddEndOfMinuteMarker(pli.Ack.Type)
			// Add the new chain in the tempChainMap
			tempChainMap[chain.ChainID.String()] = chain
		}
	}

	// Add it to the entry credit chain
	entries := cchain.NextBlock.CBEntries
	if len(entries) > 0 && entries[len(entries)-1].Type() != common.TYPE_MINUTE_NUMBER {
		cchain.NextBlock.AddEndOfMinuteMarker(pli.Ack.Type)
	}
}

// build Genesis blocks
func buildGenesisBlocks() error {
	util.Trace()

	// Send an End of Minute message to the Factoid component to create a genesis block
	eomMsg := &wire.MsgInt_EOM{
		EOM_Type:         wire.END_MINUTE_10,
		NextDBlockHeight: 0,
	}
	outCtlMsgQueue <- eomMsg

	// Allocate the first two dbentries for ECBlock and Factoid block
	dchain.AddDBEntry(&common.DBEntry{}) // ECBlock
	dchain.AddDBEntry(&common.DBEntry{}) // Factoid block

	util.Trace()
	// Wait for Factoid block to be built and update the DbEntry
	msg := <-doneFBlockQueue
	util.Trace(spew.Sdump(msg))
	doneFBlockMsg, ok := msg.(*wire.MsgInt_FactoidBlock)
	util.Trace(spew.Sdump(doneFBlockMsg))
	//?? to be restored: if ok && doneFBlockMsg.BlockHeight == dchain.NextBlockID {
	// double check MR ??
	if ok {
		util.Trace("ok")
		dbEntryUpdate := new(common.DBEntry)
		dbEntryUpdate.ChainID = fchainID
		dbEntryUpdate.MerkleRoot = doneFBlockMsg.ShaHash.ToFactomHash()

		util.Trace("before dchain")
		dchain.AddFBlockMRToDBEntry(dbEntryUpdate)
		util.Trace("after dchain")
	} else {
		panic("Error in processing msg from doneFBlockQueue:" + fmt.Sprintf("%+v", msg))
	}

	// Entry Credit Chain
	cBlock := newEntryCreditBlock(cchain)
	fmt.Printf("buildGenesisBlocks: cBlock=%s\n", spew.Sdump(cBlock))
	dchain.AddCBlockToDBEntry(cBlock)
	saveCChain(cchain)

	// Directory Block chain
	dbBlock := newDirectoryBlock(dchain)
	// Check block hash if genesis block here??

	saveDChain(dchain)

	// Only Servers can write the anchor to Bitcoin network
	if nodeMode == SERVER_NODE && dbBlock != nil && false { //?? for testing
		//dbInfo := common.NewDBInfoFromDBlock(dbBlock)
		//saveDBMerkleRoottoBTC(dbInfo) //goroutine??
	}

	return nil
}

// build blocks from all process lists
func buildBlocks() error {
	util.Trace()

	// Allocate the first two dbentries for ECBlock and Factoid block
	dchain.AddDBEntry(&common.DBEntry{}) // ECBlock
	dchain.AddDBEntry(&common.DBEntry{}) // Factoid block

	if plMgr != nil && plMgr.MyProcessList.IsValid() {
		buildFromProcessList(plMgr.MyProcessList)
	}

	// Wait for Factoid block to be built and update the DbEntry
	msg := <-doneFBlockQueue
	doneFBlockMsg, ok := msg.(*wire.MsgInt_FactoidBlock)
	//?? to be restored: if ok && doneFBlockMsg.BlockHeight == dchain.NextBlockID {
	if ok {
		dbEntryUpdate := new(common.DBEntry)
		dbEntryUpdate.ChainID = fchainID
		dbEntryUpdate.MerkleRoot = doneFBlockMsg.ShaHash.ToFactomHash()
		dchain.AddFBlockMRToDBEntry(dbEntryUpdate)
	} else {
		panic("Error in processing msg from doneFBlockQueue:" + fmt.Sprintf("%+v", msg))
	}

	// Entry Credit Chain
	cBlock := newEntryCreditBlock(cchain)
	if cBlock != nil { // to be removed??
		dchain.AddCBlockToDBEntry(cBlock)
		saveCChain(cchain)
	}

	// sort the echains by chain id
	var keys []string
	for k := range chainIDMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Entry Chains
	for _, k := range keys {
		chain := chainIDMap[k]
		eblock := newEntryBlock(chain)
		if eblock != nil {
			dchain.AddEBlockToDBEntry(eblock)
		}
		saveEChain(chain)
	}

	// Directory Block chain
	dbBlock := newDirectoryBlock(dchain)
	// Check block hash if genesis block here??

	// Generate the inventory vector and relay it.
	binary, _ := dbBlock.MarshalBinary()
	commonHash := common.Sha(binary)
	hash, _ := wire.NewShaHash(commonHash.Bytes)
	outMsgQueue <- (&wire.MsgInt_DirBlock{hash})

	saveDChain(dchain)

	// re-initialize the process lit manager
	initProcessListMgr()

	// Initialize timer for the new dblock
	if nodeMode == SERVER_NODE {
		timer := &BlockTimer{
			nextDBlockHeight: dchain.NextBlockHeight,
			inCtlMsgQueue:    inCtlMsgQueue,
		}
		go timer.StartBlockTimer()
	}

	// Only Servers can write the anchor to Bitcoin network
	if nodeMode == SERVER_NODE && dbBlock != nil && false { //?? for testing
		// dbInfo := common.NewDBInfoFromDBlock(dbBlock)

		// FIXME
		// TODO
		// anchoring can't be done via this code; the source is no longer Bitcoin-compatible
		//		saveDBMerkleRoottoBTC(dbInfo) //goroutine??
		util.Trace("NOT IMPLEMENTED! IMPORTANT: Anchoring code 2 !!!")

	}

	return nil
}

// build blocks from a process lists
func buildFromProcessList(pl *consensus.ProcessList) error {
	for _, pli := range pl.GetPLItems() {
		if pli.Ack.Type == wire.ACK_COMMIT_CHAIN {
			buildCommitChain(pli.Msg.(*wire.MsgCommitChain))
		} else if pli.Ack.Type == wire.ACK_COMMIT_ENTRY {
			buildCommitEntry(pli.Msg.(*wire.MsgCommitEntry))
		} else if pli.Ack.Type == wire.ACK_REVEAL_CHAIN {
			buildRevealChain(pli.Msg.(*wire.MsgRevealChain))
		} else if pli.Ack.Type == wire.ACK_REVEAL_ENTRY {
			buildRevealEntry(pli.Msg.(*wire.MsgRevealEntry))
		} else if pli.Ack.Type == wire.ACK_FACTOID_TX {
			buildFactoidObj(pli.Msg.(*wire.MsgInt_FactoidObj))
			//Send the notification to Factoid component
			outMsgQueue <- pli.Msg.(*wire.MsgInt_FactoidObj)
		} else if wire.END_MINUTE_1 <= pli.Ack.Type && pli.Ack.Type <= wire.END_MINUTE_10 {
			buildEndOfMinute(pl, pli)
		}
	}

	return nil
}

func newEntryBlock(chain *common.EChain) *common.EBlock {

	// acquire the last block
	block := chain.NextBlock

	if len(block.EBEntries) < 1 {
		//log.Println("No new entry found. No block created for chain: "  + common.EncodeChainID(chain.ChainID))
		return nil
	}

	// Create the block and add a new block for new coming entries

	block.Header.DBHeight = dchain.NextBlockHeight
	block.Header.EntryCount = uint32(len(block.EBEntries))
	block.Header.StartTime = dchain.NextBlock.Header.StartTime

	if devNet {
		block.Header.NetworkID = common.NETWORK_ID_TEST
	} else {
		block.Header.NetworkID = common.NETWORK_ID_EB
	}

	// Create the Entry Block Boday Merkle Root from EB Entries
	hashes := make([]*common.Hash, 0, len(block.EBEntries))
	for _, entry := range block.EBEntries {
		hashes = append(hashes, entry.EntryHash)
	}
	merkle := common.BuildMerkleTreeStore(hashes)
	block.Header.BodyMR = merkle[len(merkle)-1]

	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root
	hashes = make([]*common.Hash, 0, 2)
	binaryEBHeader, _ := block.Header.MarshalBinary()
	hashes = append(hashes, common.Sha(binaryEBHeader))
	hashes = append(hashes, block.Header.BodyMR)
	merkle = common.BuildMerkleTreeStore(hashes)
	block.MerkleRoot = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Entry Block
	fmt.Println("block.MerkleRoot:%v", block.MerkleRoot.String())
	blkhash, _ := common.CreateHash(block)
	block.EBHash = blkhash
	log.Println("blkhash:%v", blkhash.Bytes)

	block.IsSealed = true
	chain.NextBlockHeight++
	chain.NextBlock, _ = common.CreateBlock(chain, block, 10)

	//Store the block in db
	db.ProcessEBlockBatch(block)
	log.Println("EntryBlock: block" + strconv.FormatUint(uint64(block.Header.EBHeight), 10) + " created for chain: " + chain.ChainID.String())
	return block
}

func newEntryCreditBlock(chain *common.CChain) *common.CBlock {

	// acquire the last block
	block := chain.NextBlock

	if chain.NextBlockHeight != dchain.NextBlockHeight {
		panic("Entry Credit Block height does not match Directory Block height:" + string(dchain.NextBlockHeight))
	}

	block.Header.BodyHash, _ = block.BuildCBBodyHash()
	block.Header.EntryCount = uint64(len(block.CBEntries))
	block.Header.BodySize = block.MarshalledSize() - block.Header.MarshalledSize()
	block.BuildCBHash()
	block.BuildMerkleRoot()

	// Create the block and add a new block for new coming entries
	chain.BlockMutex.Lock()
	chain.NextBlockHeight++
	chain.NextBlock, _ = common.CreateCBlock(chain, block, 10)
	chain.BlockMutex.Unlock()

	//Store the block in db
	db.ProcessCBlockBatch(block)
	log.Println("EntryCreditBlock: block" + strconv.FormatUint(uint64(block.Header.DBHeight), 10) + " created for chain: " + chain.ChainID.String())

	return block
}

func newDirectoryBlock(chain *common.DChain) *common.DirectoryBlock {

	// acquire the last block
	block := chain.NextBlock

	if devNet {
		block.Header.NetworkID = common.NETWORK_ID_TEST
	} else {
		block.Header.NetworkID = common.NETWORK_ID_EB
	}

	// Create the block add a new block for new coming entries
	chain.BlockMutex.Lock()
	block.Header.EntryCount = uint32(len(block.DBEntries))
	// Calculate Merkle Root for FBlock and store it in header
	if block.Header.BodyMR == nil {
		block.Header.BodyMR, _ = block.BuildBodyMR()
	}
	block.IsSealed = true
	chain.AddDBlockToDChain(block)
	chain.NextBlockHeight++
	chain.NextBlock, _ = common.CreateDBlock(chain, block, 10)
	chain.BlockMutex.Unlock()

	block.DBHash, _ = common.CreateHash(block)
	//Store the block in db
	db.ProcessDBlockBatch(block)

	log.Println("DirectoryBlock: block" + strconv.FormatUint(uint64(block.Header.BlockHeight), 10) + " created for directory block chain: " + chain.ChainID.String())

	return block
}

func GetEntryCreditBalance(pubKey *common.Hash) (int32, error) {

	return eCreditMap[pubKey.String()], nil
}

// Validate dir chain from genesis block
func validateDChain(c *common.DChain) error {

	if uint32(len(c.Blocks)) != c.NextBlockHeight {
		return errors.New("Dir chain doesn't have an expected Next Block ID: " + string(c.NextBlockHeight))
	}

	//prevBlk := c.Blocks[0]
	prevMR, prevBlkHash, err := validateDBlock(c, c.Blocks[0])
	if err != nil {
		return err
	}

	//validate the genesis block here??

	for i := 1; i < len(c.Blocks); i++ {
		if !prevBlkHash.IsSameAs(c.Blocks[i].Header.PrevBlockHash) {
			return errors.New("Previous block hash not matching for Dir block: " + string(i))
		}
		if !prevMR.IsSameAs(c.Blocks[i].Header.BodyMR) { //??

		}
		mr, dblkHash, err := validateDBlock(c, c.Blocks[i])
		if err != nil {
			return err
		}

		prevMR = mr
		prevBlkHash = dblkHash
		//prevBlk = c.Blocks[i]
	}

	return nil
}

func validateDBlock(c *common.DChain, b *common.DirectoryBlock) (merkleRoot *common.Hash, dbHash *common.Hash, err error) {

	merkleRoot, err = b.BuildBodyMR()
	if err != nil {
		return nil, nil, err
	}

	for _, dbEntry := range b.DBEntries {
		switch dbEntry.ChainID.String() {
		case cchain.ChainID.String():
			err := validateCBlockByMR(dbEntry.MerkleRoot)
			if err != nil {
				return nil, nil, err
			}

		case fchainID.String():
			err := validateFBlockByMR(dbEntry.MerkleRoot)
			if err != nil {
				return nil, nil, err
			}

		default:
			err := validateEBlockByMR(dbEntry.ChainID, dbEntry.MerkleRoot)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	dbBinary, _ := b.MarshalBinary()
	dbHash = common.Sha(dbBinary)

	return merkleRoot, dbHash, nil
}

func validateFBlockByMR(mr *common.Hash) error {
	// Call BTCD side for factoid block validation??

	return nil
}

func validateCBlockByMR(mr *common.Hash) error {
	cb, _ := db.FetchCBlockByHash(mr)

	if cb == nil {
		return errors.New("Entry block not found in db for merkle root: " + mr.String())
	}

	return nil
}

func validateEBlockByMR(cid *common.Hash, mr *common.Hash) error {

	eb, _ := db.FetchEBlockByMR(mr)

	if eb == nil {
		return errors.New("Entry block not found in db for merkle root: " + mr.String())
	}

	eb.BuildMerkleRoot()

	if !mr.IsSameAs(eb.MerkleRoot) {
		return errors.New("Entry block's merkle root does not match with: " + mr.String())
	}

	for _, ebEntry := range eb.EBEntries {
		entry, _ := db.FetchEntryByHash(ebEntry.EntryHash)
		if entry == nil {
			return errors.New("Entry not found in db for entry hash: " + ebEntry.EntryHash.String())
		}
	}

	return nil
}

func saveDChain(chain *common.DChain) {
	if len(chain.Blocks) == 0 {
		//log.Println("no blocks to save for chain: " + string (*chain.ChainID))
		return
	}

	bcp := make([]*common.DirectoryBlock, len(chain.Blocks))

	chain.BlockMutex.Lock()
	copy(bcp, chain.Blocks)
	chain.BlockMutex.Unlock()

	for i, block := range bcp {
		//the open block is not saved
		if block == nil || block.IsSealed == false {
			continue
		}

		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
		if fileNotExists(dataStorePath + strChainID) {
			err := os.MkdirAll(dataStorePath+strChainID, 0777)
			if err == nil {
				log.Println("Created directory " + dataStorePath + strChainID)
			} else {
				log.Println(err)
			}
		}
		err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", i), data, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func saveEChain(chain *common.EChain) {

	eBlocks, _ := db.FetchAllEBlocksByChain(chain.ChainID)
	sort.Sort(util.ByEBlockIDAccending(*eBlocks))

	for i, block := range *eBlocks {

		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
		if fileNotExists(dataStorePath + strChainID) {
			err := os.MkdirAll(dataStorePath+strChainID, 0777)
			if err == nil {
				log.Println("Created directory " + dataStorePath + strChainID)
			} else {
				log.Println(err)
			}
		}

		err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", i), data, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func saveCChain(chain *common.CChain) {

	// get all cBlocks from db
	cBlocks, _ := db.FetchAllCBlocks()
	sort.Sort(util.ByCBlockIDAccending(cBlocks))

	for i, block := range cBlocks {

		data, err := block.MarshalBinary()
		if err != nil {
			panic(err)
		}

		strChainID := chain.ChainID.String()
		if fileNotExists(dataStorePath + strChainID) {
			err := os.MkdirAll(dataStorePath+strChainID, 0777)
			if err == nil {
				log.Println("Created directory " + dataStorePath + strChainID)
			} else {
				log.Println(err)
			}
		}
		err = ioutil.WriteFile(fmt.Sprintf(dataStorePath+strChainID+"/store.%09d.block", i), data, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func initDChain() {
	dchain = new(common.DChain)

	//Initialize dbInfoMap
	dbInfoMap = make(map[string]*common.DBInfo)

	//Initialize the Directory Block Chain ID
	dchain.ChainID = new(common.Hash)
	barray := common.D_CHAINID
	dchain.ChainID.SetBytes(barray)

	// get all dBlocks from db
	dBlocks, _ := db.FetchAllDBlocks()
	sort.Sort(util.ByDBlockIDAccending(dBlocks))

	//fmt.Printf("initDChain: dBlocks=%s\n", spew.Sdump(dBlocks))

	dchain.Blocks = make([]*common.DirectoryBlock, len(dBlocks), len(dBlocks)+1)

	for i := 0; i < len(dBlocks); i = i + 1 {
		if dBlocks[i].Header.BlockHeight != uint32(i) {
			panic("Error in initializing dChain:" + dchain.ChainID.String())
		}
		dBlocks[i].Chain = dchain
		dBlocks[i].IsSealed = true
		dBlocks[i].IsSavedInDB = true
		dchain.Blocks[i] = &dBlocks[i]
	}

	// double check the block ids
	for i := 0; i < len(dchain.Blocks); i = i + 1 {
		if uint32(i) != dchain.Blocks[i].Header.BlockHeight {
			panic(errors.New("BlockID does not equal index for chain:" + dchain.ChainID.String() + " block:" + fmt.Sprintf("%v", dchain.Blocks[i].Header.BlockHeight)))
		}
	}

	//Create an empty block and append to the chain
	if len(dchain.Blocks) == 0 {
		dchain.NextBlockHeight = 0
		dchain.NextBlock, _ = common.CreateDBlock(dchain, nil, 10)
		//buildGenesisBlocks() // empty genesis block??
	} else {
		dchain.NextBlockHeight = uint32(len(dchain.Blocks))
		dchain.NextBlock, _ = common.CreateDBlock(dchain, dchain.Blocks[len(dchain.Blocks)-1], 10)
	}

	// only for debug??
	saveDChain(dchain)

	//Double check the sealed flag
	if dchain.NextBlock.IsSealed == true {
		panic("dchain.Blocks[dchain.NextBlockID].IsSealed for chain:" + dchain.ChainID.String())
	}

}

func initCChain() {

	eCreditMap = make(map[string]int32)
	prePaidEntryMap = make(map[string]int32)

	//Initialize the Entry Credit Chain ID
	cchain = new(common.CChain)
	barray := common.EC_CHAINID
	cchain.ChainID = new(common.Hash)
	cchain.ChainID.SetBytes(barray)

	// get all cBlocks from db
	cBlocks, _ := db.FetchAllCBlocks()
	sort.Sort(util.ByCBlockIDAccending(cBlocks))

	//fmt.Printf("initCChain: cBlocks=%s\n", spew.Sdump(cBlocks))

	for i := 0; i < len(cBlocks); i = i + 1 {
		if cBlocks[i].Header.DBHeight != uint32(i) {
			panic("Error in initializing dChain:" + cchain.ChainID.String())
		}

		// Calculate the EC balance for each account
		initializeECreditMap(&cBlocks[i])
	}

	// double check the block ids
	for i := 0; i < len(cBlocks); i = i + 1 {
		if uint32(i) != cBlocks[i].Header.DBHeight {
			panic(errors.New("BlockID does not equal index for chain:" + cchain.ChainID.String() + " block:" + fmt.Sprintf("%v", cBlocks[i].Header.DBHeight)))
		}
	}

	//Create an empty block and append to the chain
	if len(cBlocks) == 0 || dchain.NextBlockHeight == 0 {
		cchain.NextBlockHeight = 0
		cchain.NextBlock, _ = common.CreateCBlock(cchain, nil, 10)

	} else {
		// Entry Credit Chain should have the same height as the dir chain
		cchain.NextBlockHeight = dchain.NextBlockHeight
		cchain.NextBlock, _ = common.CreateCBlock(cchain, &cBlocks[cchain.NextBlockHeight-1], 10)
	}

	// create a backup copy before processing entries
	copyCreditMap(eCreditMap, eCreditMapBackup)

	//ONly for debug??
	saveCChain(cchain)

	// ONly for debug??
	//printCChain()
	printCreditMap()
	printPaidEntryMap()

}

func initEChains() {

	chainIDMap = make(map[string]*common.EChain)

    /******************************
     * 
     * TODO:  Make our system into a lazy evaluation system.  DON'T assume
     *        data is in memory, or even in the database!  We may have to 
     *        pull from the network!
     * 
     * This should not be.  Any time we do not have a chain that we need,
     * we can pull it by its chainID from the database.  If it does not 
     * exist, then it has not been created yet.
     * 
     * I see no reason to pull all chains into memory by default.  
     * 
     ******************************/
    
}

func initializeECreditMap(block *common.CBlock) {
	for _, cbEntry := range block.CBEntries {
		// Only process: TYPE_PAY_CHAIN, TYPE_PAY_ENTRY, TYPE_BUY
		if cbEntry.Type() >= common.TYPE_PAY_CHAIN {
			credits, _ := eCreditMap[cbEntry.PublicKey().String()]
			eCreditMap[cbEntry.PublicKey().String()] = credits + cbEntry.Credits()
		}
	}
}
func initProcessListMgr() {
	plMgr = consensus.NewProcessListMgr(dchain.NextBlockHeight, 1, 10)

}

func getPrePaidChainKey(entryHash *common.Hash, chainIDHash *common.Hash) string {
	return chainIDHash.String() + entryHash.String()
}

func copyCreditMap(originalMap map[string]int32, newMap map[string]int32) {

	// clean up the new map
	if newMap != nil {
		for k, _ := range newMap {
			delete(newMap, k)
		}
	} else {
		newMap = make(map[string]int32)
	}

	// copy every element from the original map
	for k, v := range originalMap {
		newMap[k] = v
	}

}

func printCreditMap() {

	fmt.Println("eCreditMap:")
	for key := range eCreditMap {
		fmt.Println("Key:", key, "Value", eCreditMap[key])
	}
}

func printPaidEntryMap() {

	fmt.Println("prePaidEntryMap:")
	for key := range prePaidEntryMap {
		fmt.Println("Key:", key, "Value", prePaidEntryMap[key])
	}
}

/*
func printCChain() {

	fmt.Println("cchain:", cchain.ChainID.String())

	for i, block := range cchain.Blocks {
		if !block.IsSealed {
			continue
		}
		var buf bytes.Buffer
		err := factomapi.SafeMarshal(&buf, block.Header)

		fmt.Println("block.Header", string(i), ":", string(buf.Bytes()))

		for _, cbentry := range block.CBEntries {
			t := reflect.TypeOf(cbentry)
			fmt.Println("cbEntry Type:", t.Name(), t.String())
			if strings.Contains(t.String(), "PayChainCBEntry") {
				fmt.Println("PayChainCBEntry - pubkey:", cbentry.PublicKey().String(), " Credits:", cbentry.Credits())
				var buf bytes.Buffer
				err := factomapi.SafeMarshal(&buf, cbentry)
				if err != nil {
					fmt.Println("Error:%v", err)
				}

				fmt.Println("PayChainCBEntry JSON", ":", string(buf.Bytes()))

			} else if strings.Contains(t.String(), "PayEntryCBEntry") {
				fmt.Println("PayEntryCBEntry - pubkey:", cbentry.PublicKey().String(), " Credits:", cbentry.Credits())
				var buf bytes.Buffer
				err := factomapi.SafeMarshal(&buf, cbentry)
				if err != nil {
					fmt.Println("Error:%v", err)
				}

				fmt.Println("PayEntryCBEntry JSON", ":", string(buf.Bytes()))

			} else if strings.Contains(t.String(), "BuyCBEntry") {
				fmt.Println("BuyCBEntry - pubkey:", cbentry.PublicKey().String(), " Credits:", cbentry.Credits())
				var buf bytes.Buffer
				err := factomapi.SafeMarshal(&buf, cbentry)
				if err != nil {
					fmt.Println("Error:%v", err)
				}
				fmt.Println("BuyCBEntry JSON", ":", string(buf.Bytes()))
			}
		}

		if err != nil {

			fmt.Println("Error:%v", err)
		}
	}

}
*/
