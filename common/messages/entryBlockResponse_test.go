// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"fmt"
	"strings"

	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNilEntryBlockResponse(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(EntryBlockResponse)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestNewEntryBlockResponse(t *testing.T) {

	state := testHelper.CreateAndPopulateTestState()
	msg := NewEntryBlockResponse(state)
	response := msg.String()
	timestamp := fmt.Sprintf("%d", msg.GetTimestamp().GetTimeMilli())
	//fmt.Println(timestamp)
	//fmt.Println(response)
	if strings.IndexAny(response, timestamp) == -1 {
		t.Errorf("Error timestamp not found in entryblock message")
	}
	timeremoved := strings.Replace(response, timestamp, "", 1)
	exampleTimeRemoved := "{\"FullMsgHash\":null,\"Origin\":0,\"NetworkOrigin\":\"\",\"Peer2Peer\":true,\"LocalOnly\":false,\"NoResend\":false,\"ResendCnt\":0,\"LeaderChainID\":null,\"MsgHash\":null,\"RepeatHash\":null,\"VMIndex\":0,\"VMHash\":null,\"Minute\":0,\"Ack\":null,\"Stalled\":false,\"MarkInvalid\":false,\"Sigvalid\":false,\"Timestamp\":,\"EBlockCount\":0,\"EBlocks\":null,\"EntryCount\":0,\"Entries\":null}"
	if timeremoved != exampleTimeRemoved {
		error := "Error entryblock message differed\n" + timeremoved + "\n" + exampleTimeRemoved
		t.Errorf(error)
	}

}
