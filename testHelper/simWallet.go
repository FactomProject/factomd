package testHelper

// test helpers for Transaction & entry creations

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	"text/template"
	"time"
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

// This account has a balance from inital coinbase
func GetBankAccount() *testAccount {
	return AccountFromFctSecret("Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK")
}

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

func ComposeChainReveal(pkey *primitives.PrivateKey) (*messages.RevealEntryMsg, error) {
	//e := entryCreditBlock.NewCommitChain()

	// FIXME
	m := new(messages.RevealEntryMsg)
	return m, nil
}

func WaitForAnyDeposit(s *state.State, ecPub string) int64 {
	return WaitForEcBalance(s, ecPub, 1)
}

func WaitForZero(s *state.State, ecPub string) int64 {
	fmt.Println("Waiting for Zero Balance")
	return WaitForEcBalance(s, ecPub, 0)
}

func WaitForEmptyHolding(s *state.State, msg string) time.Time {
	t := time.Now()
	s.LogPrintf(logName, "WaitForEmptyHolding %v", msg)

	for len(s.Holding) > 0 {
		time.Sleep(time.Millisecond * 10)
	}

	t = time.Now()
	s.LogPrintf(logName, "EmptyHolding %v", msg)

	return t
}

func WaitForEcBalance(s *state.State, ecPub string, target int64) int64 {

	s.LogPrintf(logName, "WaitForBalance%v:  %v", target, ecPub)

	for {
		bal := engine.GetBalanceEC(s, ecPub)
		time.Sleep(time.Millisecond * 200)
		//fmt.Printf("WaitForBalance: %v => %v\n", ecPub, bal)

		if (target == 0 && bal == 0) || (target > 0 && bal >= target) {
			s.LogPrintf(logName, "FoundBalance%v: %v", target, bal)
			return bal
		}
	}
}

func WatchMessageLists() *time.Ticker {

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			for _, n := range engine.GetFnodes() {

				f := n.State

				list := []interface{}{
					len(f.Holding),
					len(f.Acks),
					len(f.MsgQueue()),
					f.InMsgQueue().Length(),
					f.APIQueue().Length(),
					len(f.AckQueue()),
					len(f.TimerMsgQueue()),
				}

				f.LogPrintf(logName, "LIST_SIZES Holding: %v, Acks: %v, MsgQueue: %v, InMsgQueue: %v, APIQueue: %v, AckQueue: %v, TimerMsg: %v ", list...)
			}

		}
	}()

	return ticker
}

