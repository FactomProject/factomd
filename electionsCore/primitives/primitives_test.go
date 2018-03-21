package primitives_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/electionsCore/primitives"
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

/*
func TestAuthSetSort(t *testing.T) {
	a := NewAuthSetHelper(10, 10)
	a.Sort()
	for i, v := range a.IdentityList {
		if int(v) != i+1 {
			t.Error("Sort did not work")
		}
	}
}
*/

func TestIdentityReadString(t *testing.T) {

	var i Identity
	s := "ID-89abcdef"
	i.ReadString(s)
	if i.String() != s {
		t.Errorf("Identity.ReadString(\"%s\")", s)
	}
}

func TestHashReadString(t *testing.T) {

	var h Hash
	s := "-000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f-"
	h.ReadString(s)
	r := h.String()
	if r != s {
		t.Errorf("Hash.ReadString(\"%s\")", s)
	}
}

func TestProcessListLocationReadString(t *testing.T) {
	var p ProcessListLocation
	s := "1/2/3"
	p.ReadString(s)
	r := p.String()
	if r != s {
		t.Errorf("ProcessListLocation.ReadString(\"%s\")", s)
	}
}

func TestAuthSetReadString(t *testing.T) {
	var a AuthSet
	var id Identity
	id.ReadString("ID-76543210")
	a.New()
	a.Add(id, 1)
	id.ReadString("ID-01234567")
	a.Add(id, 0)
	s := a.String()
	fmt.Print(s)
}

func TestAuthorityStatusReadString(t *testing.T) {
	var a AuthorityStatus

	for n, s := range []string{"AUDIT", "LEADER", "INVALID:2"} {
		a.ReadString(s)
		if a != AuthorityStatus(n) {
			t.Errorf("AuthorityStatus.ReadString(\"%s\")", s)
		}
		a = AuthorityStatus(n) // just in case the rad failed
		if a.String() != s {
			t.Errorf("AuthorityStatus.String() = \"%s\"", s)
		}
	}

}
