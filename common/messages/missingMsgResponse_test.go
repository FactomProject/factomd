// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

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

func TestSomething(t *testing.T) {
	General = new(msgsupport.GeneralFactory)
	primitives.General = General

	a := new(Bounce)
	a.Timestamp = primitives.NewTimestampNow()
	b := NewSignedAck()

	var buf primitives.Buffer
	buf.PushMsg(a)
	buf.PushMsg(b)
	buf2 := primitives.NewBuffer(buf.Bytes())
	a2, _ := buf2.PopMsg()
	b2, _ := buf2.PopMsg()

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
		m := NewMissingMsgResponse(s, b, NewSignedAck()).(*MissingMsgResponse)

		d, err := m.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		m1 := new(MissingMsgResponse)
		fmt.Printf("%x\n", d)
		nd, err := m1.UnmarshalBinaryData(d)
		if err != nil {
			t.Error(err)
		}

		if len(nd) > 0 {
			t.Errorf("Should not have leftover bytes, found %d", len(nd))
		}

		if !m.IsSameAs(m1) {
			t.Error("Unmarshal gave back a different message")
		}
		fmt.Println("************")
		fmt.Printf("%x\n", d)
		fmt.Println("************")
		d2, _ := m1.MarshalBinary()
		fmt.Printf("%x\n", d2)
		fmt.Println("************")
		d3, _ := m.AckResponse.MarshalBinary()
		d4, _ := m1.AckResponse.MarshalBinary()
		ak2 := NewSignedAck()
		ak2.UnmarshalBinary(d4)
		d5, _ := ak2.MarshalBinary()
		fmt.Println("************")
		fmt.Printf("%x\n", d3)
		fmt.Println("************")
		fmt.Printf("%x\n", d4)
		fmt.Println("************")
		fmt.Printf("%x\n", d5)
		fmt.Println("************")
	}

}

func TestSillyMarshaling(t *testing.T) {

	const cnt = 10
	b := new(Bounce)
	b.Timestamp = primitives.NewTimestampNow()

	var ack []*Ack
	var mmr []*MissingMsgResponse
	var data [][]byte
	var data2 [][]byte

	s := testHelper.CreateEmptyTestState()

	for i := 0; i < cnt; i++ {
		ack = append(ack, NewSignedAck())
		for j := 0; j < 8; j++ {
			ack[i].Salt[j] = byte(j + 1)
		}
		mmr = append(mmr, NewMissingMsgResponse(s, b, ack[i]).(*MissingMsgResponse))
		d, _ := mmr[i].MarshalBinary()
		mmr2 := NewMissingMsgResponse(s, b, ack[i]).(*MissingMsgResponse)
		mmr2.UnmarshalBinary(d)
		d2, _ := mmr2.MarshalBinary()
		data = append(data, d)
		data2 = append(data2, d2)
	}

	for i, d := range data {
		ack := mmr[i].AckResponse
		ad, _ := ack.MarshalBinary()
		var _ = ad
		var _ = d
		//fmt.Printf("mmr  %d %x\n", i, d)
		//fmt.Printf("ack  %d %x\n", i, ad)
		//fmt.Printf("mmr2 %d %x\n", i, data2[i])
	}

}
