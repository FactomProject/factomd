package p2p

import (
	"crypto/sha1"
	"math/rand"
	"testing"
	"time"
)

func _newTestPeerResend(buckets int, interval time.Duration) *PeerResend {
	pr := new(PeerResend)
	pr.interval = interval
	pr.buckets = make([]*PRBucket, buckets)
	for i := 0; i < buckets; i++ {
		pr.buckets[i] = newPRBucket()
	}
	pr.stopper = make(chan interface{}, 1)
	return pr
}

func rhash(l int) [sha1.Size]byte {
	r := make([]byte, l)
	rand.Read(r)
	return sha1.Sum(r)
}

func TestPRBucket(t *testing.T) {
	bucket := newPRBucket()

	testset := make([][sha1.Size]byte, 128)
	for i := range testset {
		testset[i] = rhash(rand.Intn(512) + 64)
	}

	for _, h := range testset {
		if bucket.Has(h) {
			t.Error("empty bucket reports it has %h", h)
		}
	}

	for i := 0; i < 64; i++ {
		bucket.Add(testset[i])
	}

	for i, h := range testset {
		has := bucket.Has(h)
		if i >= 64 {
			if has {
				t.Error("full bucket reports it has %h when it shouldn't", h)
			}
		} else {
			if !has {
				t.Error("bucket should have %h but doesn't", h)
			}
		}
	}
}

func TestPRBucket_MultiThreaded_Try(t *testing.T) {

	pr := _newTestPeerResend(16, time.Hour)

	done := make(chan bool, 8)
	testroutine := func() {
		testset := make([][sha1.Size]byte, 1024)
		for i := range testset {
			testset[i] = rhash(rand.Intn(512) + 64)
			pr.Add(testset[i])
		}

		for _, p := range testset {
			if !pr.Has(p) {
				t.Errorf("data missing")
			}
		}

		done <- true
	}

	for i := 0; i < 8; i++ {
		go testroutine()
	}

	for i := 0; i < 8; i++ {
		<-done
	}
}

func TestPRBucket_MultiThreaded_Cleanup(t *testing.T) {
	testpl := sha1.Sum([]byte{0xff, 0x00, 0x00})

	pr := _newTestPeerResend(1, time.Hour)

	pr.Add(testpl)
	pr.dropOldestBucket()

	if pr.Has(testpl) {
		t.Errorf("single bucket didn't get cleaned properly")
	}

	pr = _newTestPeerResend(2, time.Hour)

	pr.Add(testpl)
	pr.dropOldestBucket()

	if !pr.Has(testpl) {
		t.Errorf("item not found but should still be in bucket #2")
	}

	pr.dropOldestBucket()
	if pr.Has(testpl) {
		t.Errorf("double bucket didn't get cleaned properly")
	}

	pr = NewPeerResend(3, time.Millisecond*50)

	pr.Add(testpl)

	time.Sleep(time.Millisecond * 75)
	if !pr.Has(testpl) {
		t.Errorf("timed item not found")
	}

	time.Sleep(time.Millisecond * 100) // 175 ms > 150 ms
	if pr.Has(testpl) {
		t.Errorf("timed bucket didn't get cleaned properly")
	}
	pr.Stop()
	pr.Add(testpl)

	time.Sleep(time.Millisecond * 200)
	if !pr.Has(testpl) {
		t.Errorf("timed bucket didn't get stopped properly")
	}
}
