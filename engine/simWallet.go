package engine

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/wsapi"
)

// FundWallet()
// Entry Point for no time offset on the transaction.
func FundWallet(st *state.State, amt uint64) (error, string) {
	return FundWalletTOFF(st, 0, amt)
}

// REVIEW: consider relocating many of these functions to testHelpers/simWallet.go

// FundWalletTOFF()
// Entry Point where test code allows the transaction to have a time offset from the current time.
func FundWalletTOFF(st *state.State, timeOffsetInMilliseconds int64, amt uint64) (error, string) {
	inSec, _ := primitives.HexToHash("FB3B471B1DCDADFEB856BD0B02D8BF49ACE0EDD372A3D9F2A95B78EC12A324D6") // private key or FCT Source
	outEC, _ := primitives.HexToHash("c23ae8eec2beb181a0da926bd2344e988149fbe839fbc7489f2096e7d6110243") // EC address

	// So what we are going to do is get the current time in ms, add to it the offset provided (usually zero, except
	// for tests)
	ts := primitives.NewTimestampFromMilliseconds(uint64(primitives.NewTimestampNow().GetTimeMilli() + timeOffsetInMilliseconds))

	return fundECWallet(st, inSec, outEC, ts, amt)
}

// FundECWallet get the current time in ms, add to it the offset provided (usually zero, except for tests)
func FundECWallet(st *state.State, inSec interfaces.IHash, outEC interfaces.IHash, amt uint64) (error, string) {
	ts := primitives.NewTimestampFromMilliseconds(uint64(primitives.NewTimestampNow().GetTimeMilli()))
	return fundECWallet(st, inSec, outEC, ts, amt)
}

// fundEDWallet() buys EC credits adds fee on top of amt
func fundECWallet(st *state.State, inSec interfaces.IHash, outEC interfaces.IHash, timeInMilliseconds *primitives.Timestamp, amt uint64) (error, string) {

	trans, err := ComposeEcTransaction(inSec, outEC, timeInMilliseconds, amt, st.GetFactoshisPerEC())
	if err != nil {
		return err, "Failed to build transaction"
	}

	// FIXME: consider building msg and pushing onto API Queue instead
	return PostTransaction(st, trans)
}

// create wsapi Post and invoke v2Request handler
func PostTransaction(st *state.State, trans *factoid.Transaction) (error, string) {

	t := new(wsapi.TransactionRequest)
	data, _ := trans.MarshalBinary()
	t.Transaction = hex.EncodeToString(data)
	j := primitives.NewJSON2Request("factoid-submit", 0, t)
	_, err := v2Request(j, st.GetPort())

	if err != nil {
		return err, ""
	}

	return nil, fmt.Sprintf("%v", trans.GetTxID())
}

// SendTxn() adds transaction to APIQueue bypassing the wsapi / json encoding
func SendTxn(s *state.State, amt uint64, userSecretIn string, userPubOut string, ecPrice uint64) (*factoid.Transaction, error) {
	txn, _ := NewTransaction(amt, userSecretIn, userPubOut, ecPrice)
	msg := new(messages.FactoidTransaction)
	msg.SetTransaction(txn)
	s.APIQueue().Enqueue(msg)
	return txn, nil
}

func GetBalance(s *state.State, userStr string) int64 {
	return s.FactoidState.GetFactoidBalance(factoid.NewAddress(primitives.ConvertUserStrToAddress(userStr)).Fixed())
}

func GetBalanceEC(s *state.State, userStr string) int64 {
	a := factoid.NewAddress(primitives.ConvertUserStrToAddress(userStr))
	return s.GetFactoidState().GetECBalance(a.Fixed())
}

// generate a pair of user-strings Fs.., FA..
func RandomFctAddressPair() (string, string) {
	pkey := primitives.RandomPrivateKey()
	privUserStr, _ := primitives.PrivateKeyStringToHumanReadableFactoidPrivateKey(pkey.PrivateKeyString())
	_, _, pubUserStr, _ := factoid.PrivateKeyStringToEverythingString(pkey.PrivateKeyString())

	return privUserStr, pubUserStr
}

// construct a new factoid transaction
func NewTransaction(amt uint64, userSecretIn string, userPublicOut string, ecPrice uint64) (*factoid.Transaction, error) {
	inSec := factoid.NewAddress(primitives.ConvertUserStrToAddress(userSecretIn))
	outPub := factoid.NewAddress(primitives.ConvertUserStrToAddress(userPublicOut))
	return ComposeFctTransaction(amt, inSec, outPub, ecPrice)
}

// create a transaction to transfer FCT between addresses
// adds EC fee on top of input amount
func ComposeFctTransaction(amt uint64, inSec interfaces.IHash, outPub interfaces.IHash, ecPrice uint64) (*factoid.Transaction, error) {

	var sec [64]byte
	copy(sec[:32], inSec.Bytes())   // pass 32 byte key in a 64 byte field for the crypto library
	pub := ed.GetPublicKey(&sec)    // get the public key for our FCT source address
	rcd := factoid.NewRCD_1(pub[:]) // build the an RCD "redeem condition data structure"

	inAdd, err := rcd.GetAddress()
	if err != nil {
		panic(err)
	}

	trans := new(factoid.Transaction)
	trans.AddInput(inAdd, amt)
	trans.AddOutput(outPub, amt)

	// REVIEW: why is this not needed?
	// seems to be different from FundWallet() ?
	//trans.AddRCD(rcd)
	trans.AddAuthorization(rcd)
	trans.SetTimestamp(primitives.NewTimestampNow())

	fee, err := trans.CalculateFee(ecPrice)
	if err != nil {
		return trans, err
	}

	input, err := trans.GetInput(0)
	if err != nil {
		return trans, err
	}
	input.SetAmount(amt + fee)

	dataSig, err := trans.MarshalBinarySig()
	if err != nil {
		return trans, err
	}
	sig := factoid.NewSingleSignatureBlock(inSec.Bytes(), dataSig)
	trans.SetSignatureBlock(0, sig)

	return trans, nil

}

// create a transaction to buy Entry Credits
// this adds the EC fee on top of the input amount
func ComposeEcTransaction(inSec interfaces.IHash, outEC interfaces.IHash, timeInMilliseconds *primitives.Timestamp, amt uint64, ecPrice uint64) (*factoid.Transaction, error) {
	var sec [64]byte
	copy(sec[:32], inSec.Bytes())   // pass 32 byte key in a 64 byte field for the crypto library
	pub := ed.GetPublicKey(&sec)    // get the public key for our FCT source address
	rcd := factoid.NewRCD_1(pub[:]) // build the an RCD "redeem condition data structure"

	inAdd, err := rcd.GetAddress()
	if err != nil {
		panic(err)
	}

	outAdd := factoid.NewAddress(outEC.Bytes())

	trans := new(factoid.Transaction)
	trans.AddInput(inAdd, amt)
	trans.AddECOutput(outAdd, amt)
	trans.AddRCD(rcd)
	trans.AddAuthorization(rcd)
	trans.SetTimestamp(timeInMilliseconds)

	fee, err := trans.CalculateFee(ecPrice)

	if err != nil {
		return nil, err
	}

	input, err := trans.GetInput(0)
	if err != nil {
		return nil, err
	}
	input.SetAmount(amt + fee)

	dataSig, err := trans.MarshalBinarySig()
	if err != nil {
		return nil, err
	}
	sig := factoid.NewSingleSignatureBlock(inSec.Bytes(), dataSig)
	trans.SetSignatureBlock(0, sig)

	return trans, nil
}
