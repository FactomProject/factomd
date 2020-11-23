// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"io/ioutil"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/identity"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"

	"github.com/FactomProject/factomd/common/messages/msgsupport"
	log "github.com/sirupsen/logrus"
)

func TestUnmarshalNilDirectoryBlockSignature(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DirectoryBlockSignature)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalDirectoryBlockSignature(t *testing.T) {
	msg := newDirectoryBlockSignature()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error("#1 ", err)
	}
	t.Logf("Marshalled - %x", hex)

	msg2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error("#2 ", err)
	}
	str := msg2.String()
	t.Logf("str - %v", str)

	if msg2.Type() != constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*DirectoryBlockSignature).MarshalBinary()
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

	if msg.IsSameAs(msg2.(*DirectoryBlockSignature)) != true {
		t.Errorf("DirectoryBlockSignature messages are not identical")
	}
}

func TestSignAndVerifyDirectoryBlockSignature(t *testing.T) {
	dbs, _, _ := newSignedDirectoryBlockSignature()

	hex, err := dbs.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	t.Logf("Sig - %x", *dbs.Signature.GetSignature())
	if len(*dbs.Signature.GetSignature()) == 0 {
		t.Error("Signature not present")
	}

	valid, err := dbs.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature is not valid")
	}

	dbs2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}

	if dbs2.Type() != constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
		t.Error("Invalid message type unmarshalled")
	}
	dbsProper := dbs2.(*DirectoryBlockSignature)

	valid, err = dbsProper.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature 2 is not valid")
	}
}

func TestInvalidSignature(t *testing.T) {
	s := testHelper.CreateEmptyTestState()
	m, a, key := newSignedDirectoryBlockSignature()
	s.IdentityControl.SetAuthority(a.AuthorityChainID, a)

	data, err := m.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2 := new(DirectoryBlockSignature)
	err = m2.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
	}

	s.ProcessLists.Get(s.PLProcessHeight).AddFedServer(m.ServerIdentityChainID)

	v := m.Validate(s)
	// Should be 0 as the authority set is unknown
	if v != 1 {
		t.Errorf("Expected 1, found %d", v)
	}
	m.Sigvalid = false

	// Now an invalid sig
	m2.DBHeight = m.DBHeight + 1
	err = m2.Sign(key)
	if err != nil {
		panic(err)
	}

	m.Signature = m2.Signature       // make message signature bad
	s.LLeaderHeight = m.DBHeight - 1 // make message in the future
	v = m.Validate(s)

	if v != -2 {
		t.Errorf("Expected -2, found %d", v)
	}
	s.LLeaderHeight = m.DBHeight // make message current but a bad signature
	v = m.Validate(s)
	if v != -1 {
		t.Errorf("Expected -1, found %d", v)
	}
}

func newDirectoryBlockSignature() *DirectoryBlockSignature {
	dbs := new(DirectoryBlockSignature)
	dbs.DBHeight = 123
	//hash, _ := primitives.NewShaHashFromStr("cbd3d09db6defdc25dfc7d57f3479b339a077183cd67022e6d1ef6c041522b40")
	//dbs.DirectoryBlockKeyMR = hash
	hash, _ := primitives.NewShaHashFromStr("a077183cd67022e6d1ef6c041522b40cbd3d09db6defdc25dfc7d57f3479b339")
	dbs.ServerIdentityChainID = hash
	tmp := directoryBlock.NewDBlockHeader()
	dbs.DirectoryBlockHeader = tmp
	return dbs
}

func newSignedDirectoryBlockSignature() (*DirectoryBlockSignature, *identity.Authority, *primitives.PrivateKey) {
	dbs := newDirectoryBlockSignature()
	dbs.SetValid()
	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = dbs.Sign(key)
	if err != nil {
		panic(err)
	}

	a := new(identity.Authority)
	a.SigningKey = *(key.Pub)
	a.AuthorityChainID = dbs.ServerIdentityChainID
	return dbs, a, key
}

// go test -bench=. directoryBlockSignature_test.go  -v

//BenchmarkValidateMakingFunction tests the creating of the log function and NOT using it.
//
// 2000000000	         1.99 ns/op
func BenchmarkValidateMakingFunctionNoUse(b *testing.B) {
	s := testHelper.CreateEmptyTestState()
	m, _, _ := newSignedDirectoryBlockSignature()
	for i := 0; i < b.N; i++ {
		//m.Validate(s)
		vlog := func(format string, args ...interface{}) {
			log.WithFields(log.Fields{"msgheight": m.DBHeight, "lheight": s.GetLeaderHeight()}).Errorf(format, args...)
		}
		var _ = vlog
	}
}

//BenchmarkValidateMakingFunction tests the creating of the logger function and NOT using it.
//
//  1000000	      1312 ns/op
func BenchmarkValidateMakingInstantiateNoUse(b *testing.B) {
	s := testHelper.CreateEmptyTestState()
	m, _, _ := newSignedDirectoryBlockSignature()
	for i := 0; i < b.N; i++ {
		//m.Validate(s)
		vlog := log.WithFields(log.Fields{"msgheight": m.DBHeight, "lheight": s.GetLeaderHeight()})
		var _ = vlog
	}
}

//BenchmarkValidateMakingFunctionUse tests the creating of the log function and using it.
//
// To ioutil.Discard
//  100000	     27277 ns/op
//
// Printing to stdout
//  30000	     60523 ns/op

func BenchmarkValidateMakingFunctionUse(b *testing.B) {
	s := testHelper.CreateEmptyTestState()
	m, _, _ := newSignedDirectoryBlockSignature()
	log.SetOutput(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		//m.Validate(s)
		vlog := func(format string, args ...interface{}) {
			log.WithFields(log.Fields{"msgheight": m.DBHeight, "lheight": s.GetLeaderHeight()}).Errorf(format, args...)
		}
		vlog("%s", "hello")
	}
}
