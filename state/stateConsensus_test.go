// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/testHelper"
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

	s.Holding[msg.GetHash().Fixed()] = msg

	s.ReviewHolding()

	if len(s.XReview) > 0 {
		t.Error("300 ms hasn't passed;  should not execute")
	}

	// set the review to now
	s.ResendHolding = s.GetTimestamp()

	s.InMsgQueue().Enqueue(msg)

	s.ReviewHolding()

	if len(s.XReview) > 0 {
		t.Error("Have too many entries in the InMsgQueue;  should not execute")
	}

	s.InMsgQueue().Dequeue()

	s.ReviewHolding()

	if len(s.XReview) == 0 {
	//	t.Error("Should have queued a message for reprocessing")
	}

	s.ReviewHolding()

}
