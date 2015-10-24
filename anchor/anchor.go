// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// logger is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package anchor

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FactomProject/factomd/log"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/btcsuitereleases/btcd/btcjson"
	"github.com/btcsuitereleases/btcd/chaincfg"
	"github.com/btcsuitereleases/btcd/txscript"
	"github.com/btcsuitereleases/btcd/wire"
	"github.com/btcsuitereleases/btcrpcclient"
	"github.com/btcsuitereleases/btcutil"
	"github.com/davecgh/go-spew/spew"

	factomwire "github.com/FactomProject/factomd/btcd/wire"
	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
)

var (
	balances            []balance // unspent balance & address & its WIF
	cfg                 *util.FactomdConfig
	dclient, wclient    *btcrpcclient.Client
	fee                 btcutil.Amount // tx fee for written into btc
	db                  interfaces.DBOverlay
	walletLocked        bool
	reAnchorAfter       = 10 // hours. For anchors that do not get bitcoin callback info for over 10 hours, then re-anchor them.
	reAnchorCheckEvery  = 1  // hour. do re-anchor check every 1 hour.
	defaultAddress      btcutil.Address
	confirmationsNeeded int

	//Server Private key for milestone 1
	serverPrivKey primitives.PrivateKey

	//Server Entry Credit private key
	serverECKey primitives.PrivateKey
	//Anchor chain ID
	anchorChainID interfaces.IHash
	//InmsgQ for submitting the entry to server
	inMsgQ chan factomwire.FtmInternalMsg
)

type balance struct {
	unspentResult btcjson.ListUnspentResult
	address       btcutil.Address
	wif           *btcutil.WIF
}

//AnchorRecord is used to construct anchor chain
type AnchorRecord struct {
	AnchorRecordVer int
	DBHeight        uint32
	KeyMR           string
	RecordHeight    uint32

	Bitcoin struct {
		Address     string //"1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1",
		TXID        string //"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5", BTC Hash - in reverse byte order
		BlockHeight int32  //345678,
		BlockHash   string //"00000000000000000cc14eacfc7057300aea87bed6fee904fd8e1c1f3dc008d4", BTC Hash - in reverse byte order
		Offset      int32  //87
	}
}

// SendRawTransactionToBTC is the main function used to anchor factom
// dir block hash to bitcoin blockchain
func SendRawTransactionToBTC(hash interfaces.IHash, blockHeight uint32) (*wire.ShaHash, error) {
	anchorLog.Debug("SendRawTransactionToBTC: hash=", hash.String(), ", dir block height=", blockHeight) //strconv.FormatUint(blockHeight, 10))
	err := sanityCheck(hash)
	if err != nil {
		return nil, err
	}
	return doTransaction(hash, blockHeight)
}

func doTransaction(hash interfaces.IHash, blockHeight uint32) (*wire.ShaHash, error) {
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

	return shaHash, nil
}

func sanityCheck(hash interfaces.IHash) error {
	if dclient == nil || wclient == nil {
		s := fmt.Sprintf("\n\n$$$ WARNING: rpc clients and/or wallet are not initiated successfully. No anchoring for now.\n")
		anchorLog.Warning(s)
		return errors.New(s)
	}
	if len(balances) == 0 {
		anchorLog.Warning("len(balances) == 0, start rescan UTXO *** ")
		updateUTXO()
	}
	if len(balances) == 0 {
		s := fmt.Sprintf("\n\n$$$ WARNING: No balance in your wallet. No anchoring for now.\n")
		anchorLog.Warning(s)
		return errors.New(s)
	}
	return nil
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

func selectInputs(eligible []btcjson.ListUnspentResult, minconf int) (selected []btcjson.ListUnspentResult, out btcutil.Amount, err error) {
	// Iterate throguh eligible transactions, appending to outputs and
	// increasing out.  This is finished when out is greater than the
	// requested amt to spend.
	selected = make([]btcjson.ListUnspentResult, 0, len(eligible))
	for _, e := range eligible {
		amount, err := btcutil.NewAmount(e.Amount)
		if err != nil {
			anchorLog.Error("err in creating NewAmount")
			continue
		}
		selected = append(selected, e)
		out += amount
		if out >= fee {
			return selected, out, nil
		}
	}
	if out < fee {
		return nil, 0, fmt.Errorf("insufficient funds: transaction requires %v fee, but only %v spendable", fee, out)
	}

	return selected, out, nil
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
		//engine, err := txscript.NewScript(txin.SignatureScript, scriptPubKey, i, msgtx, flags)
		engine, err := txscript.NewEngine(scriptPubKey, msgtx, i, flags)
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
		// Hitting OOM by growing or writing to a bytes.Buffer already
		// panics, and all returned errors are unexpected.
		//panic(err)
		//TODO: should we have retry logic?
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
		OnAccountBalance: func(account string, balance btcutil.Amount, confirmed bool) {
			//go newBalance(account, balance, confirmed)
			//anchorLog.Info("wclient: OnAccountBalance, account=", account, ", balance=",
			//balance.ToUnit(btcutil.AmountBTC), ", confirmed=", confirmed)
		},

		OnWalletLockState: func(locked bool) {
			anchorLog.Info("wclient: OnWalletLockState, locked=", locked)
			walletLocked = locked
		},

		OnUnknownNotification: func(method string, params []json.RawMessage) {
			//anchorLog.Info("wclient: OnUnknownNotification: method=", method, "\nparams[0]=",
			//string(params[0]), "\nparam[1]=", string(params[1]))
		},
	}

	return ntfnHandlers
}

func createBtcdNotificationHandlers() btcrpcclient.NotificationHandlers {

	ntfnHandlers := btcrpcclient.NotificationHandlers{

		OnBlockConnected: func(hash *wire.ShaHash, height int32) {
			//anchorLog.Info("dclient: OnBlockConnected: hash=", hash, ", height=", height)
			//go newBlock(hash, height)	// no need
		},

		OnRecvTx: func(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
			//anchorLog.Info("dclient: OnRecvTx: details=%#v\n", details)
			//anchorLog.Info("dclient: OnRecvTx: tx=%#v,  tx.Sha=%#v, tx.index=%d\n",
			//transaction, transaction.Sha().String(), transaction.Index())
		},

		OnRedeemingTx: func(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
			//anchorLog.Info("dclient: OnRedeemingTx: details=%#v\n", details)
			//anchorLog.Info("dclient: OnRedeemingTx: tx.Sha=%#v,  tx.index=%d\n",
			//transaction.Sha().String(), transaction.Index())
		},
	}

	return ntfnHandlers
}

// InitAnchor inits rpc clients for factom
// and load up unconfirmed DirBlockInfo from leveldb
func InitAnchor(do interfaces.DBOverlay, q chan factomwire.FtmInternalMsg, serverKey primitives.PrivateKey) {
	anchorLog.Debug("InitAnchor")
	db = do
	inMsgQ = q
	serverPrivKey = serverKey

	var err error

	if err = initRPCClient(); err != nil {
		anchorLog.Error(err.Error())
		return
	}
	//defer shutdown(dclient)
	//defer shutdown(wclient)

	if err = initWallet(); err != nil {
		anchorLog.Error(err.Error())
		return
	}

	ticker := time.NewTicker(time.Hour * time.Duration(reAnchorCheckEvery))
	go func() {
		for _ = range ticker.C {
			// check init rpc client
			if dclient == nil || wclient == nil {
				if err = initRPCClient(); err != nil {
					anchorLog.Error(err.Error())
				}
			}
		}
	}()

	return
}

func initRPCClient() error {
	anchorLog.Debug("init RPC client")
	cfg = util.ReadConfig()
	certHomePath := cfg.Btc.CertHomePath
	rpcClientHost := cfg.Btc.RpcClientHost
	rpcClientEndpoint := cfg.Btc.RpcClientEndpoint
	rpcClientUser := cfg.Btc.RpcClientUser
	rpcClientPass := cfg.Btc.RpcClientPass
	certHomePathBtcd := cfg.Btc.CertHomePathBtcd
	rpcBtcdHost := cfg.Btc.RpcBtcdHost
	confirmationsNeeded = cfg.Anchor.ConfirmationsNeeded

	//Added anchor parameters
	var err error
	//TODO: double check if the key is properly generated
	serverECKey, err = primitives.NewPrivateKeyFromHex(cfg.Anchor.ServerECPrivKey)
	if err != nil {
		panic("Cannot parse Server EC Key from configuration file: " + err.Error())
	}
	anchorChainID, err = primitives.HexToHash(cfg.Anchor.AnchorChainID)
	anchorLog.Debug("anchorChainID: ", anchorChainID)
	if err != nil || anchorChainID == nil {
		panic("Cannot parse Server AnchorChainID from configuration file: " + err.Error())
	}

	// Connect to local btcwallet RPC server using websockets.
	ntfnHandlers := createBtcwalletNotificationHandlers()
	certHomeDir := btcutil.AppDataDir(certHomePath, false)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return fmt.Errorf("cannot read rpc.cert file: %s", err)
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
		return fmt.Errorf("cannot create rpc client for btcwallet: %s", err)
	}

	// Connect to local btcd RPC server using websockets.
	dntfnHandlers := createBtcdNotificationHandlers()
	certHomeDir = btcutil.AppDataDir(certHomePathBtcd, false)
	certs, err = ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return fmt.Errorf("cannot read rpc.cert file for btcd rpc server: %s", err)
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
		return fmt.Errorf("cannot create rpc client for btcd: %s", err)
	}

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

func initWallet() error {
	balances = make([]balance, 0, 200)
	fee, _ = btcutil.NewAmount(cfg.Btc.BtcTransFee)
	walletLocked = true
	err := updateUTXO()
	if err == nil && len(balances) > 0 {
		defaultAddress = balances[0].address
	}
	return err
}

func updateUTXO() error {
	anchorLog.Info("updateUTXO: walletLocked=", walletLocked)
	balances = make([]balance, 0, 200)
	//if walletLocked {
	err := unlockWallet(int64(6)) //600
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	//}

	unspentResults, err := wclient.ListUnspentMin(confirmationsNeeded) //minConf=1
	if err != nil {
		return fmt.Errorf("cannot list unspent. %s", err)
	}
	anchorLog.Info("updateUTXO: unspentResults.len=", len(unspentResults))

	if len(unspentResults) > 0 {
		var i int
		for _, b := range unspentResults {
			if b.Amount > fee.ToBTC() {
				balances = append(balances, balance{unspentResult: b})
				i++
			}
		}
	}
	anchorLog.Info("updateUTXO: balances.len=", len(balances))

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

	//time.Sleep(1 * time.Second)
	return nil
}

func shutdown(client *btcrpcclient.Client) {
	if client == nil {
		return
	}
	// For this example gracefully shutdown the client after 10 seconds.
	// Ordinarily when to shutdown the client is highly application
	// specific.
	log.Println("Client shutdown in 2 seconds...")
	time.AfterFunc(time.Second*2, func() {
		log.Println("Going down...")
		client.Shutdown()
	})
	defer log.Println("btcsuite client shutdown is done!")
	// Wait until the client either shuts down gracefully (or the user
	// terminates the process with Ctrl+C).
	client.WaitForShutdown()
}

func prependBlockHeight(height uint32, hash []byte) ([]byte, error) {
	// dir block genesis block height starts with 0, for now
	// similar to bitcoin genesis block
	//if (0 == height) || (0xFFFFFFFFFFFF&height != height) {
	//if 0xFFFFFFFFFFFF&height != height {
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

func toHash(txHash interfaces.IHash) interfaces.IHash {
	h := new(primitives.Hash)
	h.SetBytes(txHash.Bytes())
	return h
}
