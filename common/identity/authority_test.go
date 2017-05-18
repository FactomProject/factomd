// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAuthorityType(t *testing.T) {
	auth := new(Authority)
	if auth.Type() != -1 {
		t.Errorf("Invalid type returned - %v", auth.Type())
	}

	auth.Status = constants.IDENTITY_FEDERATED_SERVER
	if auth.Type() != 1 {
		t.Errorf("Invalid type returned - %v", auth.Type())
	}

	auth.Status = constants.IDENTITY_AUDIT_SERVER
	if auth.Type() != 0 {
		t.Errorf("Invalid type returned - %v", auth.Type())
	}

}

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

func TestMarshalJSON(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	idindex := s.CreateBlankFactomIdentity(primitives.NewZeroHash())
	s.Identities[idindex].ManagementChainID = primitives.NewZeroHash()

	index := s.AddAuthorityFromChainID(primitives.NewZeroHash())
	s.Authorities[index].SigningKey = *(s.GetServerPublicKey())
	s.Authorities[index].Status = 1

	j, err := s.Authorities[index].MarshalJSON()
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := `{"chainid":"0000000000000000000000000000000000000000000000000000000000000000","manageid":"0000000000000000000000000000000000000000000000000000000000000000","matroyshka":null,"signingkey":"cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a","status":"federated","anchorkeys":null}`
	if string(j) != expected {
		t.Errorf("Invalid json returned - %v vs %v", string(j), expected)
	}
}
