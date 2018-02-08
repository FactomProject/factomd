package primitives

import (
	"crypto/sha256"
	"fmt"
)

type ProcessListLocation struct {
	Vm int
	MinuteLocation
}

type MinuteLocation struct {
	Minute int
	Height int
}

func (m *MinuteLocation) Sum() int {
	return m.Height + m.Minute
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

func (a *AuthSet) GetFeds() []Identity {
	var feds []Identity
	for i, id := range a.IdentityList {
		if a.StatusArray[i] > 0 {
			feds = append(feds, id)
		}
	}
	return feds
}

func (a *AuthSet) GetAuds() []Identity {
	var auds []Identity
	for i, id := range a.IdentityList {
		if a.StatusArray[i] <= 0 {
			auds = append(auds, id)
		}
	}
	return auds
}

func (a *AuthSet) VMForIdentity(id Identity, location MinuteLocation) int {
	count := -1
	feds := a.GetFeds()
	for i, f := range feds {
		if f == id {
			count = i
			break
		}
	}
	if count == -1 {
		return -1
	}

	return (location.Sum() + count) % NumberOfMinutes
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

func (a *AuthSet) Majority() int {
	totalf := 0
	for _, s := range a.StatusArray {
		if s > 0 {
			totalf++
		}
	}
	return totalf/2 + 1
}

func (a *AuthSet) IsLeader(id Identity) bool {
	index, ok := a.IdentityMap[id]
	if !ok {
		panic("Bad Identity")
	}
	return a.StatusArray[index] > 0
}

func (a *AuthSet) Hash() Hash {
	str := ""
	for _, i := range a.IdentityList {
		str += fmt.Sprintf("%d%d", i, a.IdentityMap[i])
	}

	return sha256.Sum256([]byte(str))
}

type Identity int

func (a Identity) Less(b Identity) bool {
	return a < b
}

type Hash [sha256.Size]byte
