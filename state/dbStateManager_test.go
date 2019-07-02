// +build all

package state_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = fmt.Printf
var _ = factoid.GetGenesisFBlock
var _ = constants.SIGNATURE_LENGTH

// Will verify a directory blc
func verifyBlocks(s *State, dbstates []interfaces.IMsg) []string {
	errs := make([]string, 0)
	for i, m := range dbstates {
		var _ = i
		if i%1000 == 0 {
			//fmt.Printf("VB: %d\n", i)
		}

		dbs := m.(*messages.DBStateMsg)
		err := foundByHeight(s, dbs)
		ht := fmt.Sprintf("%d", dbs.DirectoryBlock.GetDatabaseHeight())
		if err != nil {
			errs = append(errs, err.Error()+" foundByHeight failed for ht: "+ht)
		}

		err = foundByKeyMR(s, dbs)
		if err != nil {
			errs = append(errs, err.Error()+" foundByKeyMR failed for ht: "+ht)
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
	dblock, err := s.DB.FetchDBlock(msg.DirectoryBlock.GetKeyMR())
	if err != nil {
		return err
	} else if dblock == nil {
		return fmt.Errorf("Dblock from database is nil")
	}

	fblock, err := s.DB.FetchFBlock(msg.FactoidBlock.GetKeyMR())
	if err != nil {
		return err
	} else if fblock == nil {
		return fmt.Errorf("Fblock from database is nil")
	}

	ecBlock, err := s.DB.FetchECBlock(msg.EntryCreditBlock.DatabasePrimaryIndex())
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

func TestSaveDBState(t *testing.T) {
	// Init
	s := testHelper.CreatePopulateAndExecuteTestState()
	msgs := testHelper.GetAllDBStateMsgsFromDatabase(s)

	// Verify blocks
	errs := verifyBlocks(s, msgs)
	if errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
	}

	// Double Check DB
	errs = verifyBlocks(s, msgs)
	if errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
	}

	// Cleanup
	// os.RemoveAll("unit-test-db/")
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

func createTestDBStateList(blockCount int, s *State) ([]interfaces.IMsg, []interfaces.IAddress) {
	// FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q
	// Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK
	/*	sec, err := hex.DecodeString("FB3B471B1DCDADFEB856BD0B02D8BF49ACE0EDD372A3D9F2A95B78EC12A324D6")
		if err != nil {
			panic(err)
		}
		pub, err := hex.DecodeString("031CCE24BCC43B596AF105167DE2C03603C20ADA3314A7CFB47BEFCAD4883E6F")
		if err != nil {
			panic(err)
		}
	*/
	var err error
	answer := make([]interfaces.IMsg, blockCount)
	var prev *testHelper.BlockSet = nil
	adds := make([]interfaces.IAddress, 0)

	for i := 0; i < blockCount; i++ {
		if i%1000 == 0 {
			//fmt.Printf("CM: %d\n", i)
		}

		timestamp := primitives.NewTimestampNow()
		timestamp.SetTime(uint64(i * 1000 * 60 * 60 * 6)) //6 hours of difference between messages

		prev = testHelper.CreateTestBlockSetWithNetworkID(prev, s.GetNetworkID(), false)
		if i == 0 {
			dblk, ablk, fblk, ecblk := GenerateGenesisBlocks(s.GetNetworkID(), nil)
			msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk, nil, nil, nil)
			msg.(*messages.DBStateMsg).IgnoreSigs = true
			prev.DBlock = dblk.(*directoryBlock.DirectoryBlock)
			prev.ABlock = ablk.(*adminBlock.AdminBlock)
			prev.FBlock = fblk
			prev.ECBlock = ecblk
			answer[i] = msg
			continue
		}

		ents := prev.DBlock.GetDBEntries()
		for i, e := range ents {
			if bytes.Compare(e.GetChainID().Bytes(), constants.FACTOID_CHAINID) == 0 {
				//fromAdd []byte, toAdd []byte, amt uint64
				add := factoid.RandomAddress()
				adds = append(adds, add)
				newF := testHelper.CreateTestFactoidBlockWithTransaction(prev.FBlock,
					"FB3B471B1DCDADFEB856BD0B02D8BF49ACE0EDD372A3D9F2A95B78EC12A324D6", add.Bytes(), 10000)

				de := new(directoryBlock.DBEntry)
				de.ChainID, err = primitives.NewShaHash(newF.GetChainID().Bytes())
				if err != nil {
					panic(err)
				}
				de.KeyMR = newF.DatabasePrimaryIndex()

				ents[i] = de
				prev.FBlock = newF
			}
		}

		prev.DBlock.SetDBEntries(ents)

		answer[i] = messages.NewDBStateMsg(timestamp, prev.DBlock, prev.ABlock, prev.FBlock, prev.ECBlock, nil, nil, nil)
		answer[i].(*messages.DBStateMsg).IgnoreSigs = true
	}
	return answer, adds
}
