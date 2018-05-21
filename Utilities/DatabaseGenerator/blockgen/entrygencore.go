package blockgen

import (
	"math/rand"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/util"
)

// EntryGenCore has functions that all entry gens can use (or override)
type EntryGenCore struct {
	ECKey  primitives.PrivateKey
	Config EntryGeneratorConfig

	// YOU MUST SET THIS
	//	Setting this allows for overrides. If you do not set this, your new implementation will not work
	Parent IEntryGenerator
}

func (r *EntryGenCore) AllEntries(height uint32, time interfaces.Timestamp) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	eblocks := make([]*entryBlock.EBlock, 0)
	commits := make([]*entryCreditBlock.CommitEntry, 0)
	entries := make([]*entryBlock.Entry, 0)
	totalCost := 0

	for i := 0; i < r.Config.EblocksPerHeight.Amount(); i++ {
		neb, nes, necs, t := r.Parent.NewEblock(height, time)
		eblocks = append(eblocks, neb)
		entries = append(entries, nes...)
		commits = append(commits, necs...)
		totalCost += t
	}
	return eblocks, entries, commits, totalCost
}

func (r *EntryGenCore) NewEblock(height uint32, time interfaces.Timestamp) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	commits := make([]*entryCreditBlock.CommitEntry, 0)
	entries := make([]*entryBlock.Entry, 0)
	totalCost := 0

	head := r.Parent.NewChainHead()
	commit := r.NewCommit(head, time)
	commit.Credits += 10
	totalCost += int(commit.Credits)
	commit = r.SignCommit(commit)

	eb := entryBlock.NewEBlock()
	eb.Header.SetChainID(head.ChainID)
	eb.Header.SetDBHeight(height)
	eb.AddEBEntry(head)

	entries = append(entries, head)
	commits = append(commits, commit)

	// now add the other entries
	for i := 0; i < r.Config.EntriesPerBlock.Amount(); i++ {
		ent := r.Parent.NewEntry(head.ChainID)
		commit := r.NewCommit(ent, time)
		commit = r.SignCommit(commit)
		totalCost += int(commit.Credits)
		eb.AddEBEntry(ent)

		entries = append(entries, ent)
		commits = append(commits, commit)
	}

	return eb, entries, commits, totalCost
}

func (r *EntryGenCore) NewChainHead() *entryBlock.Entry {
	head := r.Parent.NewEntry(primitives.NewZeroHash())
	// First one needs an extid
	head.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{random.RandByteSliceOfLen(10)}, primitives.ByteSlice{random.RandByteSliceOfLen(10)}}
	head.ChainID = entryBlock.ExternalIDsToChainID(head.ExternalIDs())
	return head
}

func (r *EntryGenCore) NewEntry(chain interfaces.IHash) *entryBlock.Entry {
	conf := r.Config
	bytes := rand.Intn(conf.EntrySize.Max) + conf.EntrySize.Max

	ent := entryBlock.NewEntry()
	ent.Content = primitives.ByteSlice{random.RandByteSliceOfLen(bytes)}
	ent.ChainID = chain
	return ent
}

func (r *EntryGenCore) SignCommit(entry *entryCreditBlock.CommitEntry) *entryCreditBlock.CommitEntry {
	entry.Sign(r.ECKey.Key[:])
	return entry
}

func (r *EntryGenCore) NewCommit(e *entryBlock.Entry, time interfaces.Timestamp) *entryCreditBlock.CommitEntry {
	commit := entryCreditBlock.NewCommitEntry()
	commit.EntryHash = e.GetHash()
	d, _ := e.MarshalBinary()
	commit.Credits, _ = util.EntryCost(d)
	var t primitives.ByteSlice6
	copy(t[:], milliTime(time.GetTimeSeconds())[:])
	commit.MilliTime = &t
	return commit
}

func (r *EntryGenCore) GetECKey() primitives.PrivateKey {
	return r.ECKey
}
