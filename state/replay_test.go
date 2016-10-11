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
	"log"
	"net/http"
	"net/http/pprof"
)

var _ = fmt.Printf
var _ = rand.New

var hour = int64(60 * 60)

var r = Replay{}

var span int64 = 1     // How many hours +/- that will be valid (1== valid 1 hour in the past to 1 hour in the future)
var speed int64 = 1000 // Speed in milliseconds (max) that we will move the clock
var _ = pprof.Cmdline

func Test_Replay(test *testing.T) {

	type mh struct {
		hash [32]byte
		time interfaces.Timestamp
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	XTrans := 102400 //61440 //145000

	var mhs []*mh

	start := primitives.NewTimestampNow()
	now := primitives.NewTimestampNow()

	for i := 0; i < XTrans; i++ {

		if (i+1)%1000 == 0 {

			buckets := len(r.Buckets)
			bcnt := 0
			for _, b := range r.Buckets {
				bcnt += len(b)
			}
			secs := now.GetTimeSeconds() - start.GetTimeSeconds()
			fmt.Printf("Test in %d minute, %d buckets, %d elements, %d center\n", Minutes(secs), buckets, bcnt, r.Center)
			// Now replay all the transactions we have collected.  NONE of them
			// should work.

			for in := 0; in < 1000; in++ {
				i := rand.Int() % len(mhs)
				x := mhs[i]
				if r.IsTSValid_(constants.INTERNAL_REPLAY, x.hash, x.time, now) {
					fmt.Printf("Failed Test ", i, "repeat")
					test.Fail()
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
		}

		if (i+1)%10000 == 0 {
			mhs = mhs[:0]
		}

		// We are going to remember some large set of transactions.
		x := new(mh)
		x.hash = primitives.Sha([]byte(fmt.Sprintf("h%d", i))).Fixed()
		mhs = append(mhs, x)

		// Build a valid transaction somewhere +/- span hours of now
		delta := (span*60*2 - 2)
		ntime := now.GetTimeSeconds() - span*60 + 1
		x.time = primitives.NewTimestampFromSeconds(uint32(ntime + delta))

		if _, ok := r.Valid(constants.INTERNAL_REPLAY, x.hash, x.time, now); !ok {
			xsec := x.time.GetTimeSeconds()
			nsec := now.GetTimeSeconds()
			fmt.Println("Failed Test ", i, "pre-test", Minutes(xsec), Minutes(nsec), Minutes(xsec-nsec))
			test.Fail()
			return
		}

		// The first time we test, it should be valid.
		if !r.IsTSValid_(constants.INTERNAL_REPLAY, x.hash, x.time, now) {
			fmt.Println("Failed Test ", i, "first")
			test.Fail()
			return
		}

		// An immediate replay!  Should fail!
		if r.IsTSValid_(constants.INTERNAL_REPLAY, x.hash, x.time, now) {
			fmt.Println("Failed Test ", i, "second")
			test.Fail()
			return
		}

		x = new(mh)
		x.hash = primitives.Sha([]byte(fmt.Sprintf("xh%d", i))).Fixed()

		// Build a invalid transaction > span in the past
		x.time = primitives.NewTimestampFromSeconds(uint32(now.GetTimeSeconds() - span*hour - ((rand.Int63() % (100)) + 1)))

		// should not be valid
		if _, ok := r.Valid(constants.INTERNAL_REPLAY, x.hash, x.time, now); ok {
			fmt.Println("Failed Test ", i, "bad-pre-test", Minutes(now.GetTimeSeconds()), Minutes(x.time.GetTimeSeconds()))
			test.Fail()
			return
		}

		// The first time we test, it should not be valid.
		if r.IsTSValid_(constants.INTERNAL_REPLAY, x.hash, x.time, now) {
			fmt.Println("Failed Test ", i, "bad first")
			test.Fail()
			return
		}

		// Move time forward somewhere between 0 to 15 minutes
		now = primitives.NewTimestampFromMilliseconds(uint64(now.GetTimeMilli() + rand.Int63()%(speed)))

	}

	fmt.Println("Simulation ran from", time.Unix(start.GetTimeSeconds(), 0), "to", time.Unix(now.GetTimeSeconds(), 0))

}
