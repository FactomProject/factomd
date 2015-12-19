// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// factomlog is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package anchor

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"github.com/btcsuitereleases/btcd/btcjson"
	"github.com/btcsuitereleases/btcd/chaincfg"
	"github.com/btcsuitereleases/btcd/txscript"
	"github.com/btcsuitereleases/btcd/wire"
	"github.com/btcsuitereleases/btcrpcclient"
	"github.com/btcsuitereleases/btcutil"
	"github.com/davecgh/go-spew/spew"

	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/util"
)

var (
	balances         []balance // unspent balance & address & its WIF
	cfg              *util.FactomdConfig
	dclient, wclient *btcrpcclient.Client
	dirBlockInfoSlice  []*dbInfo.DirBlockInfo
	db               *databaseOverlay.Overlay
	walletLocked     = true
	reAnchorAfter    = 4 // hours. For anchors that do not get bitcoin callback info for over 10 hours, then re-anchor them.
	tenMinutes       = 10 // 10 minute mark
	defaultAddress   btcutil.Address
	minBalance       btcutil.Amount	// 0.01 btc
	fee                 btcutil.Amount // tx fee for written into btc
	confirmationsNeeded int

	serverPrivKey primitives.PrivateKey		//Server Private key for milestone 1
	serverECKey primitives.PrivateKey		//Server Entry Credit private key
	anchorChainID interfaces.IHash
	state interfaces.IState
)

type balance struct {
	unspentResult btcjson.ListUnspentResult
	address       btcutil.Address
	wif           *btcutil.WIF
}

// SendRawTransactionToBTC is the main function used to anchor factom
// dir block hash to bitcoin blockchain
func SendRawTransactionToBTC(hash interfaces.IHash, blockHeight uint32) (*wire.ShaHash, error) {
	anchorLog.Debug("SendRawTransactionToBTC: hash=", hash.String(), ", dir block height=", blockHeight) //strconv.FormatUint(blockHeight, 10))
	dirBlockInfo, err := sanityCheck(hash)
	if err != nil {
		return nil, err
	}
	return doTransaction(hash, blockHeight, dirBlockInfo)
}

func doTransaction(hash interfaces.IHash, blockHeight uint32, dirBlockInfo *dbInfo.DirBlockInfo) (*wire.ShaHash, error) {
	b := balances[0]
	balances = balances[1:]
	anchorLog.Info("new balances.len=", len(balances))

	msgtx, err := createRawTransaction(b, hash.Bytes(), blockHeight)
	if err != nil {
		return nil, fmt.Errorf("cannot create Raw Transaction: %s", err)
	}

	shaHash, err := sendRawTransaction(msgtx)
	if err != nil {
		return nil, fmt.Errorf("cannot send Raw Transaction: %s", err)
	}

	if dirBlockInfo != nil {
		dirBlockInfo.BTCTxHash = toHash(shaHash)
		dirBlockInfo.Timestamp = time.Now().Unix()
		db.SaveDirBlockInfo(dirBlockInfo)
	}

	return shaHash, nil
}

func sanityCheck(hash interfaces.IHash) (*dbInfo.DirBlockInfo, error) {
	index := sort.Search(len(dirBlockInfoSlice), func(i int) bool {
		return dirBlockInfoSlice[i].DBMerkleRoot.IsSameAs(hash)
		})
	var dirBlockInfo *dbInfo.DirBlockInfo
	if index < len(dirBlockInfoSlice) {
		dirBlockInfo = dirBlockInfoSlice[index]
	}
	if dirBlockInfo == nil {
		s := fmt.Sprintf("Anchor Error: hash %s does not exist in dirBlockInfoSlice.\n", hash.String())
		anchorLog.Error(s)
		return nil, errors.New(s)
	}
	if dirBlockInfo.BTCConfirmed {
		s := fmt.Sprintf("Anchor Warning: hash %s has already been confirmed in btc block chain.\n", hash.String())
		anchorLog.Error(s)
		return nil, errors.New(s)
	}
	if dclient == nil || wclient == nil {
		s := fmt.Sprintf("\n\n$$$ WARNING: rpc clients and/or wallet are not initiated successfully. No anchoring for now.\n")
		anchorLog.Warning(s)
		return nil, errors.New(s)
	}
	if len(balances) == 0 {
		anchorLog.Warning("len(balances) == 0, start rescan UTXO *** ")
		updateUTXO(minBalance)
	}
	if len(balances) == 0 {
		anchorLog.Warning("len(balances) == 0, start rescan UTXO *** ")
		updateUTXO(fee)
	}
	if len(balances) == 0 {
		s := fmt.Sprintf("\n\n$$$ WARNING: No balance in your wallet. No anchoring for now.\n")
		anchorLog.Warning(s)
		return nil, errors.New(s)
	}
	return dirBlockInfo, nil
}

func createRawTransaction(b balance, hash []byte, blockHeight uint32) (*wire.MsgTx, error) {
	msgtx := wire.NewMsgTx()

	if err := addTxOuts(msgtx, b, hash, blockHeight); err != nil {
		return nil, fmt.Errorf("cannot addTxOuts: %s", err)
	}

	if err := addTxIn(msgtx, b); err != nil {
		return nil, fmt.Errorf("cannot addTxIn: %s", err)
	}

	if err := validateMsgTx(msgtx, []btcjson.ListUnspentResult{b.unspentResult}); err != nil {
		return nil, fmt.Errorf("cannot validateMsgTx: %s", err)
	}

	return msgtx, nil
}

func addTxIn(msgtx *wire.MsgTx, b balance) error {
	output := b.unspentResult
	//anchorLog.Infof("unspentResult: %s\n", spew.Sdump(output))
	prevTxHash, err := wire.NewShaHashFromStr(output.TxID)
	if err != nil {
		return fmt.Errorf("cannot get sha hash from str: %s", err)
	}
	if prevTxHash == nil {
		anchorLog.Error("prevTxHash == nil")
	}

	outPoint := wire.NewOutPoint(prevTxHash, output.Vout)
	msgtx.AddTxIn(wire.NewTxIn(outPoint, nil))
	if outPoint == nil {
		anchorLog.Error("outPoint == nil")
	}

	// OnRedeemingTx
	err = dclient.NotifySpent([]*wire.OutPoint{outPoint})
	if err != nil {
		anchorLog.Error("NotifySpent err: ", err.Error())
	}

	subscript, err := hex.DecodeString(output.ScriptPubKey)
	if err != nil {
		return fmt.Errorf("cannot decode scriptPubKey: %s", err)
	}
	if subscript == nil {
		anchorLog.Error("subscript == nil")
	}

	sigScript, err := txscript.SignatureScript(msgtx, 0, subscript, txscript.SigHashAll, b.wif.PrivKey, true)
	if err != nil {
		return fmt.Errorf("cannot create scriptSig: %s", err)
	}
	if sigScript == nil {
		anchorLog.Error("sigScript == nil")
	}

	msgtx.TxIn[0].SignatureScript = sigScript
	return nil
}

func addTxOuts(msgtx *wire.MsgTx, b balance, hash []byte, blockHeight uint32) error {
	anchorHash, err := prependBlockHeight(blockHeight, hash)
	if err != nil {
		anchorLog.Errorf("ScriptBuilder error: %v\n", err)
	}

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	builder.AddData(anchorHash)

	// latest routine from Conformal btcsuite returns 2 parameters, not 1... not sure what to do for people with the old conformal libraries :(
	opReturn, err := builder.Script()
	msgtx.AddTxOut(wire.NewTxOut(0, opReturn))
	if err != nil {
		anchorLog.Errorf("ScriptBuilder error: %v\n", err)
	}

	amount, _ := btcutil.NewAmount(b.unspentResult.Amount)
	change := amount - fee

	// Check if there are leftover unspent outputs, and return coins back to
	// a new address we own.
	if change > 0 {
		// Spend change.
		pkScript, err := txscript.PayToAddrScript(b.address)
		if err != nil {
			return fmt.Errorf("cannot create txout script: %s", err)
		}
		msgtx.AddTxOut(wire.NewTxOut(int64(change), pkScript))
	}
	return nil
}

func validateMsgTx(msgtx *wire.MsgTx, inputs []btcjson.ListUnspentResult) error {
	flags := txscript.ScriptBip16 | txscript.ScriptStrictMultiSig //ScriptCanonicalSignatures
	bip16 := time.Now().After(txscript.Bip16Activation)
	if bip16 {
		flags |= txscript.ScriptBip16
	}

	for i := range msgtx.TxIn {
		scriptPubKey, err := hex.DecodeString(inputs[i].ScriptPubKey)
		if err != nil {
			return fmt.Errorf("cannot decode scriptPubKey: %s", err)
		}
		engine, err := txscript.NewEngine(scriptPubKey, msgtx, i, flags)
		//engine, err := txscript.NewEngine(scriptPubKey, msgtx, i, flags, nil)
		if err != nil {
			anchorLog.Errorf("cannot create script engine: %s\n", err)
			return fmt.Errorf("cannot create script engine: %s", err)
		}
		if err = engine.Execute(); err != nil {
			anchorLog.Errorf("cannot execute script engine: %s\n  === UnspentResult: %s", err, spew.Sdump(inputs[i]))
			return fmt.Errorf("cannot execute script engine: %s", err)
		}
	}
	return nil
}

func sendRawTransaction(msgtx *wire.MsgTx) (*wire.ShaHash, error) {
	//anchorLog.Debug("sendRawTransaction: msgTx=", spew.Sdump(msgtx))
	buf := bytes.Buffer{}
	buf.Grow(msgtx.SerializeSize())
	if err := msgtx.BtcEncode(&buf, wire.ProtocolVersion); err != nil {
		return nil, err
	}

	// use rpc client for btcd here for better callback info
	// this should not require wallet to be unlocked
	shaHash, err := dclient.SendRawTransaction(msgtx, false)
	if err != nil {
		return nil, fmt.Errorf("failed in rpcclient.SendRawTransaction: %s", err)
	}
	anchorLog.Info("btc txHash returned: ", shaHash) // new tx hash
	return shaHash, nil
}

func createBtcwalletNotificationHandlers() btcrpcclient.NotificationHandlers {
	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnWalletLockState: func(locked bool) {
			anchorLog.Info("wclient: OnWalletLockState, locked=", locked)
			walletLocked = locked
		},
	}
	return ntfnHandlers
}

func createBtcdNotificationHandlers() btcrpcclient.NotificationHandlers {
	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnRedeemingTx: func(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
			if details != nil {
				// do not block OnRedeemingTx callback
				//anchorLog.Info(" saveDirBlockInfo.")
				go saveDirBlockInfo(transaction, details)
			}
		},
	}
	return ntfnHandlers
}

func checkMissingDirBlockInfo() {
	anchorLog.Debug("checkMissingDirBlockInfo for those unsaved DirBlocks in database")
	dblocks, _ := db.FetchAllDBlocks()
	dirBlockInfos, _ := db.FetchAllDirBlockInfos()		//FetchAllDirBlockInfos()
	for _, dblock := range dblocks {
		var found = false
		for i, dbinfo := range dirBlockInfos {
			if dbinfo.GetDatabaseHeight() == dblock.GetDatabaseHeight() {
				dirBlockInfos = append(dirBlockInfos[:i], dirBlockInfos[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			if dblock.GetKeyMR() == nil || bytes.Compare(dblock.GetKeyMR().Bytes(), primitives.NewZeroHash().Bytes()) == 0 {
				dblock.BuildKeyMerkleRoot()
			}
			dirBlockInfo := dbInfo.NewDirBlockInfoFromDirBlock(dblock)
			dirBlockInfo.Timestamp = time.Now().Unix()
			anchorLog.Debug("add missing dirBlockInfo to map: ", spew.Sdump(dirBlockInfo))
			db.SaveDirBlockInfo(dirBlockInfo)
			dirBlockInfoSlice = append(dirBlockInfoSlice, dirBlockInfo)
		}
	}
}

// InitAnchor inits rpc clients for factom
// and load up unconfirmed DirBlockInfo from leveldb
func InitAnchor(s interfaces.IState) {
	anchorLog.Debug("InitAnchor")
	state = s
	db = s.GetDB().(*databaseOverlay.Overlay)
	minBalance, _ = btcutil.NewAmount(0.01)

	var err error
	dirBlockInfoSlice, err = db.FetchAllUnconfirmedDirBlockInfos()
	if err != nil {
		anchorLog.Error("InitAnchor error - " + err.Error())
		return
	}
	anchorLog.Debug("init dirBlockInfoSlice.len=", len(dirBlockInfoSlice))
	// this might take a while to check missing DirBlockInfo for existing DirBlocks in database
	go checkMissingDirBlockInfo()

	readConfig()
	if err = InitRPCClient(); err != nil {
		anchorLog.Error(err.Error())
	} else {
		updateUTXO(minBalance)
	}

	ticker0 := time.NewTicker(time.Minute * time.Duration(1))
	go func() {
		for _ = range ticker0.C {
			checkForAnchor()
		}
	}()

	ticker := time.NewTicker(time.Minute * time.Duration(tenMinutes))
	go func() {
		for _ = range ticker.C {
			anchorLog.Info("In 10 minutes ticker...")
			readConfig()
			if dclient == nil || wclient == nil {
				if err = InitRPCClient(); err != nil {
					anchorLog.Error(err.Error())
				}
			}
			if wclient != nil {
				checkTxConfirmations()
			}
		}
	}()
}

func readConfig() {
	anchorLog.Info("readConfig")
	cfg = util.ReadConfig("")
	confirmationsNeeded = cfg.Anchor.ConfirmationsNeeded
	fee, _ = btcutil.NewAmount(cfg.Btc.BtcTransFee)

	var err error
	serverPrivKey, err = primitives.NewPrivateKeyFromHex(cfg.App.LocalServerPrivKey)
	if err != nil {
		panic("Cannot parse Server Private Key from configuration file: " + err.Error())
	}
	serverECKey, err = primitives.NewPrivateKeyFromHex(cfg.Anchor.ServerECPrivKey)
	if err != nil {
		panic("Cannot parse Server EC Key from configuration file: " + err.Error())
	}
	anchorChainID, err = primitives.HexToHash(cfg.Anchor.AnchorChainID)
	anchorLog.Debug("anchorChainID: ", anchorChainID)
	if err != nil || anchorChainID == nil {
		panic("Cannot parse Server AnchorChainID from configuration file: " + err.Error())
	}
}

// InitRPCClient is used to create rpc client for btcd and btcwallet
// and it can be used to test connecting to btcd / btcwallet servers
// running in different machine.
func InitRPCClient() error {
	anchorLog.Debug("init RPC client")
	if cfg == nil {
		readConfig()
	}
	certHomePath := cfg.Btc.CertHomePath
	rpcClientHost := cfg.Btc.RpcClientHost
	rpcClientEndpoint := cfg.Btc.RpcClientEndpoint
	rpcClientUser := cfg.Btc.RpcClientUser
	rpcClientPass := cfg.Btc.RpcClientPass
	certHomePathBtcd := cfg.Btc.CertHomePathBtcd
	rpcBtcdHost := cfg.Btc.RpcBtcdHost

	// Connect to local btcwallet RPC server using websockets.
	ntfnHandlers := createBtcwalletNotificationHandlers()
	certHomeDir := btcutil.AppDataDir(certHomePath, false)
	anchorLog.Debug("btcwallet.cert.home=", certHomeDir)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return fmt.Errorf("cannot read rpc.cert file: %s\n", err)
	}
	connCfg := &btcrpcclient.ConnConfig{
		Host:         rpcClientHost,
		Endpoint:     rpcClientEndpoint,
		User:         rpcClientUser,
		Pass:         rpcClientPass,
		Certificates: certs,
	}
	wclient, err = btcrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		return fmt.Errorf("cannot create rpc client for btcwallet: %s\n", err)
	}
	anchorLog.Debug("successfully created rpc client for btcwallet")

	// Connect to local btcd RPC server using websockets.
	dntfnHandlers := createBtcdNotificationHandlers()
	certHomeDir = btcutil.AppDataDir(certHomePathBtcd, false)
	anchorLog.Debug("btcd.cert.home=", certHomeDir)
	certs, err = ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return fmt.Errorf("cannot read rpc.cert file for btcd rpc server: %s\n", err)
	}
	dconnCfg := &btcrpcclient.ConnConfig{
		Host:         rpcBtcdHost,
		Endpoint:     rpcClientEndpoint,
		User:         rpcClientUser,
		Pass:         rpcClientPass,
		Certificates: certs,
	}
	dclient, err = btcrpcclient.New(dconnCfg, &dntfnHandlers)
	if err != nil {
		return fmt.Errorf("cannot create rpc client for btcd: %s\n", err)
	}
	anchorLog.Debug("successfully created rpc client for btcd")

	return nil
}

func unlockWallet(timeoutSecs int64) error {
	err := wclient.WalletPassphrase(cfg.Btc.WalletPassphrase, int64(timeoutSecs))
	if err != nil {
		return fmt.Errorf("cannot unlock wallet with passphrase: %s", err)
	}
	walletLocked = false
	return nil
}

// ByAmount defines the methods needed to satisify sort.Interface to
// sort a slice of UTXOs by their amount.
type ByAmount []balance

func (u ByAmount) Len() int           { return len(u) }
func (u ByAmount) Less(i, j int) bool { return u[i].unspentResult.Amount < u[j].unspentResult.Amount }
func (u ByAmount) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

func updateUTXO(base btcutil.Amount) error {
	anchorLog.Info("updateUTXO: base=", base.ToBTC())
	if wclient == nil {
		anchorLog.Info("updateUTXO: wclient is nil")
		return nil
	}
	balances = make([]balance, 0, 200)
	err := unlockWallet(int64(6)) //600
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	unspentResults, err := wclient.ListUnspentMin(confirmationsNeeded) //minConf=1
	if err != nil {
		return fmt.Errorf("cannot list unspent. %s", err)
	}
	anchorLog.Info("updateUTXO: unspentResults.len=", len(unspentResults))

	if len(unspentResults) > 0 {
		var i int
		for _, b := range unspentResults {
			if b.Amount > base.ToBTC() { //fee.ToBTC()
				balances = append(balances, balance{unspentResult: b})
				i++
			}
		}
	}
	anchorLog.Info("updateUTXO: balances.len=", len(balances))

	// Sort eligible balances so that we first pick the ones with highest one
	sort.Sort(sort.Reverse(ByAmount(balances)))

	for i, b := range balances {
		addr, err := btcutil.DecodeAddress(b.unspentResult.Address, &chaincfg.TestNet3Params)
		if err != nil {
			return fmt.Errorf("cannot decode address: %s", err)
		}
		balances[i].address = addr

		wif, err := wclient.DumpPrivKey(addr)
		if err != nil {
			return fmt.Errorf("cannot get WIF: %s", err)
		}
		balances[i].wif = wif
		//anchorLog.Infof("balance[%d]=%s \n", i, spew.Sdump(balances[i]))
	}

	if len(balances) > 0 {
		defaultAddress = balances[0].address
	}
	return nil
}

func prependBlockHeight(height uint32, hash []byte) ([]byte, error) {
	// dir block genesis block height starts with 0, for now
	// similar to bitcoin genesis block
	h := uint64(height)
	if 0xFFFFFFFFFFFF&h != h {
		return nil, errors.New("bad block height")
	}
	header := []byte{'F', 'a'}
	big := make([]byte, 8)
	binary.BigEndian.PutUint64(big, h) //height)
	newdata := append(big[2:8], hash...)
	newdata = append(header, newdata...)
	return newdata, nil
}

func saveDirBlockInfo(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
	anchorLog.Debug("in saveDirBlockInfo")
	var saved = false
	for _, dirBlockInfo := range dirBlockInfoSlice {
		if bytes.Compare(dirBlockInfo.BTCTxHash.Bytes(), transaction.Sha().Bytes()) == 0 {
			doSaveDirBlockInfo(transaction, details, dirBlockInfo, false)
			saved = true
			break
		}
	}
	// This happends when there's a double spending or tx malleated(for dir block 122 and its btc tx)
	// Original: https://www.blocktrail.com/BTC/tx/ac82f4173259494b22f4987f1e18608f38f1ff756fb4a3c637dfb5565aa5e6cf
	// malleated: https://www.blocktrail.com/BTC/tx/a9b2d6b5d320c7f0f384a49b167524aca9c412af36ed7b15ca7ea392bccb2538
	// re-anchored: https://www.blocktrail.com/BTC/tx/ac82f4173259494b22f4987f1e18608f38f1ff756fb4a3c637dfb5565aa5e6cf
	// In this case, if tx malleation is detected, then use the malleated tx to replace the original tx;
	// Otherwise, it will end up being re-anchored.
	if !saved {
		anchorLog.Infof("Not saved to db, (maybe btc tx malleated): btc.tx=%s\n blockDetails=%s\n", spew.Sdump(transaction), spew.Sdump(details))
		checkTxMalleation(transaction, details)
	}
}

func doSaveDirBlockInfo(transaction *btcutil.Tx, details *btcjson.BlockDetails, dirBlockInfo *dbInfo.DirBlockInfo, replace bool) {
	if replace {
		dirBlockInfo.BTCTxHash = toHash(transaction.Sha()) // in case of tx being malleated
	}
	dirBlockInfo.BTCTxOffset = int32(details.Index)
	dirBlockInfo.BTCBlockHeight = details.Height
	btcBlockHash, _ := wire.NewShaHashFromStr(details.Hash)
	dirBlockInfo.BTCBlockHash = toHash(btcBlockHash)
	dirBlockInfo.Timestamp = time.Now().Unix()
	db.SaveDirBlockInfo(dirBlockInfo)
	anchorLog.Infof("In doSaveDirBlockInfo, dirBlockInfo:%s saved to db\n", spew.Sdump(dirBlockInfo))

	// to make factom / explorer more user friendly, instead of waiting for
	// over 2 hours to know it's anchored, we can create the anchor chain instantly
	// then change it when the btc main chain re-org happens.
	saveToAnchorChain(dirBlockInfo)
}

func saveToAnchorChain(dirBlockInfo *dbInfo.DirBlockInfo) {
	anchorLog.Debug("in saveToAnchorChain")
	anchorRec := new(AnchorRecord)
	anchorRec.AnchorRecordVer = 1
	anchorRec.DBHeight = dirBlockInfo.DBHeight
	anchorRec.KeyMR = dirBlockInfo.DBMerkleRoot.String()
	anchorRec.RecordHeight = uint32(state.GetDBHeight() + 1) // need the next block height
	anchorRec.Bitcoin.Address = defaultAddress.String()
	anchorRec.Bitcoin.TXID = dirBlockInfo.BTCTxHash.(*primitives.Hash).BTCString()
	anchorRec.Bitcoin.BlockHeight = dirBlockInfo.BTCBlockHeight
	anchorRec.Bitcoin.BlockHash = dirBlockInfo.BTCBlockHash.(*primitives.Hash).BTCString()
	anchorRec.Bitcoin.Offset = dirBlockInfo.BTCTxOffset
	anchorLog.Info("before submitting Entry To AnchorChain. anchor.record: " + spew.Sdump(anchorRec))

	err := submitEntryToAnchorChain(anchorRec)
	if err != nil {
		anchorLog.Error("Error in writing anchor into anchor chain: ", err.Error())
	}
}

func toHash(txHash *wire.ShaHash) *primitives.Hash {
	h := new(primitives.Hash)
	h.SetBytes(txHash.Bytes())
	return h
}

func toShaHash(hash interfaces.IHash) *wire.ShaHash {
	h, _ := wire.NewShaHash(hash.Bytes())
	return h
}

// UpdateDirBlockInfoMap allows factom processor to update DirBlockInfo
// when a new Directory Block is saved to db
func UpdateDirBlockInfoMap(dirBlockInfo *dbInfo.DirBlockInfo) {
	anchorLog.Debug("UpdateDirBlockInfoMap: ", spew.Sdump(dirBlockInfo))
	dirBlockInfoSlice = append(dirBlockInfoSlice, dirBlockInfo)
}

func checkForAnchor() {
	timeNow := time.Now().Unix()
	time0 := 60 * 60 * reAnchorAfter
	// anchor the latest dir block first
	sort.Sort(sort.Reverse(ByTimestamp(dirBlockInfoSlice)))
	for _, dirBlockInfo := range dirBlockInfoSlice {
		if bytes.Compare(dirBlockInfo.BTCTxHash.Bytes(), primitives.NewZeroHash().Bytes()) == 0 {
			anchorLog.Debug("first time anchor: ", spew.Sdump(dirBlockInfo))
			SendRawTransactionToBTC(dirBlockInfo.DBMerkleRoot, dirBlockInfo.DBHeight)
		} else {
			// This is the re-anchor case for the missed callback of malleated tx
			lapse := timeNow - dirBlockInfo.Timestamp
			if lapse > int64(time0) {
				anchorLog.Debugf("re-anchor: time lapse=%d, %s\n", lapse, spew.Sdump(dirBlockInfo))
				SendRawTransactionToBTC(dirBlockInfo.DBMerkleRoot, dirBlockInfo.DBHeight)
			}
		}
	}
}

func checkTxConfirmations() {
	timeNow := time.Now().Unix()
	time1 := 60 * 5 * confirmationsNeeded
	sort.Sort(ByTimestamp(dirBlockInfoSlice))
	for i, dirBlockInfo := range dirBlockInfoSlice {
		lapse := timeNow - dirBlockInfo.Timestamp
		if lapse > int64(time1) {
			anchorLog.Debugf("checkTxConfirmations: time lapse=%d", lapse)
			checkConfirmations(dirBlockInfo, i)
		}
	}
}

func checkConfirmations(dirBlockInfo *dbInfo.DirBlockInfo, index int) error {
	anchorLog.Debug("check Confirmations for btc tx: ", toShaHash(dirBlockInfo.BTCTxHash).String())
	txResult, err := wclient.GetTransaction(toShaHash(dirBlockInfo.BTCTxHash))
	if err != nil {
		anchorLog.Debugf(err.Error())
		return err
	}
	anchorLog.Debugf("GetTransactionResult: %s\n", spew.Sdump(txResult))
	if txResult.Confirmations >= int64(confirmationsNeeded) {
		btcBlockHash, _ := wire.NewShaHashFromStr(txResult.BlockHash)
		var rewrite = false
		// Either the call back is not recorded in case of BTCBlockHash is zero hash,
		// or bad things like re-organization of btc main chain happened
		if bytes.Compare(dirBlockInfo.BTCBlockHash.Bytes(), btcBlockHash.Bytes()) != 0 {
			anchorLog.Debugf("BTCBlockHash changed: original BTCBlockHeight=%d, original BTCBlockHash=%s, original tx offset=%d\n", dirBlockInfo.BTCBlockHeight, toShaHash(dirBlockInfo.BTCBlockHash).String(), dirBlockInfo.BTCTxOffset)
			dirBlockInfo.BTCBlockHash = toHash(btcBlockHash)
			btcBlock, err := wclient.GetBlockVerbose(btcBlockHash, true)
			if err != nil {
				anchorLog.Debugf(err.Error())
			}
			if btcBlock.Height > 0 {
				dirBlockInfo.BTCBlockHeight = int32(btcBlock.Height)
			}
			anchorLog.Debugf("BTCBlockHash changed: new BTCBlockHeight=%d, new BTCBlockHash=%s, btcBlockVerbose.Height=%d\n", dirBlockInfo.BTCBlockHeight, btcBlockHash.String(), btcBlock.Height)
			rewrite = true
		}
		dirBlockInfo.BTCConfirmed = true // needs confirmationsNeeded (20) to be confirmed.
		dirBlockInfo.Timestamp = time.Now().Unix()
		db.SaveDirBlockInfo(dirBlockInfo)
		dirBlockInfoSlice = append(dirBlockInfoSlice[:index], dirBlockInfoSlice[index+1:]...)		//delete it
		anchorLog.Debugf("Fully confirmed %d times. txid=%s, dirblockInfo=%s\n", txResult.Confirmations, txResult.TxID, spew.Sdump(dirBlockInfo))
		if rewrite {
			anchorLog.Debug("rewrite to anchor chain: ", spew.Sdump(dirBlockInfo))
			saveToAnchorChain(dirBlockInfo)
		}
	}
	return nil
}

// ByTimestamp defines the methods needed to satisify sort.Interface to
// sort a slice of DirBlockInfo by their Timestamp.
type ByTimestamp []*dbInfo.DirBlockInfo

func (u ByTimestamp) Len() int { return len(u) }
func (u ByTimestamp) Less(i, j int) bool {
	if u[i].Timestamp == u[j].Timestamp {
		return u[i].DBHeight < u[j].DBHeight
	}
	return u[i].Timestamp < u[j].Timestamp
}
func (u ByTimestamp) Swap(i, j int) { u[i], u[j] = u[j], u[i] }

func checkTxMalleation(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
	anchorLog.Debug("in checkTxMalleation")
	dirBlockInfos := make([]*dbInfo.DirBlockInfo, 0, len(dirBlockInfoSlice))
	for _, v := range dirBlockInfoSlice {
		// find those already anchored but no call back yet
		if v.BTCBlockHeight == 0 && bytes.Compare(v.BTCTxHash.Bytes(), primitives.NewZeroHash().Bytes()) != 0 {
			dirBlockInfos = append(dirBlockInfos, v)
		}
	}
	sort.Sort(ByTimestamp(dirBlockInfos))
	anchorLog.Debugf("malleated tx candidate count=%d, dirBlockInfo list=%s\n", len(dirBlockInfos), spew.Sdump(dirBlockInfos))
	for _, dirBlockInfo := range dirBlockInfos {
		tx, err := wclient.GetRawTransaction(toShaHash(dirBlockInfo.BTCTxHash))
		if err != nil {
			anchorLog.Debugf(err.Error())
			continue
		}
		anchorLog.Debugf("GetRawTransaction=%s, dirBlockInfo=%s\n", spew.Sdump(tx), spew.Sdump(dirBlockInfo))
		// compare OP_RETURN
		if reflect.DeepEqual(transaction.MsgTx().TxOut[0], tx.MsgTx().TxOut[0]) {
			anchorLog.Debugf("Tx Malleated: original.txid=%s, malleated.txid=%s\n", dirBlockInfo.BTCTxHash.(*primitives.Hash).BTCString(), transaction.Sha().String())
			doSaveDirBlockInfo(transaction, details, dirBlockInfo, true)
			break
		}
	}
}
