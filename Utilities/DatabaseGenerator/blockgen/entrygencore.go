package blockgen

import (
	"math/rand"

	"sync"

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

	threadpoolon bool
	jobs         chan Job
	results      chan *Resp
	quit         chan bool
}

func (r *EntryGenCore) AllEntries(height uint32, time interfaces.Timestamp) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	eblocks := make([]*entryBlock.EBlock, 0)
	commits := make([]*entryCreditBlock.CommitEntry, 0)
	entries := make([]*entryBlock.Entry, 0)
	totalCost := 0

	if r.Config.Multithreaded {
		r.InitThreadPool()
		// Multithread the EBlocks
		var wg sync.WaitGroup
		quit := make(chan bool, 2)
		go func() {
			for {
				select {
				case resp := <-r.results:
					eblocks = append(eblocks, resp.Neb)
					entries = append(entries, resp.Nes...)
					commits = append(commits, resp.Nec...)
					totalCost += resp.T
					wg.Done()
				case <-quit:
					return
				}
			}
		}()

		for i := 0; i < r.Config.EblocksPerHeight.Amount(); i++ {
			r.jobs <- Job{height, time}
			wg.Add(1)
		}

		wg.Wait()
		quit <- true

	} else {
		// Single thread
		for i := 0; i < r.Config.EblocksPerHeight.Amount(); i++ {
			neb, nes, necs, t := r.Parent.NewEblock(height, time)
			eblocks = append(eblocks, neb)
			entries = append(entries, nes...)
			commits = append(commits, necs...)
			totalCost += t
		}
	}
	return eblocks, entries, commits, totalCost
}

type Resp struct {
	Neb *entryBlock.EBlock
	Nec []*entryCreditBlock.CommitEntry
	Nes []*entryBlock.Entry
	T   int
}

type Job struct {
	Height uint32
	Time   interfaces.Timestamp
}

func (r *EntryGenCore) InitThreadPool() {
	if !r.threadpoolon {
		r.jobs = make(chan Job, 20)
		r.results = make(chan *Resp, 20)
		r.quit = make(chan bool, 10)
		if r.Config.ThreadpoolCount == 0 {
			r.Config.ThreadpoolCount = 8
		}
		for i := 0; i < r.Config.ThreadpoolCount; i++ {
			go r.multiThreadWorker()
		}
	}
}

func (r *EntryGenCore) multiThreadWorker() {
	for {
		select {
		case j := <-r.jobs:
			res := new(Resp)
			res.Neb, res.Nes, res.Nec, res.T = r.Parent.NewEblock(j.Height, j.Time)
			r.results <- res
		case <-r.quit:
			r.quit <- true
			return
		}
	}
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
	for i := 0; i < r.Config.EntriesPerEBlock.Amount(); i++ {
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
