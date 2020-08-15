package blockgen

import (
	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

// RecordEntryGenerator
type RecordEntryGenerator struct {
	EntryGenCore // Has supporting functions and fields
}

func NewRecordEntryGenerator(ecKey primitives.PrivateKey, config EntryGeneratorConfig) *RecordEntryGenerator {
	r := new(RecordEntryGenerator)
	r.ECKey = ecKey
	r.Config = config
	r.Parent = r

	return r
}

func (r *RecordEntryGenerator) Name() string {
	return "RecordEntryGenerator"
}

func (r *RecordEntryGenerator) NewChainHead() *entryBlock.Entry {
	return r.EntryGenCore.NewChainHead()
}
func (r *RecordEntryGenerator) NewEntry(chain interfaces.IHash) *entryBlock.Entry {
	ent := entryBlock.NewEntry()
	// Putting the ASCII number so you can read in explorer
	ent.ExtIDs = []primitives.ByteSlice{
		primitives.ByteSlice{primitives.RandomHash().Bytes()},
		primitives.ByteSlice{primitives.RandomHash().Bytes()},
		primitives.ByteSlice{primitives.RandomHash().Bytes()},
		primitives.ByteSlice{primitives.RandomHash().Bytes()},
		primitives.ByteSlice{primitives.RandomHash().Bytes()},
	}

	ent.ChainID = chain
	return ent
}

// Default implementation
func (r *RecordEntryGenerator) AllEntries(height uint32, time interfaces.Timestamp) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	return r.EntryGenCore.AllEntries(height, time)
}
func (r *RecordEntryGenerator) NewEblock(height uint32, time interfaces.Timestamp) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	return r.EntryGenCore.NewEblock(height, time)
}
