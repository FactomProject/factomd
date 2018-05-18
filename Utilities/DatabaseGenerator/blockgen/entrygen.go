package blockgen

import (
	"math/rand"

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

type IFullEntryGenerator interface {
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

func (f *FullEntryGenerator) NewBlockSet(dbs *state.DBState) {
	newDBState := new(state.DBState)
	// Need all the entries and commits
	// Then we need to build an ECBlock, and a factoid transaction to fund these entries

	//  Step 1: Get the entries
	eblocks, entries, commits, totalcost := f.IEntryGenerator.AllEntries()
	newDBState.EntryBlocks = eblocks
	newDBState.Entries = entries

	// Step 2: ECBlock
	ecb := entryCreditBlock.NewECBlock()
	for _, c := range commits {
		ecb.GetBody().AddEntry(c)
	}

	// Step3: FactoidBlock with funds

}

func BuyEC(from primitives.PrivateKey, to primitives.PublicKey, amt int64, ecrate uint64) (*factoid.Transaction, error) {
	trans := new(factoid.Transaction)
	pub := from.Pub.Fixed()
	inHash := primitives.Sha(append([]byte{0x01}, pub[:]...))
	rcd := factoid.NewRCD_1(pub[:])

	trans.AddInput(inHash, 0)
	trans.AddECOutput(factoid.NewAddress(to[:]), 0)
	trans.AddRCD(rcd)
	trans.SetTimestamp(primitives.NewTimestampNow())
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
	sig := factoid.NewSingleSignatureBlock(from[:], dataSig)
	trans.SetSignatureBlock(0, sig)
	return trans, nil

	/*
			func fundWallet(st *state.State, amt uint64) error {
			inSec, _ := primitives.HexToHash("FB3B471B1DCDADFEB856BD0B02D8BF49ACE0EDD372A3D9F2A95B78EC12A324D6")
			outEC, _ := primitives.HexToHash("c23ae8eec2beb181a0da926bd2344e988149fbe839fbc7489f2096e7d6110243")
			inHash, _ := primitives.HexToHash("646F3E8750C550E4582ECA5047546FFEF89C13A175985E320232BACAC81CC428")
			var sec [64]byte
			copy(sec[:32], inSec.Bytes())

			pub := ed.GetPublicKey(&sec)
			//inRcd := shad(inPub.Bytes())

			rcd := factoid.NewRCD_1(pub[:])
			inAdd := factoid.NewAddress(inHash.Bytes())
			outAdd := factoid.NewAddress(outEC.Bytes())

			trans := new(factoid.Transaction)
			trans.AddInput(inAdd, amt)
			trans.AddECOutput(outAdd, amt)

			trans.AddRCD(rcd)
			trans.AddAuthorization(rcd)
			trans.SetTimestamp(primitives.NewTimestampNow())

			fee, err := trans.CalculateFee(st.GetFactoshisPerEC())
			if err != nil {
				return err
			}
			input, err := trans.GetInput(0)
			if err != nil {
				return err
			}
			input.SetAmount(amt + fee)

			dataSig, err := trans.MarshalBinarySig()
			if err != nil {
				return err
			}
			sig := factoid.NewSingleSignatureBlock(inSec.Bytes(), dataSig)
			trans.SetSignatureBlock(0, sig)

			t := new(wsapi.TransactionRequest)
			data, _ := trans.MarshalBinary()
			t.Transaction = hex.EncodeToString(data)
			j := primitives.NewJSON2Request("factoid-submit", 0, t)
			_, err = v2Request(j, st.GetPort())
			//_, err = wsapi.HandleV2Request(st, j)
			if err != nil {
				return err
			}
			_ = err

			return nil
		}
	*/
}

type IEntryGenerator interface {
	AllEntries(height uint32) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int)
	NewEblock(height uint32) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int)
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

func (r *RandomEntryGenerator) AllEntries(height uint32) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	eblocks := make([]*entryBlock.EBlock, 0)
	commits := make([]*entryCreditBlock.CommitEntry, 0)
	entries := make([]*entryBlock.Entry, 0)
	totalCost := 0

	for i := 0; i < r.Config.EblocksPerHeight.Amount(); i++ {
		neb, nes, necs, t := r.NewEblock(height)
		eblocks = append(eblocks, neb)
		entries = append(entries, nes...)
		commits = append(commits, necs...)
		totalCost += t
	}
	return eblocks, entries, commits, totalCost
}

func (r *RandomEntryGenerator) NewEblock(height uint32) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	commits := make([]*entryCreditBlock.CommitEntry, 0)
	entries := make([]*entryBlock.Entry, 0)
	totalCost := 0

	head := r.NewEntry(primitives.NewZeroHash())
	// First one needs an extid
	head.ExternalIDs() = [][]byte{random.RandByteSliceOfLen(10), random.RandByteSliceOfLen(10)}
	head.ChainID = head.GetChainID()
	commit := r.newCommit(head)
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
		commit := r.newCommit(ent)
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
	ent.Content = random.RandByteSliceOfLen(bytes)
	ent.ChainID = chain
	return ent
}

func (r *RandomEntryGenerator) signCommit(entry *entryCreditBlock.CommitEntry) *entryCreditBlock.CommitEntry {
	entry.Sign(r.ECKey.Key[:])
	return entry
}

func (r *RandomEntryGenerator) newCommit(e *entryBlock.Entry) *entryCreditBlock.CommitEntry {
	commit := entryCreditBlock.NewCommitEntry()
	commit.EntryHash = e.GetHash()
	d, _ := e.MarshalBinary()
	commit.Credits, _ = util.EntryCost(d)
	commit.MilliTime = primitives.NewTimestampNow().GetTimeMilli()
	return commit
}
