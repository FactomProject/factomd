package blockgen

import (
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

type BlockGen struct {
	EntryGenerator  IEntryGenerator
	AuthoritySigner IAuthSigner
}

func NewBlockGen() *BlockGen {
	b := new(BlockGen)
	b.AuthoritySigner = new(DefaultAuthSigner)

	return b
}

func (bg *BlockGen) NewBlock(height uint32, prev *state.DBState) {
	next := new(state.DBState)

	// ABlock
	nab := bg.AuthoritySigner.SignBlock(prev)

	// Entries (need entries for ecblock)
	entries := bg.EntryGenerator.NewEntry()

	// ECBlock

	// FBlock
}

func newDblock(prev interfaces.IDirectoryBlock) interfaces.IDirectoryBlock {
	dblock := directoryBlock.NewDirectoryBlock(prev)
	return dblock
}
