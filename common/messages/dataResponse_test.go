// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNilDataResponse(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DataResponse)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalDataResponse(t *testing.T) {
	msgs := []*DataResponse{newDataResponseEntry(), newDataResponseEntryBlock()}
	for _, msg := range msgs {
		hex, err := msg.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		t.Logf("Marshalled - %x", hex)

		msg2, err := msgsupport.UnmarshalMessage(hex)
		if err != nil {
			t.Error(err)
		}
		str := msg2.String()
		t.Logf("str - %v", str)

		if msg2.Type() != constants.DATA_RESPONSE {
			t.Error("Invalid message type unmarshalled")
		}

		hex2, err := msg2.(*DataResponse).MarshalBinary()
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

		if msg.IsSameAs(msg2.(*DataResponse)) != true {
			t.Errorf("DataResponse messages are not identical")
		}
	}
}

func newDataResponseEntry() *DataResponse {
	dr := new(DataResponse)
	dr.Timestamp = primitives.NewTimestampNow()
	dr.DataType = 0
	entry := testHelper.CreateFirstTestEntry()
	dr.DataObject = entry
	dr.DataHash = entry.GetHash()
	return dr
}

func newDataResponseEntryBlock() *DataResponse {
	dr := new(DataResponse)
	dr.Timestamp = primitives.NewTimestampNow()
	dr.DataType = 1
	entry, _ := testHelper.CreateTestEntryBlock(nil)
	dr.DataObject = entry
	dr.DataHash, _ = entry.KeyMR()
	return dr
}
