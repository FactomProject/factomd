// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
		"testing"
"fmt"

	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
)

var _ = fmt.Print

func TestBadUnmarshal(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	m := new(MissingMsgResponse)
	err := m.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Should error")
	}
}

func TestSomething(t *testing.T){
	General = new(msgsupport.GeneralFactory)
	primitives.General = General

	a := new(Bounce)
	a.Timestamp = primitives.NewTimestampNow()
	b := NewSignedAck()

	var buf primitives.Buffer
	buf.PushMsg(a)
	buf.PushMsg(b)
	buf2 := primitives.NewBuffer(buf.Bytes())
	a2,_ := buf2.PopMsg()
	b2,_ := buf2.PopMsg()

	fmt.Println(a.String(), a2.String())
	fmt.Println(b.String(), b2.String())
}

func TestMissingMessageResponseMarshaling(t *testing.T) {

	General = new(msgsupport.GeneralFactory)
	primitives.General = General

	s := testHelper.CreateEmptyTestState()

	for i := 0; i < 1; i++ {
		b := new(Bounce)
		b.Timestamp = primitives.NewTimestampNow()
		m := NewMissingMsgResponse(s, b, NewSignedAck())
		m.GetHash()
		m.GetMsgHash()
		d, err := m.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		m2 := new(MissingMsgResponse)
		fmt.Printf("%x\n",d)
		nd, err := m2.UnmarshalBinaryData(d)
		if err != nil {
			t.Error(err)
		}

		if len(nd) > 0 {
			t.Errorf("Should not have leftover bytes, found %d", len(nd))
		}

		m1 := m.(*MissingMsgResponse)
		if !m1.IsSameAs(m2) {
			t.Error("Unmarshal gave back a different message")
		}
	}

}
