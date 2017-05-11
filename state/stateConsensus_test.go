// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = time.Now
var _ = fmt.Print
var _ = ESAsking

func TestHolding(t *testing.T) {
	s := testHelper.CreateEmptyTestState()

	msgs := makeXBounces(100)
	for _, m := range msgs {
		m.SetValid(0)
		s.MsgQueue() <- m
	}

	s.Process()
	if len(s.Holding)+len(s.XReview) != 100 {
		t.Errorf("Holding should have 100 elements, found %d and %d in XReview", len(s.Holding), len(s.XReview))
	}

	for _, m := range msgs {
		if m.Processed() {
			t.Error("Should not be processed but is")
		}
	}

	for _, m := range msgs {
		m.SetValid(1)
	}

	time.Sleep(300 * time.Millisecond)
	s.ReviewHolding()
	s.Process()
	if len(s.Holding)+len(s.XReview) != 0 {
		t.Errorf("Holding should have 0 elements, found %d and %d in XReview", len(s.Holding), len(s.XReview))
	}

	for _, m := range msgs {
		if !m.Processed() {
			t.Error("Should be processed but is not")
		}
	}
}

func makeXBounces(x int) []*messages.Bounce {
	arr := make([]*messages.Bounce, x)
	for i := range arr {
		arr[i] = makeBounce()
	}
	return arr
}

func makeBounce() *messages.Bounce {
	b := new(messages.Bounce)
	b.Timestamp = primitives.NewTimestampNow()
	b.Stamps = nil
	b.Name = random.RandomString()
	return b
}
