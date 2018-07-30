package engine

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"

	"crypto/sha256"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/util/atomic"
)

type LoadGenerator struct {
	ECKey *primitives.PrivateKey

	ToSend    int
	PerSecond atomic.AtomicInt
	stop      chan bool

	running atomic.AtomicBool
}

// NewLoadGenerator makes a new load generator. The state is used for funding the transaction
func NewLoadGenerator() *LoadGenerator {
	lg := new(LoadGenerator)
	lg.ECKey, _ = primitives.NewPrivateKeyFromHex(ecSec)
	lg.stop = make(chan bool, 5)

	return lg
}

func (lg *LoadGenerator) Run() {
	if lg.running.Load() {
		return
	}
	lg.running.Store(true)
	fundWallet(fnodes[wsapiNode].State, 15000e8)

	// Every second add the per second amount
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		select {
		case <-lg.stop:
			lg.running.Store(false)
			return
		default:

		}
		addSend := lg.PerSecond.Load()
		lg.ToSend += addSend
		top := lg.ToSend / 10
		lg.ToSend = lg.ToSend % 10
		if addSend == 0 {
			lg.running.Store(false)
			return
		}
		var chain interfaces.IHash = nil
		for i := 0; i < top; i++ {
			var c interfaces.IMsg
			e := RandomEntry()
			if chain == nil {
				c = lg.NewCommitChain(e)
				chain = e.ChainID
			} else {
				e.ChainID = chain
				c = lg.NewCommitEntry(e)
			}
			r := lg.NewRevealEntry(e)

			fnodes[wsapiNode].State.APIQueue().Enqueue(c)
			fnodes[wsapiNode].State.APIQueue().Enqueue(r)
		}
	}
}

func (lg *LoadGenerator) Stop() {
	lg.stop <- true
}

func RandomEntry() *entryBlock.Entry {
	entry := entryBlock.NewEntry()
	entry.Content = primitives.ByteSlice{random.RandByteSliceOfLen(rand.Intn(4000))}
	entry.ExtIDs = make([]primitives.ByteSlice, rand.Intn(4)+1)
	raw := make([][]byte, len(entry.ExtIDs))
	for i := range entry.ExtIDs {
		entry.ExtIDs[i] = primitives.ByteSlice{random.RandByteSliceOfLen(rand.Intn(300))}
		raw[i] = entry.ExtIDs[i].Bytes
	}

	sum := sha256.New()
	for _, v := range entry.ExtIDs {
		x := sha256.Sum256(v.Bytes)
		sum.Write(x[:])
	}
	originalHash := sum.Sum(nil)
	checkHash := primitives.Shad(originalHash)

	entry.ChainID = checkHash
	return entry
}

func (lg *LoadGenerator) NewRevealEntry(entry *entryBlock.Entry) *messages.RevealEntryMsg {
	msg := messages.NewRevealEntryMsg()
	msg.Entry = entry
	msg.Timestamp = primitives.NewTimestampNow()

	return msg
}

func (lg *LoadGenerator) NewCommitChain(entry *entryBlock.Entry) *messages.CommitChainMsg {
	msg := new(messages.CommitChainMsg)

	// form commit
	commit := entryCreditBlock.NewCommitChain()
	data, _ := entry.MarshalBinary()
	commit.Credits, _ = util.EntryCost(data)
	commit.Credits += 10
	commit.EntryHash = entry.GetHash()
	var b6 primitives.ByteSlice6
	copy(b6[:], milliTime()[:])
	commit.MilliTime = &b6
	var b32 primitives.ByteSlice32
	copy(b32[:], lg.ECKey.Pub[:])
	commit.ECPubKey = &b32
	commit.Weld = entry.GetWeldHash()
	commit.ChainIDHash = entry.ChainID

	commit.Sign(lg.ECKey.Key[:])

	// form msg
	msg.CommitChain = commit
	msg.Sign(lg.ECKey)

	return msg
}

func (lg *LoadGenerator) NewCommitEntry(entry *entryBlock.Entry) *messages.CommitEntryMsg {
	msg := messages.NewCommitEntryMsg()

	// form commit
	commit := entryCreditBlock.NewCommitEntry()
	data, _ := entry.MarshalBinary()
	commit.Credits, _ = util.EntryCost(data)
	commit.EntryHash = entry.GetHash()
	var b6 primitives.ByteSlice6
	copy(b6[:], milliTime()[:])
	commit.MilliTime = &b6
	var b32 primitives.ByteSlice32
	copy(b32[:], lg.ECKey.Pub[:])
	commit.ECPubKey = &b32
	commit.Sign(lg.ECKey.Key[:])

	// form msg
	msg.CommitEntry = commit
	msg.Sign(lg.ECKey)

	return msg
}

// milliTime returns a 6 byte slice representing the unix time in milliseconds
func milliTime() (r []byte) {
	buf := new(bytes.Buffer)
	t := time.Now().UnixNano()
	m := t / 1e6
	binary.Write(buf, binary.BigEndian, m)
	return buf.Bytes()[2:]
}
