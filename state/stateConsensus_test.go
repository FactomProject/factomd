// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/common/primitives"
	"time"
)


func TestReveiwHolding(t *testing.T) {

	s := testHelper.CreateAndPopulateTestState()

	// create us a cheap message
	msg := messages.NewDBStateMissing(s, 10, 20)
	for s.InMsgQueue().Length() < constants.INMSGQUEUE_LOW-1 {
		s.InMsgQueue().Enqueue(msg)
	}

	// get rid of any "leavings"
	s.XReview = s.XReview[:0]

	for k := range s.Holding {
		delete(s.Holding,k)
	}

	s.Holding[msg.GetHash().Fixed()] = msg

	{
		m := messages.NewServerFault(primitives.NewHash([]byte("ha")),primitives.NewHash([]byte("ha")),0,10,0,0,s.GetTimestamp())
		s.Holding[m.GetHash().Fixed()]=m
	}
	{
		m := messages.NewMissingMsgResponse(s,msg,msg)
		s.Holding[primitives.Sha([]byte("missing message")).Fixed()]= m
	}
	{
		sf := messages.NewServerFault(primitives.NewHash([]byte("ha")),primitives.NewHash([]byte("ha")),0,10,0,0,s.GetTimestamp())
		m := messages.NewFullServerFault(nil,sf,nil,0)
		s.Holding[m.GetHash().Fixed()]=m
	}

	s.ReviewHolding()

	if len(s.XReview) > 0 {
		t.Error("300 ms hasn't passed;  should not execute")
	}

	// set the review to now
	s.ResendHolding = s.GetTimestamp()

	time.Sleep(400 * time.Millisecond)

	// get rid of any "leavings"
	s.XReview = s.XReview[:0]

	s.InMsgQueue().Enqueue(msg)

	s.ReviewHolding()
	
	s.InMsgQueue().Dequeue()

	s.ReviewHolding()

	if len(s.XReview) == 0 {
		t.Error("Should have queued a message for reprocessing")
	}

	s.ReviewHolding()

}
