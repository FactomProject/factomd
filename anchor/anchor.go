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

	"github.com/FactomProject/go-spew/spew"
	"github.com/btcsuitereleases/btcd/btcjson"
	"github.com/btcsuitereleases/btcd/chaincfg"
	"github.com/btcsuitereleases/btcd/txscript"
	"github.com/btcsuitereleases/btcd/wire"
	"github.com/btcsuitereleases/btcrpcclient"
	"github.com/btcsuitereleases/btcutil"

	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
)

type Anchor struct {
	balances            []balance // unspent balance & address & its WIF
	cfg                 *util.FactomdConfig
	dclient, wclient    *btcrpcclient.Client
	dirBlockInfoSlice   []interfaces.IDirBlockInfo
	db                  interfaces.DBOverlay
	walletLocked        bool
	reAnchorAfter       int // hours. For anchors that do not get bitcoin callback info for over 10 hours, then re-anchor them.
	tenMinutes          int // 10 minute mark
	defaultAddress      btcutil.Address
	minBalance          btcutil.Amount // 0.01 btc
	fee                 btcutil.Amount // tx fee for written into btc
	confirmationsNeeded int

	serverPrivKey *primitives.PrivateKey //Server Private key for milestone 1
	serverECKey   *primitives.PrivateKey //Server Entry Credit private key
	anchorChainID interfaces.IHash
	state         interfaces.IState
}

var _ interfaces.IAnchor = (*Anchor)(nil)

func NewAnchor() *Anchor {
	a := new(Anchor)
	a.walletLocked = true
	a.reAnchorAfter = 4 // hours. For anchors that do not get bitcoin callback info for over 10 hours, then re-anchor them.
	a.tenMinutes = 10   // 10 minute mark
	return a
}

type balance struct {
	unspentResult btcjson.ListUnspentResult
	address       btcutil.Address
	wif           *btcutil.WIF
}

// SendRawTransactionToBTC is the main function used to anchor factom
// dir block hash to bitcoin blockchain
func (a *Anchor) SendRawTransactionToBTC(hash interfaces.IHash, blockHeight uint32) (*wire.ShaHash, error) {
	anchorLog.Debug("SendRawTransactionToBTC: hash=", hash.String(), ", dir block height=", blockHeight) //strconv.FormatUint(blockHeight, 10))
	dirBlockInfo, err := a.sanityCheck(hash)
	if err != nil {
		return nil, err
	}
	return a.doTransaction(hash, blockHeight, dirBlockInfo)
}

func (a *Anchor) doTransaction(hash interfaces.IHash, blockHeight uint32, dirBlockInfo *dbInfo.DirBlockInfo) (*wire.ShaHash, error) {
	b := a.balances[0]
	a.balances = a.balances[1:]
	anchorLog.Info("new balances.len=", len(a.balances))

	msgtx, err := a.createRawTransaction(b, hash.Bytes(), blockHeight)
	if err != nil {
		return nil, fmt.Errorf("cannot create Raw Transaction: %s", err)
	}

	shaHash, err := a.sendRawTransaction(msgtx)
	if err != nil {
		return nil, fmt.Errorf("cannot send Raw Transaction: %s", err)
	}

	if dirBlockInfo != nil {
		dirBlockInfo.BTCTxHash = toHash(shaHash)
		dirBlockInfo.SetTimestamp(primitives.NewTimestampNow())
		a.db.SaveDirBlockInfo(dirBlockInfo)
	}

	return shaHash, nil
}

func (a *Anchor) sanityCheck(hash interfaces.IHash) (*dbInfo.DirBlockInfo, error) {
	index := sort.Search(len(a.dirBlockInfoSlice), func(i int) bool {
		return a.dirBlockInfoSlice[i].GetDBMerkleRoot().IsSameAs(hash)
	})
	var dirBlockInfo *dbInfo.DirBlockInfo
	ok := false
	if index < len(a.dirBlockInfoSlice) {
		dirBlockInfo, ok = a.dirBlockInfoSlice[index].(*dbInfo.DirBlockInfo)
	}
	if dirBlockInfo == nil || !ok {
		s := fmt.Sprintf("Anchor Error: hash %s does not exist in dirBlockInfoSlice.\n", hash.String())
		anchorLog.Error(s)
		return nil, errors.New(s)
	}
	if dirBlockInfo.BTCConfirmed {
		s := fmt.Sprintf("Anchor Warning: hash %s has already been confirmed in btc block chain.\n", hash.String())
		anchorLog.Error(s)
		return nil, errors.New(s)
	}
	if a.dclient == nil || a.wclient == nil {
		s := fmt.Sprintf("\n\n$$$ WARNING: rpc clients and/or wallet are not initiated successfully. No anchoring for now.\n")
		anchorLog.Warning(s)
		return nil, errors.New(s)
	}
	if len(a.balances) == 0 {
		anchorLog.Warning("len(balances) == 0, start rescan UTXO *** ")
		a.updateUTXO(a.minBalance)
	}
	if len(a.balances) == 0 {
		anchorLog.Warning("len(balances) == 0, start rescan UTXO *** ")
		a.updateUTXO(a.fee)
	}
	if len(a.balances) == 0 {
		s := fmt.Sprintf("\n\n$$$ WARNING: No balance in your wallet. No anchoring for now.\n")
		anchorLog.Warning(s)
		return nil, errors.New(s)
	}
	return dirBlockInfo, nil
}

func (a *Anchor) createRawTransaction(b balance, hash []byte, blockHeight uint32) (*wire.MsgTx, error) {
	msgtx := wire.NewMsgTx()

	if err := a.addTxOuts(msgtx, b, hash, blockHeight); err != nil {
		return nil, fmt.Errorf("cannot addTxOuts: %s", err)
	}

	if err := a.addTxIn(msgtx, b); err != nil {
		return nil, fmt.Errorf("cannot addTxIn: %s", err)
	}

	if err := validateMsgTx(msgtx, []btcjson.ListUnspentResult{b.unspentResult}); err != nil {
		return nil, fmt.Errorf("cannot validateMsgTx: %s", err)
	}

	return msgtx, nil
}

func (a *Anchor) addTxIn(msgtx *wire.MsgTx, b balance) error {
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
	err = a.dclient.NotifySpent([]*wire.OutPoint{outPoint})
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

func (a *Anchor) addTxOuts(msgtx *wire.MsgTx, b balance, hash []byte, blockHeight uint32) error {
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
	change := amount - a.fee

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

func (a *Anchor) sendRawTransaction(msgtx *wire.MsgTx) (*wire.ShaHash, error) {
	//anchorLog.Debug("sendRawTransaction: msgTx=", spew.Sdump(msgtx))
	buf := bytes.Buffer{}
	buf.Grow(msgtx.SerializeSize())
	if err := msgtx.BtcEncode(&buf, wire.ProtocolVersion); err != nil {
		return nil, err
	}

	// use rpc client for btcd here for better callback info
	// this should not require wallet to be unlocked
	shaHash, err := a.dclient.SendRawTransaction(msgtx, false)
	if err != nil {
		return nil, fmt.Errorf("failed in rpcclient.SendRawTransaction: %s", err)
	}
	anchorLog.Info("btc txHash returned: ", shaHash) // new tx hash
	return shaHash, nil
}

func (a *Anchor) createBtcwalletNotificationHandlers() btcrpcclient.NotificationHandlers {
	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnWalletLockState: func(locked bool) {
			anchorLog.Info("wclient: OnWalletLockState, locked=", locked)
			a.walletLocked = locked
		},
	}
	return ntfnHandlers
}

func (a *Anchor) createBtcdNotificationHandlers() btcrpcclient.NotificationHandlers {
	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnRedeemingTx: func(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
			if details != nil {
				// do not block OnRedeemingTx callback
				//anchorLog.Info(" saveDirBlockInfo.")

				//TODO: handle concurrance better
				go a.saveDirBlockInfo(transaction, details)
			}
		},
	}
	return ntfnHandlers
}

func (a *Anchor) checkMissingDirBlockInfo() {
	anchorLog.Debug("checkMissingDirBlockInfo for those unsaved DirBlocks in database")
	dblocks, _ := a.db.FetchAllDBlocks()
	dirBlockInfos, _ := a.db.FetchAllDirBlockInfos() //FetchAllDirBlockInfos()
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
			dirBlockInfo.SetTimestamp(primitives.NewTimestampNow())
			anchorLog.Debug("add missing dirBlockInfo to map: ", spew.Sdump(dirBlockInfo))
			a.db.SaveDirBlockInfo(dirBlockInfo)
			a.dirBlockInfoSlice = append(a.dirBlockInfoSlice, dirBlockInfo)
		}
	}
}

// InitAnchor inits rpc clients for factom
// and load up unconfirmed DirBlockInfo from leveldb

func (a *Anchor) readConfig() {
	anchorLog.Info("readConfig")
	a.cfg = util.ReadConfig("")
	a.confirmationsNeeded = a.cfg.Anchor.ConfirmationsNeeded
	a.fee, _ = btcutil.NewAmount(a.cfg.Btc.BtcTransFee)

	var err error
	a.serverPrivKey, err = primitives.NewPrivateKeyFromHex(a.cfg.App.LocalServerPrivKey)
	if err != nil {
		panic("Cannot parse Server Private Key from configuration file: " + err.Error())
	}
	a.serverECKey, err = primitives.NewPrivateKeyFromHex(a.cfg.Anchor.ServerECPrivKey)
	if err != nil {
		panic("Cannot parse Server EC Key from configuration file: " + err.Error())
	}
	a.anchorChainID, err = primitives.HexToHash(a.cfg.Anchor.AnchorChainID)
	anchorLog.Debug("anchorChainID: ", a.anchorChainID)
	if err != nil || a.anchorChainID == nil {
		panic("Cannot parse Server AnchorChainID from configuration file: " + err.Error())
	}
}

// InitRPCClient is used to create rpc client for btcd and btcwallet
// and it can be used to test connecting to btcd / btcwallet servers
// running in different machine.
func (a *Anchor) InitRPCClient() error {
	anchorLog.Debug("init RPC client")
	if a.cfg == nil {
		a.readConfig()
	}
	certHomePath := a.cfg.Btc.CertHomePath
	rpcClientHost := a.cfg.Btc.RpcClientHost
	rpcClientEndpoint := a.cfg.Btc.RpcClientEndpoint
	rpcClientUser := a.cfg.Btc.RpcClientUser
	rpcClientPass := a.cfg.Btc.RpcClientPass
	certHomePathBtcd := a.cfg.Btc.CertHomePathBtcd
	rpcBtcdHost := a.cfg.Btc.RpcBtcdHost

	// Connect to local btcwallet RPC server using websockets.
	ntfnHandlers := a.createBtcwalletNotificationHandlers()
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
	a.wclient, err = btcrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		return fmt.Errorf("cannot create rpc client for btcwallet: %s\n", err)
	}
	anchorLog.Debug("successfully created rpc client for btcwallet")

	// Connect to local btcd RPC server using websockets.
	dntfnHandlers := a.createBtcdNotificationHandlers()
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
	a.dclient, err = btcrpcclient.New(dconnCfg, &dntfnHandlers)
	if err != nil {
		return fmt.Errorf("cannot create rpc client for btcd: %s\n", err)
	}
	anchorLog.Debug("successfully created rpc client for btcd")

	return nil
}

func (a *Anchor) unlockWallet(timeoutSecs int64) error {
	err := a.wclient.WalletPassphrase(a.cfg.Btc.WalletPassphrase, int64(timeoutSecs))
	if err != nil {
		return fmt.Errorf("cannot unlock wallet with passphrase: %s", err)
	}
	a.walletLocked = false
	return nil
}

// ByAmount defines the methods needed to satisify sort.Interface to
// sort a slice of UTXOs by their amount.
type ByAmount []balance

func (u ByAmount) Len() int           { return len(u) }
func (u ByAmount) Less(i, j int) bool { return u[i].unspentResult.Amount < u[j].unspentResult.Amount }
func (u ByAmount) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

func (a *Anchor) updateUTXO(base btcutil.Amount) error {
	anchorLog.Info("updateUTXO: base=", base.ToBTC())
	if a.wclient == nil {
		anchorLog.Info("updateUTXO: wclient is nil")
		return nil
	}
	a.balances = make([]balance, 0, 200)
	err := a.unlockWallet(int64(6)) //600
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	unspentResults, err := a.wclient.ListUnspentMin(a.confirmationsNeeded) //minConf=1
	if err != nil {
		return fmt.Errorf("cannot list unspent. %s", err)
	}
	anchorLog.Info("updateUTXO: unspentResults.len=", len(unspentResults))

	if len(unspentResults) > 0 {
		var i int
		for _, b := range unspentResults {
			if b.Amount > base.ToBTC() { //fee.ToBTC()
				a.balances = append(a.balances, balance{unspentResult: b})
				i++
			}
		}
	}
	anchorLog.Info("updateUTXO: balances.len=", len(a.balances))

	// Sort eligible balances so that we first pick the ones with highest one
	sort.Sort(sort.Reverse(ByAmount(a.balances)))

	for i, b := range a.balances {
		addr, err := btcutil.DecodeAddress(b.unspentResult.Address, &chaincfg.TestNet3Params)
		if err != nil {
			return fmt.Errorf("cannot decode address: %s", err)
		}
		a.balances[i].address = addr

		wif, err := a.wclient.DumpPrivKey(addr)
		if err != nil {
			return fmt.Errorf("cannot get WIF: %s", err)
		}
		a.balances[i].wif = wif
		//anchorLog.Infof("balance[%d]=%s \n", i, spew.Sdump(balances[i]))
	}

	if len(a.balances) > 0 {
		a.defaultAddress = a.balances[0].address
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

func (a *Anchor) saveDirBlockInfo(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
	anchorLog.Debug("in saveDirBlockInfo")
	var saved = false
	for _, dirBlockInfo := range a.dirBlockInfoSlice {
		if bytes.Compare(dirBlockInfo.GetBTCTxHash().Bytes(), transaction.Sha().Bytes()) == 0 {
			a.doSaveDirBlockInfo(transaction, details, dirBlockInfo.(*dbInfo.DirBlockInfo), false)
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
		a.checkTxMalleation(transaction, details)
	}
}

func (a *Anchor) doSaveDirBlockInfo(transaction *btcutil.Tx, details *btcjson.BlockDetails, dirBlockInfo *dbInfo.DirBlockInfo, replace bool) {
	if replace {
		dirBlockInfo.BTCTxHash = toHash(transaction.Sha()) // in case of tx being malleated
	}
	dirBlockInfo.BTCTxOffset = int32(details.Index)
	dirBlockInfo.BTCBlockHeight = details.Height
	btcBlockHash, _ := wire.NewShaHashFromStr(details.Hash)
	dirBlockInfo.BTCBlockHash = toHash(btcBlockHash)
	dirBlockInfo.SetTimestamp(primitives.NewTimestampNow())
	a.db.SaveDirBlockInfo(dirBlockInfo)
	anchorLog.Infof("In doSaveDirBlockInfo, dirBlockInfo:%s saved to db\n", spew.Sdump(dirBlockInfo))

	// to make factom / explorer more user friendly, instead of waiting for
	// over 2 hours to know it's anchored, we can create the anchor chain instantly
	// then change it when the btc main chain re-org happens.
	a.saveToAnchorChain(dirBlockInfo)
}

func (a *Anchor) saveToAnchorChain(dirBlockInfo *dbInfo.DirBlockInfo) {
	anchorLog.Debug("in saveToAnchorChain")
	anchorRec := new(AnchorRecord)
	anchorRec.AnchorRecordVer = 1
	anchorRec.DBHeight = dirBlockInfo.GetDBHeight()
	anchorRec.KeyMR = dirBlockInfo.GetDBMerkleRoot().String()
	anchorRec.RecordHeight = a.state.GetHighestCompletedBlock() // need the next block height
	anchorRec.Bitcoin = new(BitcoinStruct)
	anchorRec.Bitcoin.Address = a.defaultAddress.String()
	anchorRec.Bitcoin.TXID = dirBlockInfo.GetBTCTxHash().(*primitives.Hash).BTCString()
	anchorRec.Bitcoin.BlockHeight = dirBlockInfo.BTCBlockHeight
	anchorRec.Bitcoin.BlockHash = dirBlockInfo.BTCBlockHash.(*primitives.Hash).BTCString()
	anchorRec.Bitcoin.Offset = dirBlockInfo.BTCTxOffset
	anchorLog.Info("before submitting Entry To AnchorChain. anchor.record: " + spew.Sdump(anchorRec))

	err := a.submitEntryToAnchorChain(anchorRec)
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
func (a *Anchor) UpdateDirBlockInfoMap(dirBlockInfo interfaces.IDirBlockInfo) {
	anchorLog.Debug("UpdateDirBlockInfoMap: ", spew.Sdump(dirBlockInfo))
	a.dirBlockInfoSlice = append(a.dirBlockInfoSlice, dirBlockInfo.(*dbInfo.DirBlockInfo))
}

func (a *Anchor) checkForAnchor() {
	timeNow := time.Now().Unix()
	time0 := 60 * 60 * a.reAnchorAfter
	// anchor the latest dir block first
	sort.Sort(sort.Reverse(util.ByDirBlockInfoTimestamp(a.dirBlockInfoSlice)))
	for _, dirBlockInfo := range a.dirBlockInfoSlice {
		if bytes.Compare(dirBlockInfo.GetBTCTxHash().Bytes(), primitives.NewZeroHash().Bytes()) == 0 {
			anchorLog.Debug("first time anchor: ", spew.Sdump(dirBlockInfo))
			a.SendRawTransactionToBTC(dirBlockInfo.GetDBMerkleRoot(), dirBlockInfo.GetDBHeight())
		} else {
			// This is the re-anchor case for the missed callback of malleated tx
			lapse := timeNow - dirBlockInfo.GetTimestamp().GetTimeSeconds()
			if lapse > int64(time0) {
				anchorLog.Debugf("re-anchor: time lapse=%d, %s\n", lapse, spew.Sdump(dirBlockInfo))
				a.SendRawTransactionToBTC(dirBlockInfo.GetDBMerkleRoot(), dirBlockInfo.GetDBHeight())
			}
		}
	}
}

func (a *Anchor) checkTxConfirmations() {
	timeNow := time.Now().Unix()
	time1 := 60 * 5 * a.confirmationsNeeded
	sort.Sort(util.ByDirBlockInfoTimestamp(a.dirBlockInfoSlice))
	for i, dirBlockInfo := range a.dirBlockInfoSlice {
		lapse := timeNow - dirBlockInfo.GetTimestamp().GetTimeSeconds()
		if lapse > int64(time1) {
			anchorLog.Debugf("checkTxConfirmations: time lapse=%d", lapse)
			a.checkConfirmations(dirBlockInfo.(*dbInfo.DirBlockInfo), i)
		}
	}
}

func (a *Anchor) checkConfirmations(dirBlockInfo *dbInfo.DirBlockInfo, index int) error {
	anchorLog.Debug("check Confirmations for btc tx: ", toShaHash(dirBlockInfo.GetBTCTxHash()).String())
	txResult, err := a.wclient.GetTransaction(toShaHash(dirBlockInfo.GetBTCTxHash()))
	if err != nil {
		anchorLog.Debugf(err.Error())
		return err
	}
	anchorLog.Debugf("GetTransactionResult: %s\n", spew.Sdump(txResult))
	if txResult.Confirmations >= int64(a.confirmationsNeeded) {
		btcBlockHash, _ := wire.NewShaHashFromStr(txResult.BlockHash)
		var rewrite = false
		// Either the call back is not recorded in case of BTCBlockHash is zero hash,
		// or bad things like re-organization of btc main chain happened
		if bytes.Compare(dirBlockInfo.BTCBlockHash.Bytes(), btcBlockHash.Bytes()) != 0 {
			anchorLog.Debugf("BTCBlockHash changed: original BTCBlockHeight=%d, original BTCBlockHash=%s, original tx offset=%d\n", dirBlockInfo.BTCBlockHeight, toShaHash(dirBlockInfo.BTCBlockHash).String(), dirBlockInfo.BTCTxOffset)
			dirBlockInfo.BTCBlockHash = toHash(btcBlockHash)
			btcBlock, err := a.wclient.GetBlockVerbose(btcBlockHash, true)
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
		dirBlockInfo.SetTimestamp(primitives.NewTimestampNow())
		a.db.SaveDirBlockInfo(dirBlockInfo)
		a.dirBlockInfoSlice = append(a.dirBlockInfoSlice[:index], a.dirBlockInfoSlice[index+1:]...) //delete it
		anchorLog.Debugf("Fully confirmed %d times. txid=%s, dirblockInfo=%s\n", txResult.Confirmations, txResult.TxID, spew.Sdump(dirBlockInfo))
		if rewrite {
			anchorLog.Debug("rewrite to anchor chain: ", spew.Sdump(dirBlockInfo))
			a.saveToAnchorChain(dirBlockInfo)
		}
	}
	return nil
}

func (a *Anchor) checkTxMalleation(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
	anchorLog.Debug("in checkTxMalleation")
	dirBlockInfos := make([]interfaces.IDirBlockInfo, 0, len(a.dirBlockInfoSlice))
	for _, v := range a.dirBlockInfoSlice {
		// find those already anchored but no call back yet
		if v.GetBTCBlockHeight() == 0 && bytes.Compare(v.GetBTCTxHash().Bytes(), primitives.NewZeroHash().Bytes()) != 0 {
			dirBlockInfos = append(dirBlockInfos, v)
		}
	}
	sort.Sort(util.ByDirBlockInfoTimestamp(dirBlockInfos))
	anchorLog.Debugf("malleated tx candidate count=%d, dirBlockInfo list=%s\n", len(dirBlockInfos), spew.Sdump(dirBlockInfos))
	for _, dirBlockInfo := range dirBlockInfos {
		tx, err := a.wclient.GetRawTransaction(toShaHash(dirBlockInfo.GetBTCTxHash()))
		if err != nil {
			anchorLog.Debugf(err.Error())
			continue
		}
		anchorLog.Debugf("GetRawTransaction=%s, dirBlockInfo=%s\n", spew.Sdump(tx), spew.Sdump(dirBlockInfo))
		// compare OP_RETURN
		if reflect.DeepEqual(transaction.MsgTx().TxOut[0], tx.MsgTx().TxOut[0]) {
			anchorLog.Debugf("Tx Malleated: original.txid=%s, malleated.txid=%s\n", dirBlockInfo.GetBTCTxHash().(*primitives.Hash).BTCString(), transaction.Sha().String())
			a.doSaveDirBlockInfo(transaction, details, dirBlockInfo.(*dbInfo.DirBlockInfo), true)
			break
		}
	}
}
