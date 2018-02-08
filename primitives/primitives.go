package primitives

import (
	"crypto/sha256"
	"fmt"
	. "github.com/FactomProject/electiontesting/errorhandling"
	"encoding/hex"
)

type ProcessListLocation struct {
	Vm     int
	Minute int
	Height int
}

func (p *ProcessListLocation) String() string {
	return fmt.Sprintf("%d/%d/%d", p.Height,p.Minute,p.Vm)
}

func (p *ProcessListLocation) ReadString(s string) {
	n,err := fmt.Sscanf(s,"%d/%d/%d", &p.Height,&p.Minute,&p.Vm)
	if err != nil || n != 3 {
		HandleErrorf("ProcessListLocation.ReadString(%v) failed: %d %v",s,n,err)
	}
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
			if a.IdentityList[j+1].less(a.IdentityList[j]) {				// Swap in both lists, change the index in the map
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

func (i *Identity)String() string{
	return fmt.Sprintf("ID-%08x", *i)
}

func (i *Identity)ReadString(s string) {
	n,err:= fmt.Sscanf(s,"ID-%x", i)
	if err != nil || n != 1 {
		HandleErrorf("Identity.ReadString(%v) failed: %d %v",s,n,err)
	}
}

//todo:  Hmm, this only makes sense in the context of an round so it has to know the authset and the dbsig and the round -- clay
func (a Identity) less(b Identity) bool { 	return a < b}

type Hash [sha256.Size]byte

func (h *Hash) String()string{
	return fmt.Sprintf("-%s-", hex.EncodeToString(h[:]))
}

func (h *Hash) ReadString(s string) {
	n, err:= fmt.Sscanf(s,"-%[^-]s",&s) // drop the delimiters
	if(err != nil || n != 1) {
		HandleErrorf("Identity.ReadString(%v) failed: %d %v",s,n,err)
	}
	b, err := hex.DecodeString(s) // decode the hash in hex
	n = len(b)
	if(err != nil || n !=sha256.Size) {
		HandleErrorf("Identity.ReadString(%v) failed: %d %v",s,n,err)
	}
	copy(h[:],b[:])
}
