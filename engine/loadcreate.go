package engine

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"

	"crypto/sha256"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/util/atomic"
)

type LoadGenerator struct {
	ECKey     *primitives.PrivateKey // Entry Credit private key
	ToSend    int                    // How much to send
	PerSecond atomic.AtomicInt       // How much per second
	stop      chan bool              // Stop the go routine
	running   atomic.AtomicBool      // We are running
	tight     atomic.AtomicBool      // Only allocate ECs as needed (more EC purchases)
	txoffset  int64                  // Offset to be added to the timestamp of created tx to test time limits.
	state     *state.State           // Access to logging
}

// NewLoadGenerator makes a new load generator. The state is used for funding the transaction
func NewLoadGenerator(s *state.State) *LoadGenerator {
	lg := new(LoadGenerator)
	lg.ECKey, _ = primitives.NewPrivateKeyFromHex(ecSec)
	lg.stop = make(chan bool, 5)
	lg.state = s

	return lg
}

func (lg *LoadGenerator) Run() {
	if lg.running.Load() {
		return
	}

	lg.running.Store(true)
	//FundWallet(fnodes[wsapiNode].State, 15000e8)

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
		top := lg.ToSend / 10      // ToSend is in tenths so get the integer part
		if top == 0 {
			top = 1
		}
		lg.ToSend = lg.ToSend % 10 // save an fractional part for next iteration
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
			s := fnodes[wsapiNode].State
			s.APIQueue().Enqueue(c)
			s.APIQueue().Enqueue(r)
			time.Sleep(time.Duration(800/top) * time.Millisecond) // spread the load out over 800ms + overhead
		}
	}
}

func (lg *LoadGenerator) Stop() {
	lg.stop <- true
}

func RandomEntry() *entryBlock.Entry {
	entry := entryBlock.NewEntry()
	entry.Content = primitives.ByteSlice{random.RandByteSliceOfLen(rand.Intn(4000) + 128)}
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

var cnt int
var goingUp bool
var limitBuys = true // We limit buys only after one attempted purchase, so people can fund identities in testing

func (lg *LoadGenerator) KeepUsFunded() {

	s := fnodes[wsapiNode].State

	var level int64

	buys := 0
	totalBought := 0
	for i := 0; ; i++ {

		ts := "false"
		if lg.tight.Load() {
			ts = "true"
		}

		//EC3Eh7yQKShgjkUSFrPbnQpboykCzf4kw9QHxi47GGz5P2k3dbab is EC address
		if lg.PerSecond == 0 && limitBuys {
			if i%100 == 0 {
				// Log our occasional realization that we have nothing to do.
				outEC, _ := primitives.HexToHash("c23ae8eec2beb181a0da926bd2344e988149fbe839fbc7489f2096e7d6110243")
				outAdd := factoid.NewAddress(outEC.Bytes())
				ecBal := s.GetE(true, outAdd.Fixed())

				lg.state.LogPrintf("loadgenerator", "Tight %7s Total TX %6d for a total of %8d entry credits balance %d.",
					ts, buys, totalBought, ecBal)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if !limitBuys {
			level = 200 // Only do this once, after that look for requests for load to drive EC buys.
		} else if lg.tight.Load() {
			level = 10
		} else {
			level = 10000
		}

		outEC, _ := primitives.HexToHash("c23ae8eec2beb181a0da926bd2344e988149fbe839fbc7489f2096e7d6110243")
		outAdd := factoid.NewAddress(outEC.Bytes())
		ecBal := s.GetE(true, outAdd.Fixed())
		ecPrice := s.GetFactoshisPerEC()

		if ecBal < level {
			buys++
			need := level - ecBal + level*2
			totalBought += int(need)
			FundWalletTOFF(s, lg.txoffset, uint64(need)*ecPrice)
		} else {
			limitBuys = true
		}

		if i%5 == 0 {
			lg.state.LogPrintf("loadgenerator", "Tight %7s Total TX %6d for a total of %8d entry credits balance %d.",
				ts, buys, totalBought, ecBal)
		}
		time.Sleep(100 * time.Millisecond)
	}
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
	copy(b6[:], milliTime(lg.txoffset)[:])
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
	copy(b6[:], milliTime(lg.txoffset)[:])
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
func milliTime(offset int64) (r []byte) {
	buf := new(bytes.Buffer)
	t := time.Now().UnixNano()
	m := t/1e6 + offset
	binary.Write(buf, binary.BigEndian, m)
	return buf.Bytes()[2:]
}
