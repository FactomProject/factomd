package primitives

import (
	"crypto/sha256"
)

type ProcessListLocation struct {
	Vm     int
	Minute int
	Height int
}

type AuthSet struct {
	IdentityList []Identity
	StatusArray  []int
	IdentityMap  map[Identity]int
}

func NewAuthSet() *AuthSet {
	a := new(AuthSet)
	a.New()
	return a
}

func (a *AuthSet) Sort() {
	for i := 1; i < len(a.IdentityList); i++ {
		for j := 0; j < len(a.IdentityList)-i; j++ {
			if a.IdentityList[j+1].Less(a.IdentityList[j]) {
				// Swap in both lists, change the index in the map
				a.IdentityList[j], a.IdentityList[j+1] = a.IdentityList[j+1], a.IdentityList[j]
				a.StatusArray[j], a.StatusArray[j+1] = a.StatusArray[j+1], a.StatusArray[j]
				a.IdentityMap[a.IdentityList[j]] = j
				a.IdentityMap[a.IdentityList[j+1]] = j + 1
			}
		}
	}
}

func (a *AuthSet) New() {
	a.IdentityList = make([]Identity, 0)
	a.StatusArray = make([]int, 0)
	a.IdentityMap = make(map[Identity]int)
}
func (a *AuthSet) Add(id Identity, status int) int {
	index := len(a.IdentityList)
	a.IdentityMap[id] = index
	a.IdentityList = append(a.IdentityList, id)
	a.StatusArray = append(a.StatusArray, status)

	// TODO: It should return the index right?
	return index
}

func (a *AuthSet) IsLeader(id Identity) bool {
	index, ok := a.IdentityMap[id]
	if !ok {
		panic("Bad Identity")
	}
	return a.StatusArray[index] > 0
}

type Identity int

func (a Identity) Less(b Identity) bool {
	return a < b
}

type Hash [sha256.Size]byte

// AuthSetHelper will help testing authority sets. Add whatever functions you need to help you
// manage the authsets in tests
type AuthSetHelper struct {
	feds int
	auds int
	AuthSet

	sequenceCounter int
}

func NewAuthSetHelper(feds, auds int) *AuthSetHelper {
	a := new(AuthSetHelper)
	a.AuthSet = *NewAuthSet()

	for i := 0; i < feds; i++ {
		a.AddFed()
	}
	for i := 0; i < feds; i++ {
		a.AddAudit()
	}

	return a
}

func (a *AuthSetHelper) NextIdentity() Identity {
	a.sequenceCounter++
	return Identity(a.sequenceCounter)
}

func (a *AuthSetHelper) AddFed() Identity {
	id := a.NextIdentity()
	a.Add(id, 1)
	return id
}

func (a *AuthSetHelper) AddAudit() Identity {
	id := a.NextIdentity()
	a.Add(id, 0)
	return id
}

func (a *AuthSetHelper) GetAuthSet() AuthSet {
	return a.AuthSet
}
