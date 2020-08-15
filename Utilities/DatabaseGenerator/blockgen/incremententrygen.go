package blockgen

import (
	"fmt"

	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
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
	r.currentCount = 0
	return r.EntryGenCore.NewChainHead()
}
func (r *IncrementEntryGenerator) NewEntry(chain interfaces.IHash) *entryBlock.Entry {
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
