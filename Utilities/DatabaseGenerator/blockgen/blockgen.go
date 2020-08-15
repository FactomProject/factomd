package blockgen

import (
	"github.com/PaulSnow/factom2d/common/directoryBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/state"
)

// BlockGen can created full blocks. EntryGen generates entries, commits, and fct
// transactions. AuthoritySigner generates DBSigs.
// BlockGen packs all the responses into a DBState
type BlockGen struct {
	EntryGenerator  IFullEntryGenerator
	AuthoritySigner IAuthSigner
}

func NewBlockGen(config DBGeneratorConfig) (*BlockGen, error) {
	b := new(BlockGen)
	b.AuthoritySigner = new(DefaultAuthSigner)

	fkey, err := primitives.NewPrivateKeyFromHex("FB3B471B1DCDADFEB856BD0B02D8BF49ACE0EDD372A3D9F2A95B78EC12A324D6")
	if err != nil {
		return nil, err
	}

	b.EntryGenerator = NewFullEntryGenerator(*primitives.RandomPrivateKey(), *fkey, config)
	return b, nil
}

// NewBlock
//	Parameters
//		prev			*DBState	Previous DBState for all linking fields
//		netid			uint32		NetworkID for blockchain db
//		firsttimestamp	timestamp	Used for block 1 timestamp if on height 1
func (bg *BlockGen) NewBlock(prev *state.DBState, netid uint32, firstTimeStamp interfaces.Timestamp) (*state.DBState, error) {
	// ABlock
	nab := bg.AuthoritySigner.SignBlock(prev)
	next := primitives.Timestamp(0)
	next.SetTimeSeconds(prev.DirectoryBlock.GetHeader().GetTimestamp().GetTimeSeconds() + 60*10)
	if prev.DirectoryBlock.GetDatabaseHeight() == 0 {
		next = *firstTimeStamp.(*primitives.Timestamp)
	}

	// Entries (need entries for ecblock)
	newDBState, err := bg.EntryGenerator.NewBlockSet(prev, &next)
	if err != nil {
		return nil, err
	}
	newDBState.AdminBlock = nab
	newDBState.ABHash = nab.DatabasePrimaryIndex()

	// DBlock
	dblock := directoryBlock.NewDirectoryBlock(prev.DirectoryBlock)
	dblock.GetHeader().SetNetworkID(netid)
	dblock.GetHeader().SetTimestamp(&next)
	dblock.SetABlockHash(nab)
	dblock.SetECBlockHash(newDBState.EntryCreditBlock)
	dblock.SetFBlockHash(newDBState.FactoidBlock)

	for _, eb := range newDBState.EntryBlocks {
		k, _ := eb.KeyMR()
		err := dblock.AddEntry(eb.GetChainID(), k)
		if err != nil {
			panic(err)
		}
	}

	dblock.GetHeaderHash()
	dblock.BuildBodyMR()
	dblock.BuildKeyMerkleRoot()

	newDBState.DirectoryBlock = dblock

	return newDBState, nil
}
