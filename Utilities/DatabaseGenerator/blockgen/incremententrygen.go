package blockgen

import (
	"fmt"

	"sync"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

/********************
 * EntryGen Example *
 *   with a state   *
 ********************/

// IncrementEntryGenerator generates entries of incrementing count
//		The count is reset per chain.
type IncrementEntryGenerator struct {
	EntryGenCore // Has supporting functions and fields

	// The stateful object
	currentCount int
	sync.Mutex
}

func NewIncrementEntryGenerator(ecKey primitives.PrivateKey, config EntryGeneratorConfig) *IncrementEntryGenerator {
	r := new(IncrementEntryGenerator)
	r.ECKey = ecKey
	r.Config = config
	r.Parent = r

	return r
}

func (r *IncrementEntryGenerator) Name() string {
	return "IncrementEntryGenerator"
}

func (r *IncrementEntryGenerator) NewChainHead() *entryBlock.Entry {
	// Reset the count for the next chain
	r.Lock()
	r.currentCount = 0
	r.Unlock()
	return r.EntryGenCore.NewChainHead()
}
func (r *IncrementEntryGenerator) NewEntry(chain interfaces.IHash) *entryBlock.Entry {
	r.Lock()
	defer r.Unlock()
	ent := entryBlock.NewEntry()
	// Putting the ASCII number so you can read in explorer
	ent.Content = primitives.ByteSlice{[]byte(fmt.Sprintf("%d", r.currentCount))}
	ent.ChainID = chain
	r.currentCount++

	return ent
}

// Default implementation
func (r *IncrementEntryGenerator) AllEntries(height uint32, time interfaces.Timestamp) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	return r.EntryGenCore.AllEntries(height, time)
}
func (r *IncrementEntryGenerator) NewEblock(height uint32, time interfaces.Timestamp) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	return r.EntryGenCore.NewEblock(height, time)
}
