package bmv

import (
	"math/rand"
	"testing"
	"time"
)

func Test_bucket_Get(t *testing.T) {
	const CASES = 1024
	now := time.Now()
	bucket := newBucket()
	bucket.SetTime(now)

	keys := make([][32]byte, CASES)
	times := make([]time.Time, CASES)

	for i := 0; i < CASES; i++ {
		rand.Read(keys[i][:])
		times[i] = now.Add(time.Duration(rand.Intn(60000)) * time.Millisecond)
		bucket.Set(keys[i], times[i])
	}

	for i, k := range keys {
		got, ok := bucket.Get(k)
		if !ok {
			t.Errorf("entry %d=%x not found", i, k)
			continue
		}
		if !got.Equal(times[i]) {
			t.Errorf("mismatching entries for %d=%x. got = %x, want = %x", i, k, got, times[i])
		}
	}

	var r [32]byte
	for i := 0; i < CASES; i++ {
		rand.Read(r[:])
		if _, ok := bucket.Get(r); ok {
			t.Errorf("found a random entry that shouldn't be there: %x", r)
		}
	}
}

func Test_bucket_Transfer(t *testing.T) {
	A, B := newBucket(), newBucket()

	A.SetTime(time.Now())
	A.SetTime(time.Now().Add(time.Minute))

	threshold := B.Time()

	const CASES = 32

	keys := make([][32]byte, CASES*2)
	times := make([]time.Time, CASES*2)

	// half above, half below
	for i := 0; i < CASES; i++ {
		rand.Read(keys[i][:])
		times[i] = threshold.Add(time.Duration(rand.Intn(60000)) * time.Millisecond)
		B.Set(keys[i], times[i])
	}

	for i := CASES; i < CASES*2; i++ {
		rand.Read(keys[i][:])
		times[i] = threshold.Add(time.Duration(rand.Intn(60000)) * -time.Millisecond)
		B.Set(keys[i], times[i])
	}

	A.Transfer(B)

	for i := 0; i < CASES; i++ {
		if _, ok := B.Get(keys[i]); !ok {
			t.Errorf("entry %d is not in bucket B", i)
		}
		if _, ok := A.Get(keys[i]); ok {
			t.Errorf("entry %d is erroneously in bucket A", i)
		}
	}

	for i := CASES; i < CASES*2; i++ {
		if _, ok := A.Get(keys[i]); !ok {
			t.Errorf("entry %d is not in bucket A", i)
		}
		if _, ok := B.Get(keys[i]); ok {
			t.Errorf("entry %d is erroneously in bucket B", i)
		}
	}

}
