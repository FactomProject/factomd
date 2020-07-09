package chainheadfix

import (
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

// Workers specifies the amounts of worker goroutines used to scan through the database.
// There is a strong diminishing return for higher values.
var Workers = 8

// Summary of Chain Head Fixing:
// To ensure that the database is consistent and that the database entry
// for "GetChainHead" points to the most recent EBlock of every chain.
//
// This is accomplished in a multithreaded three stage process:
//   1. Iterate over all DBlocks and save the maximum height & keymr of every chain
//   2. Compare the accurate chain heads of 1. with the database results of chain heads
//   3. (if enabled) Update the database for any entries that mismatch in 2.
//
// The order that DBlocks are iterated over doesn't matter. Currently the limiting factor
// is the i/o from the database.
type finder struct {
	// general
	db       *databaseOverlay.Overlay
	log      *log.Entry
	fixHeads bool

	// keeping track of actual chain heads
	// results of the first step
	heads    map[[32]byte]*head
	headsMtx sync.Mutex

	// multithreaded administrative vars
	wg     sync.WaitGroup
	gather chan uint32
	fix    chan *head

	// counter of how many anomalies were detected
	errors int64

	// results of the second st ep
	batchMtx    sync.Mutex
	batchKeys   []interfaces.IHash
	batchValues []interfaces.IHash
}

// retrieve the head struct for a chain id.
// creates a new head if none is found.
func (f *finder) get(id interfaces.IHash) *head {
	f.headsMtx.Lock()
	var ch *head
	var ok bool
	key := id.Fixed()
	if ch, ok = f.heads[key]; !ok {
		if ch, ok = f.heads[key]; !ok {
			ch = newhead(id)
			f.heads[key] = ch
		}
	}
	f.headsMtx.Unlock()
	return ch
}

// step 1
// launches workers and writes all database heights to worker channel.
// returns an error if the dblock height could not be determined.
func (f *finder) gatherHeads(workers int) error {
	head, err := f.db.FetchDBlockHead()
	if err != nil {
		return err
	}

	// empty database
	if head == nil {
		return nil
	}

	f.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go f.gatherer()
	}

	height := head.GetDatabaseHeight()
	log.Infof("checking chainheads of %d DBlocks", height)
	for i := uint32(0); i <= height; i++ {
		f.gather <- i
	}
	close(f.gather)
	f.wg.Wait()

	return nil
}

// step 2
// launches workers and writes all known heads to worker channel.
func (f *finder) checkHeads(workers int) {
	f.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go f.fixer()
	}

	for _, ch := range f.heads {
		f.fix <- ch
	}

	close(f.fix)
	f.wg.Wait()
}

// run the algorithm with the specified number of workers.
func (f *finder) run(workers int) error {
	if workers < 1 {
		workers = 1
	}

	start := time.Now()

	// step 1
	if err := f.gatherHeads(workers); err != nil {
		return err
	}
	f.log.WithField("time", time.Since(start)).Info("chainheads gathered")

	// step 2
	f.checkHeads(workers)

	// step 3
	fixed := 0
	if len(f.batchKeys) > 0 {
		if err := f.db.SetChainHeads(f.batchKeys, f.batchValues); err != nil {
			return err
		}
		fixed = len(f.batchKeys)
	}

	f.log.WithFields(log.Fields{
		"errors":  f.errors,
		"fixed":   fixed,
		"checked": len(f.heads),
		"total":   time.Since(start),
	}).Infof("chainheadfix finished")

	return nil
}

// worker routine for step 1.
// grabs dblock from database for heights and updates the head.
func (f *finder) gatherer() {
	for h := range f.gather {
		dblock, err := f.db.FetchDBlockByHeight(h)
		if err != nil {
			f.log.Errorf("unable to retrieve dblock for height %d: %v", h, err)
			continue
		}

		for _, eblock := range dblock.GetEBlockDBEntries() {
			f.get(eblock.GetChainID()).Update(int64(h), eblock.GetKeyMR())
		}
	}

	f.wg.Done()
}

// worker routine for step 2.
// takes max heads from step 1 and compares them to the database value. stores anomalies.
func (f *finder) fixer() {
	for ch := range f.fix {
		head, err := f.db.FetchHeadIndexByChainID(ch.id)
		if err != nil {
			f.log.Errorf("unable to retrieve a head index for chain %x: %v", ch.id, err)
		}

		if !head.IsSameAs(ch.head) {
			atomic.AddInt64(&f.errors, 1)
			if f.fixHeads {
				f.batchMtx.Lock()
				f.batchKeys = append(f.batchKeys, ch.id)
				f.batchValues = append(f.batchValues, ch.head)
				f.batchMtx.Unlock()
			}
			f.log.WithFields(log.Fields{
				"chain":    ch.id.String(),
				"got":      head.String(),
				"expected": ch.head.String(),
				"height":   ch.height,
			}).Warn("bad chainhead found")
		}
	}
	f.wg.Done()
}

// FindHeads checkes the databases to find any chains where the chain head
// in the database does not match the actual chain head.
// If specified, these anomalies are also fixed.
func FindHeads(db *databaseOverlay.Overlay, fixHeads bool) error {
	find := new(finder)
	find.db = db
	find.log = log.WithField("module", "chainheadfix")
	find.heads = make(map[[32]byte]*head)
	find.gather = make(chan uint32, 32)
	find.fix = make(chan *head, 32)
	find.fixHeads = fixHeads

	return find.run(Workers)
}
