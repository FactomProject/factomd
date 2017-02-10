package state

import (
	"fmt"

	//"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (s *State) SetUseTorrent(setVal bool) {
	s.useDBStateManager = setVal
}

func (s *State) UsingTorrent() bool {
	return s.useDBStateManager
}

func (s *State) UploadDBState(msg interfaces.IMsg) error {
	// Create the torrent
	if s.UsingTorrent() {
		//msg, err := s.LoadDBState(uint32(dbheight))
		//if err != nil {
		//		panic("[1] Error creating torrent in SaveDBStateToDB: " + err.Error())
		//	}
		d := msg.(*messages.DBStateMsg)
		fmt.Printf("Uploading DBState %d, Sigs: %d\n", d.DirectoryBlock.GetDatabaseHeight(), len(d.SignatureList.List))
		block := NewWholeBlock()
		block.DBlock = d.DirectoryBlock
		block.ABlock = d.AdminBlock
		block.FBlock = d.FactoidBlock
		block.ECBlock = d.EntryCreditBlock

		eHashes := make([]interfaces.IHash, 0)
		for _, e := range d.EBlocks {
			block.AddEblock(e)
			for _, eh := range e.GetEntryHashes() {
				eHashes = append(eHashes, eh)
			}
		}

		for _, e := range eHashes {
			if e.String()[:62] != "00000000000000000000000000000000000000000000000000000000000000" {
				//} else {
				ent, err := s.DB.FetchEntry(e)
				if err != nil {
					return fmt.Errorf("[2] Error creating torrent in SaveDBStateToDB: " + err.Error())
				}
				block.AddIEBEntry(ent)
			}
		}

		if len(d.SignatureList.List) == 0 {
			return fmt.Errorf("No signatures given, signatures must be in to be able to torrent")
		}
		block.SigList = d.SignatureList.List

		data, err := block.MarshalBinary()
		if err != nil {
			return fmt.Errorf("[3] Error creating torrent in SaveDBStateToDB: " + err.Error())

		}

		if s.IsLeader() {
			s.DBStateManager.UploadDBStateBytes(data, true)
		} else {
			s.DBStateManager.UploadDBStateBytes(data, false)
		}
	}
	return nil
}

func (s *State) GetMissingDBState(height uint32) {
	s.DBStateManager.RetrieveDBStateByHeight(height)
}
