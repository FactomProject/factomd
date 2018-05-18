package blockgen

import (
	"math/rand"

	"time"

	"bytes"
	"encoding/binary"

	"fmt"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
)

type Range struct {
	Min int
	Max int
}

func (r Range) Amount() int {
	if r.Max == r.Min {
		return r.Max
	}
	return rand.Intn(r.Max) + r.Min
}

type EntryGeneratorConfig struct {
	EntriesPerBlock  Range
	EntrySize        Range
	EblocksPerHeight Range
}

func NewDefaultEntryGeneratorConfig() *EntryGeneratorConfig {
	e := new(EntryGeneratorConfig)
	e.EntriesPerBlock = Range{10, 100}
	e.EntrySize = Range{100, 1000}
	e.EblocksPerHeight = Range{5, 10}

	return e
}

type IFullEntryGenerator interface {
	NewBlockSet(dbs *state.DBState, newtime interfaces.Timestamp) (*state.DBState, error)
}

// Generates ECBlock, EBlocks, Entries, and Factoid transactions
type FullEntryGenerator struct {
	FKey primitives.PrivateKey
	IEntryGenerator
	Config EntryGeneratorConfig
}

func NewFullEntryGenerator(ecKey, fKey primitives.PrivateKey, config EntryGeneratorConfig) *FullEntryGenerator {
	f := new(FullEntryGenerator)
	f.IEntryGenerator = NewRandomEntryGenerator(ecKey, config)
	f.FKey = fKey
	f.Config = config

	return f
}

func (f *FullEntryGenerator) NewBlockSet(dbs *state.DBState, newtime interfaces.Timestamp) (*state.DBState, error) {
	newDBState := new(state.DBState)
	// Need all the entries and commits
	// Then we need to build an ECBlock, and a factoid transaction to fund these entries

	//  Step 1: Get the entries
	eblocks, entries, commits, totalcost := f.IEntryGenerator.AllEntries(dbs.DirectoryBlock.GetDatabaseHeight()+1, newtime)
	newDBState.EntryBlocks = make([]interfaces.IEntryBlock, len(eblocks))
	for i, eb := range eblocks {
		newDBState.EntryBlocks[i] = eb
	}

	newDBState.Entries = make([]interfaces.IEBEntry, len(entries))
	for i, e := range entries {
		newDBState.Entries[i] = e
	}

	// Step 2: ECBlock
	ecb := entryCreditBlock.NewECBlock()
	for _, c := range commits {
		ecb.GetBody().AddEntry(c)
	}

	// Step3: FactoidBlock with funds
	fb := factoid.NewFBlock(dbs.FactoidBlock)
	coinbase := new(factoid.Transaction)
	coinbase.MilliTimestamp = newtime.GetTimeMilliUInt64()
	fmt.Println(coinbase.MilliTimestamp)
	fb.AddCoinbase(coinbase)

	ect, err := BuyEC(f.FKey, f.IEntryGenerator.GetECKey().Pub, uint64(totalcost), dbs.FactoidBlock.GetExchRate(), newtime)
	if err != nil {
		return nil, err
	}
	ect.BlockHeight = fb.GetDatabaseHeight()
	ect.GetTxID()
	err = fb.AddTransaction(ect)
	if err != nil {
		return nil, err
	}

	// Assemble
	newDBState.EntryCreditBlock = ecb

	newDBState.EntryBlocks = make([]interfaces.IEntryBlock, len(eblocks))
	for i, eb := range eblocks {
		newDBState.EntryBlocks[i] = eb
	}

	newDBState.Entries = make([]interfaces.IEBEntry, len(entries))
	for i, e := range entries {
		newDBState.Entries[i] = e
	}

	newDBState.FactoidBlock = fb

	// hashes
	newDBState.FBHash = fb.DatabasePrimaryIndex()
	newDBState.ECHash = ecb.DatabasePrimaryIndex()

	return newDBState, nil
}

func BuyEC(from primitives.PrivateKey, to *primitives.PublicKey, ecamount uint64, ecrate uint64, time interfaces.Timestamp) (*factoid.Transaction, error) {
	trans := new(factoid.Transaction)
	pub := from.Pub.Fixed()
	inHash := primitives.Sha(primitives.Sha(append([]byte{0x01}, pub[:]...)).Bytes())
	rcd := factoid.NewRCD_1(pub[:])

	amt := uint64(ecamount * ecrate)

	trans.AddInput(inHash, amt)
	trans.AddECOutput(factoid.NewAddress(to[:]), amt)
	trans.AddRCD(rcd)
	trans.SetTimestamp(time)
	fee, err := trans.CalculateFee(ecrate)
	if err != nil {
		return nil, err
	}
	input, err := trans.GetInput(0)
	if err != nil {
		return nil, err
	}
	input.SetAmount(input.GetAmount() + fee)
	dataSig, err := trans.MarshalBinarySig()
	if err != nil {
		return nil, err
	}
	sig := factoid.NewSingleSignatureBlock(from.Key[:], dataSig)
	trans.SetSignatureBlock(0, sig)
	return trans, nil
}

type IEntryGenerator interface {
	AllEntries(height uint32, time interfaces.Timestamp) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int)
	NewEblock(height uint32, time interfaces.Timestamp) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int)
	NewEntry(chain interfaces.IHash) *entryBlock.Entry

	GetECKey() primitives.PrivateKey
}

// RandomEntryGenerator generates random entries between 0-10kbish
type RandomEntryGenerator struct {
	ECKey  primitives.PrivateKey
	Config EntryGeneratorConfig
}

func NewRandomEntryGenerator(ecKey primitives.PrivateKey, config EntryGeneratorConfig) *RandomEntryGenerator {
	r := new(RandomEntryGenerator)
	r.ECKey = ecKey
	r.Config = config

	return r
}

func (r *RandomEntryGenerator) GetECKey() primitives.PrivateKey {
	return r.ECKey
}

func (r *RandomEntryGenerator) AllEntries(height uint32, time interfaces.Timestamp) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	eblocks := make([]*entryBlock.EBlock, 0)
	commits := make([]*entryCreditBlock.CommitEntry, 0)
	entries := make([]*entryBlock.Entry, 0)
	totalCost := 0

	for i := 0; i < r.Config.EblocksPerHeight.Amount(); i++ {
		neb, nes, necs, t := r.NewEblock(height, time)
		eblocks = append(eblocks, neb)
		entries = append(entries, nes...)
		commits = append(commits, necs...)
		totalCost += t
	}
	return eblocks, entries, commits, totalCost
}

func (r *RandomEntryGenerator) NewEblock(height uint32, time interfaces.Timestamp) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	commits := make([]*entryCreditBlock.CommitEntry, 0)
	entries := make([]*entryBlock.Entry, 0)
	totalCost := 0

	head := r.NewEntry(primitives.NewZeroHash())
	// First one needs an extid
	head.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{random.RandByteSliceOfLen(10)}, primitives.ByteSlice{random.RandByteSliceOfLen(10)}}
	head.ChainID = head.GetChainID()
	commit := r.newCommit(head, time)
	commit.Credits += 10
	totalCost += int(commit.Credits)
	commit = r.signCommit(commit)

	eb := entryBlock.NewEBlock()
	eb.Header.SetChainID(head.ChainID)
	eb.Header.SetDBHeight(height)
	eb.AddEBEntry(head)

	entries = append(entries, head)
	commits = append(commits, commit)

	// now add the other entries
	for i := 0; i < r.Config.EntriesPerBlock.Amount(); i++ {
		ent := r.NewEntry(head.ChainID)
		commit := r.newCommit(ent, time)
		commit = r.signCommit(commit)
		totalCost += int(commit.Credits)
		eb.AddEBEntry(ent)

		entries = append(entries, ent)
		commits = append(commits, commit)
	}

	return eb, entries, commits, totalCost
}

func (r *RandomEntryGenerator) NewEntry(chain interfaces.IHash) *entryBlock.Entry {
	conf := r.Config
	bytes := rand.Intn(conf.EntrySize.Max) + conf.EntrySize.Max

	ent := entryBlock.NewEntry()
	ent.Content = primitives.ByteSlice{random.RandByteSliceOfLen(bytes)}
	ent.ChainID = chain
	return ent
}

func (r *RandomEntryGenerator) signCommit(entry *entryCreditBlock.CommitEntry) *entryCreditBlock.CommitEntry {
	entry.Sign(r.ECKey.Key[:])
	return entry
}

func (r *RandomEntryGenerator) newCommit(e *entryBlock.Entry, time interfaces.Timestamp) *entryCreditBlock.CommitEntry {
	commit := entryCreditBlock.NewCommitEntry()
	commit.EntryHash = e.GetHash()
	d, _ := e.MarshalBinary()
	commit.Credits, _ = util.EntryCost(d)
	var t primitives.ByteSlice6
	copy(t[:], milliTime(time.GetTimeSeconds())[:])
	commit.MilliTime = &t
	return commit
}

// milliTime returns a 6 byte slice representing the unix time in milliseconds
func milliTime(unix int64) (r []byte) {
	buf := new(bytes.Buffer)
	t := time.Unix(unix, 0).UnixNano()
	m := t / 1e6
	binary.Write(buf, binary.BigEndian, m)
	return buf.Bytes()[2:]
}
