package blockgen

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// RandomEntryGenerator generates random entries between 0-10kbish
//	It does not override any of the core functions
type RandomEntryGenerator struct {
	EntryGenCore // Has supporting functions and fields
}

func NewRandomEntryGenerator(ecKey primitives.PrivateKey, config EntryGeneratorConfig) *RandomEntryGenerator {
	r := new(RandomEntryGenerator)
	r.ECKey = ecKey
	r.Config = config
	r.Parent = r

	return r
}

func (r *RandomEntryGenerator) Name() string {
	return "RandomEntryGenerator"
}

// Default implementation
func (r *RandomEntryGenerator) AllEntries(height uint32, time interfaces.Timestamp) ([]*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	return r.EntryGenCore.AllEntries(height, time)
}
func (r *RandomEntryGenerator) NewEblock(height uint32, time interfaces.Timestamp) (*entryBlock.EBlock, []*entryBlock.Entry, []*entryCreditBlock.CommitEntry, int) {
	return r.EntryGenCore.NewEblock(height, time)
}
func (r *RandomEntryGenerator) NewChainHead() *entryBlock.Entry {
	return r.EntryGenCore.NewChainHead()
}
func (r *RandomEntryGenerator) NewEntry(chain interfaces.IHash) *entryBlock.Entry {
	return r.EntryGenCore.NewEntry(chain)
}
