package state

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

type State struct {
	DB                    IDatabase
	currentDirectoryBlock IDirectoryBlock
	dBHeight              int
}

var _ IState = (*State)(nil)

func (s *State) CurrentDirectoryBlock() IDirectoryBlock {
	return s.currentDirectoryBlock
}

func (s *State) SetCurrentDirectoryBlock(dirblk IDirectoryBlock) {
	s.currentDirectoryBlock = dirblk
}

func (s *State) DB() IDatabase {
	return s.DB
}

func (s *State) SetDB(IDatabase) {
	s.DB = dirblk
}


func (s *State) DBHeight() int {
	return s.dBHeight
}

func (s *State) SetDBHeight(dbheight int) {
	s.dBHeight = dbheight
}
