// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	//"bytes"
	//"encoding/binary"
	"testing"
	//"time"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAuthoritySignature(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	idindex := s.CreateBlankFactomIdentity(primitives.NewZeroHash())
	s.Identities[idindex].ManagementChainID = primitives.NewZeroHash()

	index := s.AddAuthorityFromChainID(primitives.NewZeroHash())
	s.Authorities[index].SigningKey = *(s.GetServerPublicKey())
	s.Authorities[index].Status = 1

	ack := new(messages.Ack)
	ack.DBHeight = s.LLeaderHeight
	ack.VMIndex = 1
	ack.Minute = byte(5)
	ack.Timestamp = s.GetTimestamp()
	ack.MessageHash = primitives.NewZeroHash()
	ack.LeaderChainID = s.IdentityChainID
	ack.SerialHash = primitives.NewZeroHash()

	err := ack.Sign(s)
	if err != nil {
		t.Error("Authority Test Failed when signing message")
	}

	msg, err := ack.MarshalForSignature()
	if err != nil {
		t.Error("Authority Test Failed when marshalling for sig")
	}

	sig := ack.GetSignature()
	server, err := s.Authorities[0].VerifySignature(msg, sig.GetSignature())
	if !server || err != nil {
		t.Error("Authority Test Failed when checking sigs")
	}
}

func TestHistoricKeyMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		hk := RandomHistoricKey()
		h, err := hk.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		hk2 := new(HistoricKey)
		err = hk2.UnmarshalBinary(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if hk.IsSameAs(hk2) == false {
			t.Errorf("Historic keys are not identical")
		}
	}
}

func TestAuthorityType(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	idindex := s.CreateBlankFactomIdentity(primitives.NewZeroHash())
	s.Identities[idindex].ManagementChainID = primitives.NewZeroHash()

	index := s.AddAuthorityFromChainID(primitives.NewZeroHash())
	s.Authorities[index].SigningKey = *(s.GetServerPublicKey())
	s.Authorities[index].Status = 1

	if s.Authorities[index].Type() != 1 {
		t.Error("Authority's Type isn't Leader even though Status is 1")
	}

	if s.GetAuthorityServerType(primitives.NewZeroHash()) != 0 {
		t.Error("GetAuthorityServerType isn't 0 even though authority is Leader")
	}

	s.Authorities[index].Status = 2

	if s.Authorities[index].Type() != 0 {
		t.Error("Authority's Type isn't Audit even though Status is 0")
	}

	if s.GetAuthorityServerType(primitives.NewZeroHash()) != 1 {
		t.Error("GetAuthorityServerType isn't 1 even though authority is Audit")
	}
}

func TestAuthorityRemoval(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	idindex := s.CreateBlankFactomIdentity(primitives.NewZeroHash())
	s.Identities[idindex].ManagementChainID = primitives.NewZeroHash()

	index := s.AddAuthorityFromChainID(primitives.NewZeroHash())
	s.Authorities[index].SigningKey = *(s.GetServerPublicKey())
	s.Authorities[index].Status = 1

	if !s.RemoveAuthority(primitives.NewZeroHash()) {
		t.Error("First call to RemoveAuthority unexpectedly failed")
	}

	if s.RemoveAuthority(primitives.NewZeroHash()) {
		t.Error("Second call to RemoveAuthority unexpectedly passed")
	}

	if s.GetAuthorityServerType(primitives.NewZeroHash()) >= 0 {
		t.Error("GetAuthorityServerType (after removal) >= 0")
	}
}
