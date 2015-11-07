package state

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (s *State) NewAdminBlock() interfaces.IAdminBlock {
	ab := new(adminBlock.AdminBlock)
	ab.Header = s.NewAdminBlockHeader()

	s.DB.SaveABlockHead(ab)

	return ab
}

func (s *State) NewAdminBlockHeader() interfaces.IABlockHeader {
	header := new(adminBlock.ABlockHeader)
	header.DBHeight = s.GetDBHeight()
	if s.GetCurrentAdminBlock() == nil {
		header.PrevLedgerKeyMR = primitives.NewHash(constants.ZERO_HASH)
	} else {
		keymr, err := s.GetCurrentAdminBlock().LedgerKeyMR()
		if err != nil {
			panic(err.Error())
		}
		header.PrevLedgerKeyMR = keymr
	}
	header.HeaderExpansionSize = 0
	header.HeaderExpansionArea = make([]byte, 0)
	header.MessageCount = 0
	header.BodySize = 0
	return header
}

func (s *State) CreateDBlock() (b interfaces.IDirectoryBlock, err error) {
	prev := s.GetCurrentDirectoryBlock()
	b = new(directoryBlock.DirectoryBlock)

	b.SetHeader(new(directoryBlock.DBlockHeader))
	b.GetHeader().SetVersion(constants.VERSION_0)

	if prev == nil {
		b.GetHeader().SetPrevLedgerKeyMR(primitives.NewZeroHash())
		b.GetHeader().SetPrevKeyMR(primitives.NewZeroHash())
	} else {
		prevLedgerKeyMR, err := primitives.CreateHash(prev)
		if err != nil {
			return nil, err
		}
		b.GetHeader().SetPrevLedgerKeyMR(prevLedgerKeyMR)
		keyMR, err := prev.BuildKeyMerkleRoot()
		if err != nil {
			return nil, err
		}
		b.GetHeader().SetPrevKeyMR(keyMR)
	}

	adminblk := s.NewAdminBlock()
	keymr, err := adminblk.GetKeyMR()
	if err != nil {
		panic(err.Error())
	}
	b.GetHeader().SetDBHeight(s.GetDBHeight())
	b.SetDBEntries(make([]interfaces.IDBEntry, 0))
	b.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID), keymr)
	b.AddEntry(primitives.NewHash(constants.EC_CHAINID), primitives.NewZeroHash())
	b.AddEntry(primitives.NewHash(constants.FACTOID_CHAINID), primitives.NewZeroHash())

	return b, err
}
