package p2p

import (
	"math/rand"
	"testing"
)

func Test_LockSource(t *testing.T) {

	rng := rand.New(rand.NewSource(1))
	src, err := newLockSource(1)
	if err != nil {
		t.Fatal(err)
	}
	myrng := rand.New(src)

	for i := 0; i < 1000; i++ {
		a := rng.Uint64()
		b := myrng.Uint64()
		if a != b {
			t.Errorf("Uint64() mismatch in iteration %d: want = %d, got = %d", i, a, b)
		}
	}

	for i := 0; i < 1000; i++ {
		a := rng.Int63()
		b := myrng.Int63()
		if a != b {
			t.Errorf("Int63() mismatch in iteration %d: want = %d, got = %d", i, a, b)
		}
	}

}
