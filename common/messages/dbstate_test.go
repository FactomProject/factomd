// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	statepkg "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func coupleOfSigs(t *testing.T) []interfaces.IFullSignature {
	priv1 := new(primitives.PrivateKey)

	err := priv1.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	msg1 := "Test Message Sign1"
	msg2 := "Test Message Sign2"

	sig1 := priv1.Sign([]byte(msg1))
	sig2 := priv1.Sign([]byte(msg2))

	var twoSigs []interfaces.IFullSignature
	twoSigs = append(twoSigs, sig1)
	twoSigs = append(twoSigs, sig2)
	return twoSigs
}

func makeSigList(t *testing.T) SigList {
	priv1 := new(primitives.PrivateKey)

	err := priv1.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	msg1 := "Test Message Sign1"
	msg2 := "Test Message Sign2"

	sig1 := priv1.Sign([]byte(msg1))
	sig2 := priv1.Sign([]byte(msg2))

	var twoSigs []interfaces.IFullSignature
	twoSigs = append(twoSigs, sig1)
	twoSigs = append(twoSigs, sig2)

	sl := new(SigList)
	sl.Length = 2
	sl.List = twoSigs
	return *sl
}

func TestUnmarshalNilDBStateMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DBStateMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalDBStateMsg(t *testing.T) {
	msg := newDBStateMsg()
	msg.String()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	//t.Logf("Marshalled - %x", hex)

	msg2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str, _ := msg.JSONString()
	//t.Logf("str1 - %v", str)
	str, _ = msg2.JSONString()
	//t.Logf("str2 - %v", str)
	var _ = str

	if msg2.Type() != constants.DBSTATE_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*DBStateMsg).MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(hex) != len(hex2) {
		t.Error("Hexes aren't of identical length")
	}
	for i := range hex {
		if hex[i] != hex2[i] {
			t.Error("Hexes do not match")
		}
	}

	if msg.IsSameAs(msg2.(*DBStateMsg)) != true {
		t.Errorf("DBStateMsg messages are not identical")
	}

	// Test Invalid IsSameAs
	// Simple testing
	if msg.IsSameAs(nil) == true {
		t.Error("DBState msg compare should be false to nil")
	}

	tmp := *msg
	tmp.Timestamp = primitives.NewTimestampNow()

	if msg.IsSameAs(&tmp) == true {
		t.Error("DBState msg compare should be false")
	}
	tmp.Timestamp = msg.Timestamp

	tmp.DirectoryBlock = testHelper.CreateTestDirectoryBlock(nil)
	if msg.IsSameAs(&tmp) == true {
		t.Error("DBState msg compare should be false")
	}
	tmp.DirectoryBlock = msg.DirectoryBlock

	tmp.AdminBlock = testHelper.CreateTestAdminBlock(nil)
	if msg.IsSameAs(&tmp) == true {
		t.Error("DBState msg compare should be false")
	}
	tmp.AdminBlock = msg.AdminBlock

	tmp.EntryCreditBlock = testHelper.CreateTestEntryCreditBlock(nil)
	if msg.IsSameAs(&tmp) == true {
		t.Error("DBState msg compare should be false")
	}
	tmp.EntryCreditBlock = msg.EntryCreditBlock

	tmp.FactoidBlock = testHelper.CreateTestFactoidBlock(nil)
	if msg.IsSameAs(&tmp) == true {
		t.Error("DBState msg compare should be false")
	}
	tmp.FactoidBlock = msg.FactoidBlock
}

func TestSimpleDBStateMsgValidate(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()

	msg := new(DBStateMsg)
	if msg.Validate(state) >= 0 {
		t.Errorf("Empty DBState validated")
	}

	msg = newDBStateMsg()
	msg.DirectoryBlock.GetHeader().SetNetworkID(0x00)
	if msg.Validate(state) >= 0 {
		t.Errorf("Wrong network ID validated")
	}

	msg = newDBStateMsg()
	msg.DirectoryBlock.GetHeader().SetNetworkID(constants.MAIN_NETWORK_ID)
	msg.DirectoryBlock.GetHeader().SetDBHeight(state.GetHighestSavedBlk() + 1)
	constants.CheckPoints[state.GetHighestSavedBlk()+1] = "123"
	if msg.Validate(state) >= 0 {
		t.Errorf("Wrong checkpoint validated")
	}

	delete(constants.CheckPoints, state.GetHighestSavedBlk()+1)
}

func TestDBStateDataValidate(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	msg := newDBStateMsg()

	if v := msg.ValidateData(state); v != 1 {
		t.Errorf("Validate data should be 1, found %d", v)
	}

	// Invalidate it
	eblock, _ := testHelper.CreateTestEntryBlock(nil)
	msg.EBlocks = append(msg.EBlocks, eblock)
	if v := msg.ValidateData(state); v != -1 {
		t.Errorf("Should be -1, found %d", v)
	}

	msg2 := newDBStateMsg()
	e := entryBlock.NewEntry()
	msg2.Entries = append(msg2.Entries, e)
	if v := msg2.ValidateData(state); v != -1 {
		t.Errorf("Should be -1, found %d", v)
	}

}

// Test known conditions
//		All sign
//		Half + 1 Sign
//		Half Sign
func TestSignedDBStateValidate(t *testing.T) {
	type SmallIdentity struct {
		ID  interfaces.IHash
		Key primitives.PrivateKey
	}

	ids := make([]SmallIdentity, 100)
	for i := range ids {
		tid, err := primitives.HexToHash("888888" + fmt.Sprintf("%058d", i))
		if err != nil {
			panic(err)
		}
		ids[i] = SmallIdentity{
			ID: tid,
			// Can you believe there isn't a 'RandomPrivateKey' that returns a private key.
			// Well this works
			Key: *primitives.RandomPrivateKey(),
		}
	}

	state := testHelper.CreateEmptyTestState()

	// Throw in a genesis block
	prev := testHelper.CreateTestBlockSetWithNetworkID(nil, state.GetNetworkID(), false)
	dblk, ablk, fblk, ecblk := statepkg.GenerateGenesisBlocks(state.GetNetworkID(), nil)
	prev.DBlock = dblk.(*directoryBlock.DirectoryBlock)
	prev.ABlock = ablk.(*adminBlock.AdminBlock)
	prev.FBlock = fblk
	prev.ECBlock = ecblk
	genDBState := NewDBStateMsg(state.GetTimestamp(), prev.DBlock, prev.ABlock, prev.FBlock, prev.ECBlock, nil, nil, nil)
	if genDBState.Validate(state) != 1 {
		t.Error("Genesis should always be valid")
	}

	state.FollowerExecuteDBState(genDBState)
	// Ok genesis set

	timestamp := primitives.NewTimestampNow()
	start := 0

	// Out of 50 DbStates.
	//		i % 6 					--> Valid (Total/2 + 1 Signing Correctly)
	//		i %2 == 0 && i % 6 != 0 --> Invalid (Total/2 Signing Correctly)
	//      Rest					--> Valid (Total Signing Correctly)
	for i := 1; i < 50; i++ {
		timestamp.SetTime(uint64(i * 1000 * 60 * 60 * 6))
		end := i
		var signers []SmallIdentity

		var a *adminBlock.AdminBlock
		good := end - start
		if i%2 == 0 { // Half bad sigs if even number
			good = good / 2
		}

		// ABlock
		a = testHelper.CreateTestAdminBlock(prev.ABlock)
		for i := start; i < end; i++ {
			id := ids[i]
			a.AddFedServer(id.ID)
			a.AddFederatedServerSigningKey(id.ID, id.Key.Pub.Fixed())
			signers = append(signers, id)
			//a := identity.NewAuthority()
			//a.AuthorityChainID = id.ID
			//a.SigningKey = *id.Key.Pub
			//state.IdentityControl.SetAuthority(id.ID, a)
		}
		a.InsertIdentityABEntries()

		set, err := createBlockFromAdmin(a, prev, state)
		if err != nil {
			t.Error(err)
			continue
		}

		d := set.DBlock
		// state.ProcessLists.Get(d.GetDatabaseHeight()).FedServers = make([]interfaces.IServer, 0)
		tot := 0
		for c := start; c < end; c++ {
			tot++
			state.ProcessLists.Get(d.GetDatabaseHeight()).AddFedServer(ids[c].ID)
		}

		// DBState
		plusone := false
		var dbSigList []interfaces.IFullSignature
		for c, s := range signers {
			// Half bad sigs if even number
			if i%2 == 0 && c%2 == 0 {
				data := []byte{0x00}
				dbSigList = append(dbSigList, s.Key.Sign(data))
			} else {
				data, _ := d.GetHeader().MarshalBinary()
				dbSigList = append(dbSigList, s.Key.Sign(data))
				continue
			}

			// Add +1 to get majority
			if !plusone && i%6 == 0 {
				good++
				plusone = true
				data, _ := d.GetHeader().MarshalBinary()
				dbSigList = append(dbSigList, s.Key.Sign(data))
			}
		}

		msg := NewDBStateMsg(timestamp, set.DBlock, set.ABlock, set.FBlock, set.ECBlock, nil, nil, dbSigList)
		m := msg.(*DBStateMsg)
		m.IgnoreSigs = true
		if i%2 == 0 {
			if i%6 == 0 {
				if m.ValidateSignatures(state) < 0 {
					t.Errorf("[0] Should be valid, found %d", m.ValidateSignatures(state))
				}
			} else {
				if m.ValidateSignatures(state) > 0 {
					t.Errorf("%s Should be invalid %d, Sigs: %d, Feds: %d", "s", m.ValidateSignatures(state), len(m.SignatureList.List), len(state.ProcessLists.Get(d.GetDatabaseHeight()).FedServers))
				}
			}
		} else {
			if m.ValidateSignatures(state) < 0 && len(m.SignatureList.List) > (len(state.ProcessLists.Get(d.GetDatabaseHeight()).FedServers)/2)+1 {
				t.Errorf("[2] %s Should be valid %d, Sigs: %d, Feds: %d", "s", m.ValidateSignatures(state), len(m.SignatureList.List), len(state.ProcessLists.Get(d.GetDatabaseHeight()).FedServers))
			}
		}

		if m.SigTally(state) != good {
			t.Errorf("TallySig found %d, should be %d", m.SigTally(state), good)
		}
		var _ = msg
	}
}

// Test Random Conditions
//		Random number of Feds
//		Random # of them sign
//		Random # of them Removed
func TestPropSignedDBStateValidate(t *testing.T) {
	type SmallIdentity struct {
		ID  interfaces.IHash
		Key primitives.PrivateKey
	}

	state := testHelper.CreateEmptyTestState()

	ids := make([]SmallIdentity, 100)
	for i := range ids {
		tid, err := primitives.HexToHash("888888" + fmt.Sprintf("%058d", i))
		if err != nil {
			panic(err)
		}
		ids[i] = SmallIdentity{
			ID: tid,
			// Can you believe there isn't a 'RandomPrivateKey' that returns a private key.
			// Well this works
			Key: *primitives.RandomPrivateKey(),
		}
	}

	// Throw in a geneis block
	prev := testHelper.CreateTestBlockSetWithNetworkID(nil, state.GetNetworkID(), false)
	dblk, ablk, fblk, ecblk := statepkg.GenerateGenesisBlocks(state.GetNetworkID(), nil)
	prev.DBlock = dblk.(*directoryBlock.DirectoryBlock)
	prev.ABlock = ablk.(*adminBlock.AdminBlock)
	prev.FBlock = fblk
	prev.ECBlock = ecblk
	genDBState := NewDBStateMsg(state.GetTimestamp(), prev.DBlock, prev.ABlock, prev.FBlock, prev.ECBlock, nil, nil, nil)
	genDBState.(*DBStateMsg).IgnoreSigs = true
	state.FollowerExecuteDBState(genDBState)
	// Ok Geneis set

	timestamp := primitives.NewTimestampNow()

	for i := 1; i < 100; i++ {
		totalFed := 0
		totalRemove := 0
		timestamp.SetTime(uint64(i * 1000 * 60 * 60 * 6))
		var signers []SmallIdentity

		a := testHelper.CreateTestAdminBlock(prev.ABlock)
		state.ProcessLists.Get(a.GetDatabaseHeight()).FedServers = make([]interfaces.IServer, 0)
		for ia := 0; ia < len(ids); ia++ {
			switch random.RandIntBetween(0, 4) {
			case 1: // Signing Fed
				if totalFed >= 60 {
					continue
				}
				state.ProcessLists.Get(a.GetDatabaseHeight()).AddFedServer(ids[ia].ID)
				a.AddFedServer(ids[ia].ID)
				a.AddFederatedServerSigningKey(ids[ia].ID, ids[ia].Key.Pub.Fixed())

				signers = append(signers, ids[ia])
				totalFed++
			case 2: // Not signing Fed
				if totalFed >= 60 {
					continue
				}
				state.ProcessLists.Get(a.GetDatabaseHeight()).AddFedServer(ids[ia].ID)
				a.AddFedServer(ids[ia].ID)
				a.AddFederatedServerSigningKey(ids[ia].ID, ids[ia].Key.Pub.Fixed())

				totalFed++
			case 3:
				if totalRemove > totalFed {
					break
				}
				err := a.RemoveFederatedServer(ids[ia].ID)
				if err != nil {
					t.Error(err)
				}
				totalRemove++
			}
		}
		a.InsertIdentityABEntries()

		set, err := createBlockFromAdmin(a, prev, state)
		if err != nil {
			t.Error(err)
			continue
		}

		d := set.DBlock

		// DBState
		var dbSigList []interfaces.IFullSignature
		for _, s := range signers {
			data, e := d.GetHeader().MarshalBinary()
			if e != nil {
				t.Error(e)
			}
			dbSigList = append(dbSigList, s.Key.Sign(data))
		}

		msg := NewDBStateMsg(timestamp, set.DBlock, set.ABlock, set.FBlock, set.ECBlock, nil, nil, dbSigList)
		m := msg.(*DBStateMsg)

		if m.SigTally(state) != len(signers) {
			t.Errorf("%s TallySig found %d, should be %d", m.DirectoryBlock.GetKeyMR().String()[:5], m.SigTally(state), len(signers))
		}

		v := m.ValidateSignatures(state)
		need := ((totalFed - totalRemove) / 2) + 1
		if len(signers) >= need {
			if v < 0 {
				t.Error(m.DirectoryBlock.GetKeyMR().String()[:5], "Should be valid", len(signers), totalFed, totalRemove, len(m.SignatureList.List), len(state.GetFedServers(m.DirectoryBlock.GetDatabaseHeight())))
			}
		} else {
			if v > 0 {
				t.Errorf("Should be invalid. V:%d Signers: %d, Feds: %d, Rem: %d, Sigs: %d, SFed: %d", v, len(signers), totalFed, totalRemove, len(m.SignatureList.List), len(state.GetFedServers(m.DirectoryBlock.GetDatabaseHeight())))
			}
		}
		var _ = msg
	}
}

func createBlockFromAdmin(a *adminBlock.AdminBlock, prev *testHelper.BlockSet, st *statepkg.State) (*testHelper.BlockSet, error) {
	s := new(testHelper.BlockSet)
	var err error

	dbEntries := []interfaces.IDBEntry{}
	de := new(directoryBlock.DBEntry)
	de.ChainID = a.GetChainID()
	de.KeyMR, err = a.GetKeyMR()
	if err != nil {
		return nil, err
	}
	dbEntries = append(dbEntries, de)

	// FBlock
	f := testHelper.CreateTestFactoidBlockWithCoinbase(prev.FBlock, testHelper.NewFactoidAddress(0), testHelper.DefaultCoinbaseAmount)
	de = new(directoryBlock.DBEntry)
	de.KeyMR = f.GetKeyMR()
	de.ChainID = f.GetChainID()
	dbEntries = append(dbEntries, de)

	//ECBlock
	ec := testHelper.CreateTestEntryCreditBlock(prev.ECBlock)
	de = new(directoryBlock.DBEntry)
	de.KeyMR = ec.DatabasePrimaryIndex()
	de.ChainID = ec.GetChainID()
	dbEntries = append(dbEntries, de)

	// DBlock
	d := testHelper.CreateTestDirectoryBlockWithNetworkID(prev.DBlock, st.GetNetworkID())
	err = d.SetDBEntries(dbEntries)
	d.MarshalBinary()

	s.DBlock = d
	s.ABlock = a
	s.ECBlock = ec
	s.FBlock = f
	return s, err
}

func newDBStateMsg() *DBStateMsg {
	msg := new(DBStateMsg)
	msg.Timestamp = primitives.NewTimestampNow()

	set := testHelper.CreateTestBlockSet(nil)
	set = testHelper.CreateTestBlockSet(set)

	msg.DirectoryBlock = set.DBlock
	msg.AdminBlock = set.ABlock
	msg.FactoidBlock = set.FBlock
	msg.EntryCreditBlock = set.ECBlock
	msg.EBlocks = []interfaces.IEntryBlock{set.EBlock, set.AnchorEBlock}
	for _, e := range set.Entries {
		msg.Entries = append(msg.Entries, e)
	}

	return msg
}
