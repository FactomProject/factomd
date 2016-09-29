// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
)

var _ = fmt.Printf
var _ = rand.New

var hour = int64(60 * 60)

var r = Replay{}

func Test_Replay(test *testing.T) {

	type mh struct {
		hash [32]byte
		time interfaces.Timestamp
	}

	XTrans := 10240 //61440 //145000

	h := make([]*mh, XTrans)

	start := primitives.NewTimestampNow()
	now := primitives.NewTimestampNow()

	for i := 0; i < XTrans; i++ {

		if i%10240 == 0 {
			fmt.Println("Testing ", i)
		}

		// We are going to remember some large set of transactions.
		h[i] = new(mh)
		h[i].hash = primitives.Sha([]byte(fmt.Sprintf("h%d", i))).Fixed()

		// Build a valid transaction somewhere +/- 12 hours of now
		h[i].time = primitives.NewTimestampFromSeconds(uint32(now.GetTimeSeconds() + (rand.Int63() % 24 * hour) - 12*hour))

		// The first time we test, it should be valid.
		if !r.IsTSValid_(constants.INTERNAL_REPLAY, h[i].hash, h[i].time, now) {
			fmt.Println("Failed Test ", i, "first")
			test.Fail()
			return
		}

		// An immediate replay!  Should fail!
		if r.IsTSValid_(constants.INTERNAL_REPLAY, h[i].hash, h[i].time, now) {
			fmt.Println("Failed Test ", i, "second")
			test.Fail()
			return
		}

		// Move time forward somewhere between 0 to 15 minutes
		now = primitives.NewTimestampFromSeconds(uint32(now.GetTimeSeconds() + rand.Int63()%hour/4))

		// Now replay all the transactions we have collected.  NONE of them
		// should work.
		for j := 0; j < i; j++ {
			if r.IsTSValid_(constants.INTERNAL_REPLAY, h[i].hash, h[i].time, now) {
				fmt.Println("Failed Test ", i, j, "repeat")
				test.Fail()
				return
			}
		}
	}

	fmt.Println("Simulation ran from", time.Unix(start.GetTimeSeconds(), 0), "to", time.Unix(now.GetTimeSeconds(), 0))

}
