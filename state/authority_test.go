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

func TestAuthorityMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		a := RandomAuthority()

		h, err := a.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		a2 := new(Authority)
		err = a2.UnmarshalBinary(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if a.IsSameAs(a2) == false {
			t.Errorf("Authorities are not identical")
		}
	}
}
