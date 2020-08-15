package blockgen

import (
	"math/rand"

	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/state"
	log "github.com/sirupsen/logrus"
)

type Range struct {
	Min int
	Max int
}

func (r Range) Amount() int {
	if r.Max < r.Min {
		return 0
	}
	if r.Max == r.Min {
		return r.Max
	}
	return rand.Intn(r.Max-r.Min) + r.Min
}

type EntryGeneratorConfig struct {
	EntriesPerEBlock Range
	EntrySize        Range
	EblocksPerHeight Range

	// MultiThread Stuff
	Multithreaded   bool
	ThreadpoolCount int
}

func NewDefaultEntryGeneratorConfig() *EntryGeneratorConfig {
	e := new(EntryGeneratorConfig)
	e.EntriesPerEBlock = Range{10, 100}
	e.EntrySize = Range{100, 1000}
	e.EblocksPerHeight = Range{5, 10}
	e.Multithreaded = true
	e.ThreadpoolCount = 8
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

func NewFullEntryGenerator(ecKey, fKey primitives.PrivateKey, config DBGeneratorConfig) *FullEntryGenerator {
	f := new(FullEntryGenerator)
	// There can other entry generators
	switch config.EntryGenerator {
	case "record":
		f.IEntryGenerator = NewRecordEntryGenerator(ecKey, config.EntryGenConfig)
	case "increment", "incr":
		f.IEntryGenerator = NewIncrementEntryGenerator(ecKey, config.EntryGenConfig)
	default:
		f.IEntryGenerator = NewRandomEntryGenerator(ecKey, config.EntryGenConfig)
	}
	log.Infof("EntryGen found : '%s'. Using %s", config.EntryGenerator, f.IEntryGenerator.Name())
	f.FKey = fKey
	f.Config = config.EntryGenConfig

	return f
}

func (f *FullEntryGenerator) NewBlockSet(prev *state.DBState, newtime interfaces.Timestamp) (*state.DBState, error) {
	newDBState := new(state.DBState)
	// Need all the entries and commits
	// Then we need to build an ECBlock, and a factoid transaction to fund these entries

	//  Step 1: Get the entries
	eblocks, entries, commits, totalcost := f.IEntryGenerator.AllEntries(prev.DirectoryBlock.GetDatabaseHeight()+1, newtime)
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
	ecb.GetHeader().SetDBHeight(prev.EntryCreditBlock.GetDatabaseHeight() + 1)

	hash, err := prev.EntryCreditBlock.HeaderHash()
	if err != nil {
		panic(err.Error())
	}
	ecb.GetHeader().SetPrevHeaderHash(hash)

	hash, err = prev.EntryCreditBlock.GetFullHash()
	if err != nil {
		panic(err.Error())
	}
	ecb.GetHeader().SetPrevFullHash(hash)

	// Step3: FactoidBlock with funds
	fb := factoid.NewFBlock(prev.FactoidBlock)
	coinbase := new(factoid.Transaction)
	coinbase.MilliTimestamp = newtime.GetTimeMilliUInt64()

	fb.AddCoinbase(coinbase)

	// This will cover the ec needed for all our commits
	ect, err := BuyEC(f.FKey, f.IEntryGenerator.GetECKey().Pub, uint64(totalcost), prev.FactoidBlock.GetExchRate(), newtime)
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

// BuyEC returns a factoid transaction to cover the ec amount
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
	NewChainHead() *entryBlock.Entry
	NewEntry(chain interfaces.IHash) *entryBlock.Entry

	GetECKey() primitives.PrivateKey
	Name() string
}
