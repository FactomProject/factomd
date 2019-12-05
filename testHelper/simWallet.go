package testHelper

// test helpers for Transaction & entry creations

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"text/template"
	"time"

	"github.com/FactomProject/factomd/simulation"

	"github.com/FactomProject/factomd/fnode"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
)

// struct to generate FCT or EC addresses
// from the same private key
type testAccount struct {
	Priv *primitives.PrivateKey
}

var logName string = "simTest"

func (d *testAccount) FctPriv() string {
	x, _ := primitives.PrivateKeyStringToHumanReadableFactoidPrivateKey(d.Priv.PrivateKeyString())
	return x
}

func (d *testAccount) FctPub() string {
	s, _ := factoid.PublicKeyStringToFactoidAddressString(d.Priv.PublicKeyString())
	return s
}

func (d *testAccount) EcPub() string {
	s, _ := factoid.PublicKeyStringToECAddressString(d.Priv.PublicKeyString())
	return s
}

func (d *testAccount) EcPriv() string {
	s, _ := primitives.PrivateKeyStringToHumanReadableECPrivateKey(d.Priv.PrivateKeyString())
	return s
}

func (d *testAccount) FctPrivHash() interfaces.IHash {
	a := primitives.ConvertUserStrToAddress(d.FctPriv())
	x, _ := primitives.HexToHash(hex.EncodeToString(a))
	return x
}

func (d *testAccount) FctAddr() interfaces.IHash {
	a := primitives.ConvertUserStrToAddress(d.FctPub())
	x, _ := primitives.HexToHash(hex.EncodeToString(a))
	return x
}

func (d *testAccount) EcPrivHash() interfaces.IHash {
	a := primitives.ConvertUserStrToAddress(d.EcPriv())
	x, _ := primitives.HexToHash(hex.EncodeToString(a))
	return x
}

func (d *testAccount) EcAddr() interfaces.IHash {
	a := primitives.ConvertUserStrToAddress(d.EcPub())
	x, _ := primitives.HexToHash(hex.EncodeToString(a))
	return x
}

// buy EC from coinbase 'bank'
func (d *testAccount) FundEC(amt uint64) {
	state0 := fnode.GetFnodes()[0].State
	simulation.FundECWallet(state0, GetBankAccount().FctPrivHash(), d.EcAddr(), uint64(amt)*state0.GetFactoshisPerEC())
}

// buy EC from account
func (d *testAccount) ConvertEC(amt uint64) {
	state0 := fnode.GetFnodes()[0].State
	simulation.FundECWallet(state0, d.FctPrivHash(), d.EcAddr(), uint64(amt)*state0.GetFactoshisPerEC())
}

// get FCT from coinbase 'bank'
func (d *testAccount) FundFCT(amt uint64) {
	state0 := fnode.GetFnodes()[0].State
	_, err := simulation.SendTxn(state0, uint64(amt), GetBankAccount().FctPriv(), d.FctPub(), state0.GetFactoshisPerEC())
	if err != nil {
		panic(err)
	}
}

// transfer FCT from account
func (d *testAccount) SendFCT(a *testAccount, amt uint64) {
	state0 := fnode.GetFnodes()[0].State
	simulation.SendTxn(state0, uint64(amt), d.FctPriv(), a.FctPub(), state0.GetFactoshisPerEC())
}

// check EC balance
func (d *testAccount) GetECBalance(args ...*fnode.FactomNode) int64 {
	var state0 *state.State

	switch len(args) {
	case 0:
		state0 = fnode.Get(0).State // default to the first one
	case 1:
		state0 = args[0].State
	default:
		panic("expect 0 or 1 fnode param")
	}

	return simulation.GetBalanceEC(state0, d.EcPub())
}

var testFormat string = `
FCT
  FctPriv: {{ .FctPriv }}
  FctPub: {{ .FctPub }}
  FctPrivHash: {{ .FctPrivHash }}
  FctAddr: {{ .FctAddr }}
EC
  EcPriv: {{ .EcPriv }}
  EcPub: {{ .EcPub }}
  EcPrivHash: {{ .EcPrivHash }}
  EcAddr: {{ .EcAddr }}
`
var testTemplate *template.Template = template.Must(
	template.New("").Parse(testFormat),
)

func (d *testAccount) String() string {
	b := &bytes.Buffer{}
	testTemplate.Execute(b, d)
	return b.String()
}

func AccountFromFctSecret(s string) *testAccount {
	d := new(testAccount)
	h, _ := primitives.HumanReadableFactoidPrivateKeyToPrivateKey(s)
	d.Priv = primitives.NewPrivateKeyFromHexBytes(h)
	return d
}

// This account has a balance from initial coinbase
func GetBankAccount() *testAccount {
	return AccountFromFctSecret("Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK")
}

// build addresses from random key
func GetRandomAccount() *testAccount {
	d := new(testAccount)
	d.Priv = primitives.RandomPrivateKey()
	return d
}

// KLUDGE duplicates code from: factom lib
// TODO: refactor factom package to export these functions
func milliTime() (r []byte) {
	buf := new(bytes.Buffer)
	t := time.Now().UnixNano()
	m := t / 1e6
	binary.Write(buf, binary.BigEndian, m)
	return buf.Bytes()[2:]
}

// KLUDGE duplicates code from: factom.ComposeEntryCommit()
// TODO: refactor factom package to export these functions
func commitEntryMsg(addr *factom.ECAddress, e *factom.Entry) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	// 1 byte version
	buf.Write([]byte{0})

	// 6 byte milliTimestamp (truncated unix time)
	buf.Write(milliTime())

	// 32 byte Entry Hash
	buf.Write(e.Hash())

	// 1 byte number of entry credits to pay
	if c, err := factom.EntryCost(e); err != nil {
		return nil, err
	} else {
		buf.WriteByte(byte(c))
	}

	// 32 byte Entry Credit Address Public Key + 64 byte Signature
	sig := addr.Sign(buf.Bytes())
	buf.Write(addr.PubBytes())
	buf.Write(sig[:])

	return buf, nil
}

// KLUDGE: copy from factom lib
// shad Double Sha256 Hash; sha256(sha256(data))
func shad(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

// KLUDGE copy from factom
func composeChainCommitMsg(c *factom.Chain, ec *factom.ECAddress) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	// 1 byte version
	buf.Write([]byte{0})

	// 6 byte milliTimestamp
	buf.Write(milliTime())

	e := c.FirstEntry
	// 32 byte ChainID Hash
	if p, err := hex.DecodeString(c.ChainID); err != nil {
		return nil, err
	} else {
		// double sha256 hash of ChainID
		buf.Write(shad(p))
	}

	// 32 byte Weld; sha256(sha256(EntryHash + ChainID))
	if cid, err := hex.DecodeString(c.ChainID); err != nil {
		return nil, err
	} else {
		s := append(e.Hash(), cid...)
		buf.Write(shad(s))
	}

	// 32 byte Entry Hash of the First Entry
	buf.Write(e.Hash())

	// 1 byte number of Entry Credits to pay
	if d, err := factom.EntryCost(e); err != nil {
		return nil, err
	} else {
		buf.WriteByte(byte(d + 10))
	}

	// 32 byte Entry Credit Address Public Key + 64 byte Signature
	sig := ec.Sign(buf.Bytes())
	buf.Write(ec.PubBytes())
	buf.Write(sig[:])

	return buf, nil
}

func PrivateKeyToECAddress(key *primitives.PrivateKey) *factom.ECAddress {
	// KLUDGE is there a better way to do this?
	ecPub, _ := factoid.PublicKeyStringToECAddress(key.PublicKeyString())
	addr := factom.ECAddress{&[32]byte{}, &[64]byte{}}
	copy(addr.Pub[:], ecPub.Bytes())
	copy(addr.Sec[:], key.Key[:])
	return &addr
}

func ComposeCommitEntryMsg(pkey *primitives.PrivateKey, e factom.Entry) (*messages.CommitEntryMsg, error) {
	msg, err := commitEntryMsg(PrivateKeyToECAddress(pkey), &e)

	commit := entryCreditBlock.NewCommitEntry()
	commit.UnmarshalBinaryData(msg.Bytes())

	m := new(messages.CommitEntryMsg)
	m.CommitEntry = commit
	m.SetValid()
	return m, err
}

func ComposeRevealEntryMsg(pkey *primitives.PrivateKey, e *factom.Entry) (*messages.RevealEntryMsg, error) {
	entry := entryBlock.NewEntry()
	entry.Content = primitives.ByteSlice{Bytes: e.Content}

	id, _ := primitives.HexToHash(e.ChainID)
	entry.ChainID = id

	for _, extID := range e.ExtIDs {
		entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: extID})
	}

	m := new(messages.RevealEntryMsg)
	m.Entry = entry
	m.Timestamp = primitives.NewTimestampNow()
	m.SetValid()

	return m, nil
}

func ComposeChainCommit(pkey *primitives.PrivateKey, c *factom.Chain) (*messages.CommitChainMsg, error) {
	msg, _ := composeChainCommitMsg(c, PrivateKeyToECAddress(pkey))
	e := entryCreditBlock.NewCommitChain()
	_, err := e.UnmarshalBinaryData(msg.Bytes())
	if err != nil {
		return nil, err
	}

	m := new(messages.CommitChainMsg)
	m.CommitChain = e
	m.SetValid()
	return m, nil
}

// wait for non-zero EC balance
func WaitForAnyEcBalance(s *state.State, ecPub string) int64 {
	s.LogPrintf(logName, "WaitForAnyEcBalance %v", ecPub)
	return WaitForEcBalanceOver(s, ecPub, 0)
}

// wait for non-zero FCT balance
func WaitForAnyFctBalance(s *state.State, fctPub string) int64 {
	s.LogPrintf(logName, "WaitForAnyFctBalance %v", fctPub)
	return WaitForFctBalanceOver(s, fctPub, 0)
}

// wait for exactly Zero EC balance
// REVIEW: should we ditch this?
func WaitForZeroEC(s *state.State, ecPub string) int64 {
	s.LogPrintf(logName, "WaitingForZeroEcBalance")
	return WaitForEcBalanceUnder(s, ecPub, 1)
}

const balanceWaitInterval = time.Millisecond * 20

// loop until balance is < target
func WaitForEcBalanceUnder(s *state.State, ecPub string, target int64) int64 {

	s.LogPrintf(logName, "WaitForEcBalanceUnder%v:  %v", target, ecPub)

	for {
		bal := simulation.GetBalanceEC(s, ecPub)
		time.Sleep(balanceWaitInterval)

		if bal < target {
			s.LogPrintf(logName, "FoundEcBalanceUnder%v: %v", target, bal)
			return bal
		}
	}
}

// loop until balance is >= target
func WaitForEcBalanceOver(s *state.State, ecPub string, target int64) int64 {

	s.LogPrintf(logName, "WaitForEcBalanceOver%v:  %v", target, ecPub)

	for {
		bal := simulation.GetBalanceEC(s, ecPub)
		time.Sleep(balanceWaitInterval)

		if bal > target {
			s.LogPrintf(logName, "FoundEcBalancerOver%v: %v", target, bal)
			return bal
		}
	}
}

// loop until balance is >= target
func WaitForFctBalanceUnder(s *state.State, fctPub string, target int64) int64 {

	s.LogPrintf(logName, "WaitForFctBalanceUnder%v:  %v", target, fctPub)

	for {
		bal := simulation.GetBalance(s, fctPub)
		time.Sleep(balanceWaitInterval)

		if bal < target {
			s.LogPrintf(logName, "FoundFctBalanceUnder%v: %v", target, bal)
			return bal
		}
	}
}

// loop until balance is <= target
func WaitForFctBalanceOver(s *state.State, fctPub string, target int64) int64 {

	s.LogPrintf(logName, "WaitForFctBalanceOver%v:  %v", target, fctPub)

	for {
		bal := simulation.GetBalance(s, fctPub)
		time.Sleep(balanceWaitInterval)

		if bal > target {
			s.LogPrintf(logName, "FoundMaxFctBalanceOver%v: %v", target, bal)
			return bal
		}
	}
}
