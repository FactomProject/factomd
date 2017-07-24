// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	_, _, err := UnmarshalMessageData(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	_, _, err = UnmarshalMessageData([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	_, _, err = UnmarshalMessageData([]byte{0xFF})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestUnmarshalMsgTypes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	s := testHelper.CreateEmptyTestState()

	err := testUnmarshalMsg(NewMissingMsg(s, 0, 0, 0))
	if err != nil {
		t.Error(err)
	}

	b := new(Bounce)
	b.Stamps = []interfaces.Timestamp{}
	b.Timestamp = primitives.NewTimestampNow()
	err = testUnmarshalMsg(b)
	if err != nil {
		t.Error(err)
	}

	br := new(BounceReply)
	br.Stamps = []interfaces.Timestamp{}
	br.Timestamp = primitives.NewTimestampNow()
	err = testUnmarshalMsg(br)
	if err != nil {
		t.Error(err)
	}
}

func testUnmarshalMsg(m interfaces.IMsg) error {
	d, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	nd, nm, err := UnmarshalMessageData(d)
	if err != nil {
		return err
	}

	if len(nd) > 0 {
		return fmt.Errorf("%d bytes remain after unmarshal", len(nd))
	}

	d2, err := nm.MarshalBinary()
	if err != nil {
		return err
	}

	if bytes.Compare(d, d2) != 0 {
		return fmt.Errorf("New message differs from original")
	}

	return nil
}
