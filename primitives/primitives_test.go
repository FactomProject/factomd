package primitives_test

import (
	. "github.com/FactomProject/electiontesting/primitives"
	"testing"
)

func TestIsLeader(t *testing.T) {
	a := NewAuthSet()
	audits := []Identity{0, 1, 2, 3, 4}
	feds := []Identity{5, 6, 7, 8}

	for _, aud := range audits {
		a.Add(aud, 0)
	}
	for _, fed := range feds {
		a.Add(fed, 1)
	}

	for _, aud := range audits {
		if a.IsLeader(aud) {
			t.Error("Audit reported as fed")
		}
	}
	for _, fed := range feds {
		if !a.IsLeader(fed) {
			t.Error("Fed reported as audit")
		}
	}
}

func TestAuthSetSort(t *testing.T) {
	a := NewAuthSetHelper(10, 10)
	a.Sort()
	for i, v := range a.IdentityList {
		if int(v) != i+1 {
			t.Error("Sort did not work")
		}
	}
}
