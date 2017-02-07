package state

import ()

func (s *State) SetUseTorrent(setVal bool) {
	s.useDBStateManager = setVal
}

func (s *State) UsingTorrent() bool {
	return s.useDBStateManager
}

func (s *State) GetMissingDBState(height uint32) {
	s.DBStateManager.RetrieveDBStateByHeight(height)
}
