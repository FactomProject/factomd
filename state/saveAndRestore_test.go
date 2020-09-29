// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/state"
)

func TestPushPopBalanceMap(t *testing.T) {
	for i := 0; i < 1000; i++ {
		m := map[[32]byte]int64{}
		l := random.RandIntBetween(0, 1000)
		for j := 0; j < l; j++ {
			h := primitives.RandomHash()
			m[h.Fixed()] = random.RandInt64()
		}
		b := primitives.NewBuffer(nil)

		err := PushBalanceMap(b, m)
		if err != nil {
			t.Errorf("%v", err)
		}

		m2, err := PopBalanceMap(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(m) != len(m2) {
			t.Errorf("Map lengths are not identical - %v vs %v", len(m), len(m2))
		}

		for k := range m {
			if m[k] != m2[k] {
				t.Errorf("Invalid balances - %v vs %v", m[k], m2[k])
			}
		}
	}
}

func TestSaveRestore(t *testing.T) {

	ss := new(SaveState)
	ss.LeaderTimestamp = primitives.NewTimestampNow()
	ss.Init()

	ss2 := new(SaveState)
	ss2.LeaderTimestamp = ss.LeaderTimestamp
	ss2.Init()

	snil := (*SaveState)(nil)
	snil2 := snil
	if !snil.IsSameAs(snil2) {
		t.Error("Should be able to compare nils")
	}
	if snil.IsSameAs(ss) {
		t.Error("Should be able to compare a nil with a state")
	}
	if ss.IsSameAs(snil) {
		t.Error("Should be able to compare a state with a nil")
	}
	if !ss.IsSameAs(ss) {
		t.Error("One should be the same as one's self")
	}

	// Can we detect changes?
	{
		v := ss.DBHeight
		ss2.DBHeight = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBHeight = v
	}
	{
		ss2.FactoidBalancesP[primitives.Sha([]byte("ha ah")).Fixed()] = 10
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss.FactoidBalancesP[primitives.Sha([]byte("ha ah")).Fixed()] = 10
	}

	if !ss.IsSameAs(ss) {
		t.Error("One should be the same as one's self")
	}

	{
		ss2.ECBalancesP[primitives.Sha([]byte("ha ah")).Fixed()] = 10
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss.ECBalancesP[primitives.Sha([]byte("ha ah")).Fixed()] = 10
	}

	if !ss.IsSameAs(ss) {
		t.Error("One should be the same as one's self")
	}

	ident := new(identity.Identity)
	ident.IdentityChainID = primitives.Sha([]byte("ID"))
	ident.Keys[0] = primitives.Sha([]byte("key1"))
	ident.Keys[1] = primitives.Sha([]byte("key2"))
	ident.Keys[2] = primitives.Sha([]byte("key3"))
	ident.Keys[3] = primitives.Sha([]byte("key4"))
	ident.ManagementChainID = primitives.Sha([]byte("MID"))
	ident.MatryoshkaHash = primitives.Sha([]byte("MH"))
	ident.SigningKey = primitives.Sha([]byte("SK"))
	//{
	//	ss2.Identities = append(ss2.Identities, ident)
	//	if ss.IsSameAs(ss2) {
	//		t.Error("Note that we should be able to detect changes.")
	//	}
	//	ss.Identities = append(ss.Identities, ident)
	//}

	if !ss.IsSameAs(ss) {
		t.Error("One should be the same as one's self")
	}

	auth := new(identity.Authority)
	auth.AuthorityChainID = primitives.Sha([]byte("ID"))
	auth.ManagementChainID = primitives.Sha([]byte("MID"))
	auth.MatryoshkaHash = primitives.Sha([]byte("MH"))

	//{
	//	ss2.Authorities = append(ss2.Authorities, auth)
	//	if ss.IsSameAs(ss2) {
	//		t.Error("Note that we should be able to detect changes.")
	//	}
	//	ss.Authorities = append(ss.Authorities, auth)
	//
	//}

	if !ss.IsSameAs(ss) {
		t.Error("One should be the same as one's self")
	}

	{
		v := ss.AuthorityServerCount
		ss2.AuthorityServerCount = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.AuthorityServerCount = v
	}
	{
		v := ss.LLeaderHeight
		ss2.LLeaderHeight = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.LLeaderHeight = v
	}
	{
		v := ss.Leader
		ss2.Leader = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.Leader = v
	}
	{
		v := ss.LeaderVMIndex
		ss2.LeaderVMIndex = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.LeaderVMIndex = v
	}
	{
		v := ss.EOMsyncing
		ss2.EOMsyncing = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EOMsyncing = v
	}
	{
		v := ss.EOM
		ss2.EOM = !ss.EOM
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EOM = v
	}
	{
		v := ss.EOMLimit
		ss2.EOMLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EOMLimit = v
	}
	{
		v := ss.EOMProcessed
		ss2.EOMProcessed = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EOMProcessed = v
	}
	{
		v := ss.EOMDone
		ss2.EOMDone = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EOMDone = v
	}
	{
		v := ss.EOMMinute
		ss2.EOMMinute = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EOMMinute = v
	}
	{
		v := ss.EOMSys
		ss2.EOMSys = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EOMSys = v
	}
	{
		v := ss.DBSig
		ss2.DBSig = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSig = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}
	{
		v := ss.DBSigProcessed
		ss2.DBSigProcessed = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigProcessed = v
	}
	{
		v := ss.DBSigDone
		ss2.DBSigDone = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigDone = v
	}
	{
		v := ss.DBSigSys
		ss2.DBSigSys = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigSys = v
	}
	{
		v := ss.Newblk
		ss2.Newblk = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.Newblk = v
	}
	{
		v := ss.Saving
		ss2.Saving = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.Saving = v
	}
	{
		v := ss.Syncing
		ss2.Syncing = !v
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.Syncing = v
	}
	{
		v := ss.LeaderTimestamp
		ss2.LeaderTimestamp = primitives.NewTimestampFromMilliseconds(1000000)
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.LeaderTimestamp = v
	}
	{
		v := ss.EntryBlockDBHeightComplete
		ss2.EntryBlockDBHeightComplete = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EntryBlockDBHeightComplete = v
	}
	{
		v := ss.EntryBlockDBHeightComplete
		ss2.EntryBlockDBHeightComplete = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.EntryBlockDBHeightComplete = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}
	{
		v := ss.DBSigLimit
		ss2.DBSigLimit = v + 1
		if ss.IsSameAs(ss2) {
			t.Error("Note that we should be able to detect changes.")
		}
		ss2.DBSigLimit = v
	}

}
