package state

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

type State struct {
	db                    IDatabase
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
	return s.db
}

func (s *State) SetDB(db IDatabase) {
	s.db = db
}


func (s *State) DBHeight() int {
	return s.dBHeight
}

func (s *State) SetDBHeight(dbheight int) {
	s.dBHeight = dbheight
}
