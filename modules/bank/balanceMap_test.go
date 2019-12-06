package state_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/primitives/random"

	. "github.com/FactomProject/factomd/state"
)

func TestBalanceMap(t *testing.T) {
	t.Run("Test correct balances", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			bm := NewBalanceMap()
			bms := bm.ThreadSafe()
			go bms.Serve()

			ds := randomSafeDeltas(random.RandIntBetween(70, 150))
			bals, err := actualBals(ds)
			if err != nil {
				t.Errorf("Test code is incorrect")
			}

			err = bms.UpdateBalancesFromDeltas(ds)
			if err != nil {
				t.Error(err)
			}
			if !sameBals(bals, trim(bm.AllBalances())) {
				t.Errorf("Should be the same balances")
			}
			bm.Close()
		}
	})

	t.Run(("Test incorrect balances and rollback"), func(t *testing.T) {
		for i := 0; i < 100; i++ {
			bm := NewBalanceMap()
			bms := bm.ThreadSafe()
			go bms.Serve()

			ds := randomSafeDeltas(100)
			injection := random.RandIntBetween(0, len(ds))
			bals, err := actualBals(ds[:injection])
			if err != nil {
				t.Errorf("Test code is incorrect")
			}

			// Grab random addr
			var key [32]byte
			for k := range bals {
				key = k
				break
			}

			ds[injection].Address = key
			ds[injection].Delta = -1 * (bals[key] + 1)

			_, err = actualBals(ds)
			if err == nil {
				t.Errorf("Test code is incorrect")
			}

			err = bms.UpdateBalancesFromDeltas(ds)
			if err == nil {
				t.Errorf("Should have an error")
			}

			all := bm.AllBalances()
			for _, b := range all {
				if b != 0 {
					t.Errorf("Exp 0 balance, found %d", b)
				}
			}

			// Now apply up to some point before injection, and then rollback
			stop := random.RandIntBetween(0, injection)
			bals, err = actualBals(ds[:stop]) // Bals we will rollback to
			if err != nil {
				t.Errorf("Test code is incorrect")
			}

			err = bm.UpdateBalancesFromDeltas(ds[:stop])
			if !sameBals(bm.AllBalances(), trim(bals)) || err != nil {
				t.Error("Should be the same balances")
			}

			err = bm.UpdateBalancesFromDeltas(ds[stop:]) // Should error out and rollback
			if err == nil {
				t.Errorf("Should have an error")
			}
			//bm.Trim()
			if !sameBals(bm.AllBalances(), trim(bals)) {
				t.Error("Should be the same balances after rollback")
			}

			bms.Close()

		}
	})
}

func sameBals(bal1, bal2 map[[32]byte]int64) bool {
	if len(bal1) != len(bal2) {
		return false
	}

	for k, _ := range bal1 {
		if bal1[k] != bal2[k] {
			return false
		}
	}
	return true
}

func trim(bals map[[32]byte]int64) map[[32]byte]int64 {
	for k, v := range bals {
		if v == 0 {
			delete(bals, k)
		}
	}
	return bals
}

func actualBals(deltas []Delta) (map[[32]byte]int64, error) {
	bals := make(map[[32]byte]int64)
	var err error
	for _, d := range deltas {
		bals[d.Address] += d.Delta
		if bals[d.Address] < 0 {
			err = fmt.Errorf("bal goes below 0")
		}
	}
	return bals, err
}

// randomSafeDeltas ensures no bals go below 0
func randomSafeDeltas(size int) []Delta {
	bals := make(map[[32]byte]int64)

	ds := make([]Delta, size)
DeltaLoop:
	for i := range ds {
		if i%3 == 0 {
			// Try a decrement
			for k, v := range bals {
				if v > 0 {
					ds[i].Address = k
					ds[i].Delta = -1 * random.RandInt64Between(0, v)
					bals[ds[i].Address] += ds[i].Delta
					continue DeltaLoop
				}
			}
		}
		// Increment
		ds[i].Address = randAddr()
		ds[i].Delta = random.RandInt64Between(0, 1e12)
		bals[ds[i].Address] += ds[i].Delta
	}
	return ds
}

func randAddr() [32]byte {
	var answer [32]byte
	_, err := rand.Read(answer[:])
	if err != nil {
		return [32]byte{}
	}
	return answer
}
