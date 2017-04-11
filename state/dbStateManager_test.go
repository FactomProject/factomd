package state_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = fmt.Printf

func TestSaveDBState(t *testing.T) {
	// Init
	s := new(State)
	s.LoadConfig("", "")
	// Set custom config options
	s.Network = "CUSTOM"
	s.CustomNetworkID = []byte("unit-test")

	// DB Type to test
	s.DBType = "LDB"

	// Path so we can celanup any created files
	s.LogPath = "unit-test-db/"
	s.LdbPath = "unit-test-db/"
	s.BoltDBPath = "unit-test-db/"

	// So it starts...
	s.CustomBootstrapIdentity = "38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9"
	s.CustomBootstrapKey = "cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a"
	s.Init()
	s.Network = "CUSTOM"
	LoadDatabase(s)

	// Create blocks
	total := 100
	msgs := createTestDBStateList(total, s)
	for i, m := range msgs {
		s.FollowerExecuteDBState(m)
		var _ = i
		if i%1000 == 0 {
			//fmt.Printf("FE: %d\n", i)
		}
	}

	// Verify blocks
	errs := verifyBlocks(s, msgs)
	if errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
	}

	// Cleanup
	os.RemoveAll("unit-test-db/")
}

// Will verify a directory blc
func verifyBlocks(s *State, dbstates []interfaces.IMsg) []error {
	errs := make([]error, 0)
	for i, m := range dbstates {
		var _ = i
		if i%1000 == 0 {
			//fmt.Printf("VB: %d\n", i)
		}

		dbs := m.(*messages.DBStateMsg)
		err := foundByHeight(s, dbs)
		if err != nil {
			errs = append(errs, err)
		}

		err = foundByKeyMR(s, dbs)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

var blockMsgs []string = []string{"DBlock", "FBlock", "ECBlock"}

func foundByHeight(s *State, msg *messages.DBStateMsg) error {
	// Check that each item can be fetched by height
	dblock, err := s.DB.FetchDBlockByHeight(msg.DirectoryBlock.GetDatabaseHeight())
	if err != nil {
		return err
	} else if dblock == nil {
		return fmt.Errorf("Dblock from database is nil")
	}

	fblock, err := s.DB.FetchFBlockByHeight(msg.FactoidBlock.GetDBHeight())
	if err != nil {
		return err
	} else if fblock == nil {
		return fmt.Errorf("Fblock from database is nil")
	}

	ecBlock, err := s.DB.FetchECBlockByHeight(msg.EntryCreditBlock.GetDatabaseHeight())
	if err != nil {
		return err
	} else if ecBlock == nil {
		return fmt.Errorf("ECblock from database is nil")
	}

	// Check that they are correct
	err = compareAll([]interfaces.BinaryMarshallable{dblock, fblock, ecBlock},
		[]interfaces.BinaryMarshallable{msg.DirectoryBlock, msg.FactoidBlock, msg.EntryCreditBlock},
		blockMsgs,
		"Fetch by height")
	if err != nil {
		return err
	}

	return nil
}

func foundByKeyMR(s *State, msg *messages.DBStateMsg) error {
	// Check that each item can be fetched by height
	dblock, err := s.DB.FetchDBlockByPrimary(msg.DirectoryBlock.GetKeyMR())
	if err != nil {
		return err
	} else if dblock == nil {
		return fmt.Errorf("Dblock from database is nil")
	}

	fblock, err := s.DB.FetchFBlockByPrimary(msg.FactoidBlock.GetKeyMR())
	if err != nil {
		return err
	} else if fblock == nil {
		return fmt.Errorf("Fblock from database is nil")
	}

	ecBlock, err := s.DB.FetchECBlockByPrimary(msg.EntryCreditBlock.DatabasePrimaryIndex())
	if err != nil {
		return err
	} else if ecBlock == nil {
		return fmt.Errorf("ECblock from database is nil")
	}

	// Check that they are correct
	err = compareAll([]interfaces.BinaryMarshallable{dblock, fblock, ecBlock},
		[]interfaces.BinaryMarshallable{msg.DirectoryBlock, msg.FactoidBlock, msg.EntryCreditBlock},
		blockMsgs,
		"Fetch by KeyMr")
	if err != nil {
		return err
	}

	return nil
}

// Compare all blocks and spit out a good error
func compareAll(as []interfaces.BinaryMarshallable, bs []interfaces.BinaryMarshallable, msg []string, general string) error {
	for i, _ := range as {
		err := compareMarshal(as[i], bs[i])
		if err != nil {
			return fmt.Errorf("%s --> %s: %s", msg[i], general, err.Error())
		}
	}
	return nil
}

func compareMarshal(a interfaces.BinaryMarshallable, b interfaces.BinaryMarshallable) error {
	rawA, err := a.MarshalBinary()
	if err != nil {
		return err
	}

	rawB, err := b.MarshalBinary()
	if err != nil {
		return err
	}

	if bytes.Compare(rawA, rawB) != 0 {
		return fmt.Errorf("marshal binary data did not match")
	}

	return nil
}

func createTestDBStateList(blockCount int, s *State) []interfaces.IMsg {
	answer := make([]interfaces.IMsg, blockCount)
	var prev *testHelper.BlockSet = nil

	for i := 0; i < blockCount; i++ {
		var _ = i
		if i%1000 == 0 {
			//fmt.Printf("CM: %d\n", i)
		}

		prev = testHelper.CreateTestBlockSetWithNetworkID(prev, s.GetNetworkID())

		timestamp := primitives.NewTimestampNow()
		timestamp.SetTime(uint64(i * 1000 * 60 * 60 * 6)) //6 hours of difference between messages

		answer[i] = messages.NewDBStateMsg(timestamp, prev.DBlock, prev.ABlock, prev.FBlock, prev.ECBlock, nil, nil, nil)
	}
	return answer
}
