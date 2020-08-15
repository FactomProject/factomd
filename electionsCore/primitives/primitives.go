package primitives

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"bytes"

	"github.com/PaulSnow/factom2d/common/interfaces"
	. "github.com/PaulSnow/factom2d/electionsCore/errorhandling"
)

var hashRegEx *regexp.Regexp

func init() {
	hashRegEx = regexp.MustCompile("-([0-9a-zA-Z]+)-") // RegEx to extra a hash from a string
}

type MinuteLocation struct {
	Minute int
	Height int
}

func (m *MinuteLocation) Sum() int {
	return m.Height + m.Minute
}

func (p *MinuteLocation) String() string {
	return fmt.Sprintf("%d/%d", p.Height, p.Minute)
}

func (p *MinuteLocation) ReadString(s string) {
	n, err := fmt.Sscanf(s, "%d/%d", &p.Height, &p.Minute)
	if err != nil || n != 2 {
		HandleErrorf("MinuteLocation.ReadString(%v) failed: %d %v", s, n, err)
	}
}

type ProcessListLocation struct {
	Vm int
	MinuteLocation
}

func (p *ProcessListLocation) String() string {
	return fmt.Sprintf("%d/%d/%d", p.Height, p.Minute, p.Vm)
}

func (p *ProcessListLocation) ReadString(s string) {
	n, err := fmt.Sscanf(s, "%d/%d/%d", &p.Height, &p.Minute, &p.Vm)
	if err != nil || n != 3 {
		HandleErrorf("ProcessListLocation.ReadString(%v) failed: %d %v", s, n, err)
	}
}

type AuthSet struct {
	IdentityList []Identity
	StatusArray  []int
	IdentityMap  map[Identity]int

	PriorityMap           map[Identity]int
	PriorityToIdentityMap map[int]Identity
}

func (a AuthSet) Copy() AuthSet {
	b := new(AuthSet)
	b.IdentityMap = make(map[Identity]int)
	b.PriorityMap = make(map[Identity]int)
	b.PriorityToIdentityMap = make(map[int]Identity)
	b.IdentityList = make([]Identity, len(a.IdentityList))
	b.StatusArray = make([]int, len(a.StatusArray))

	for k, v := range a.IdentityMap {
		b.IdentityMap[k] = v
	}

	for i, v := range a.IdentityList {
		b.IdentityList[i] = v
	}

	for i, v := range a.StatusArray {
		b.StatusArray[i] = v
	}

	for k, v := range a.PriorityMap {
		b.PriorityMap[k] = v
	}

	for k, v := range a.PriorityToIdentityMap {
		b.PriorityToIdentityMap[k] = v
	}

	return *b
}

func (r *AuthSet) String() string {
	rval, err := json.Marshal(r)
	if err != nil {
		HandleErrorf("%T.String(...) failed: %v", r, err)
	}
	return string(rval[:])
}

func (r *AuthSet) ReadString(s string) {
	err := json.Unmarshal([]byte(s), r)
	if err != nil {
		HandleErrorf("%T.ReadString(%s) failed: %v", r, s, err)
	}
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
			if a.IdentityList[j+1].less(a.IdentityList[j]) {
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
	a.PriorityMap = make(map[Identity]int)
	a.PriorityToIdentityMap = make(map[int]Identity)
}

func (a *AuthSet) AddHash(id interfaces.IHash, status int) int {
	return a.Add(Identity(id.Fixed()), status)
}

func (a *AuthSet) Add(id Identity, status int) int {
	index := len(a.IdentityList)
	a.IdentityMap[id] = index
	a.IdentityList = append(a.IdentityList, id)
	a.StatusArray = append(a.StatusArray, status)
	// a.Sort()

	a.PriorityMap = make(map[Identity]int)
	auds := a.GetAuds()
	for i, aud := range auds {
		a.PriorityMap[aud] = len(auds) - i - 1
		a.PriorityToIdentityMap[len(auds)-i-1] = aud
	}

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

func (a *AuthSet) FedIDtoIndex(id Identity) int {
	for i, f := range a.GetFeds() {
		if bytes.Compare(f[:], id[:]) == 0 {
			return i
		}
	}
	return -1
}

func (a *AuthSet) GetVolunteerPriority(vol Identity) int {
	return a.PriorityMap[vol]

	// Reverse logic
	//l := len(auds)
	//v := -1
	//for i, a := range auds {
	//	if a == vol {
	//		return l - i
	//	}
	//}
	//return v
}

type Identity [32]byte

func NewIdentityFromInt(u int) Identity {
	v := Uint32ToBytes(uint32(u))
	var i Identity
	copy(i[:4], v)
	return i
}

func (a Identity) less(b Identity) bool {
	return bytes.Compare(a[:], b[:]) < 0
}

func (i *Identity) String() string {
	return fmt.Sprintf("ID-%08x", *i)
}

func (i *Identity) MarshalJSON() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Identity.MarshalJSON err:%v", *pe)
		}
	}(&err)
	s := `"` + i.String() + `"`
	return []byte(s), nil
}
func (i *Identity) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = s[1 : len(s)-1] // trim off the surrounding "'s
	i.ReadString(s)
	return nil
}

func (i *Identity) ReadString(s string) {
	n, err := fmt.Sscanf(s, "ID-%x", i)
	if err != nil || n != 1 {
		HandleErrorf("Identity.ReadString(%v) failed: %d %v", s, n, err)
	}
}

// --------------------------------------------------------------------------------------------------------------------
type AuthorityStatus int

func (s *AuthorityStatus) String() string {
	switch *s {
	case 0:
		return "AUDIT"
	case 1:
		return "LEADER"
	default:
		return fmt.Sprintf("INVALID:%d", *s)
	}
}

func (a *AuthorityStatus) ReadString(s string) {
	switch s {
	case "AUDIT":
		*a = 0
	case "LEADER":
		*a = 1
	default:
		_, err := fmt.Sscanf(s, "INVALID:%d", a)
		if err != nil {
			HandleErrorf("AuthorityStatus.ReadString(\"%s\") failed: %v", s, err)

		}
	}
}
func (i *AuthorityStatus) MarshalJSON() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AuthorityStatus.MarshalJSON err:%v", *pe)
		}
	}(&err)
	s := `"` + i.String() + `"`
	return []byte(s), nil
}
func (i *AuthorityStatus) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = s[1 : len(s)-1] // trim off the surrounding "'s
	i.ReadString(s)
	return nil
}

// --------------------------------------------------------------------------------------------------------------------

type Hash [sha256.Size]byte

func (h *Hash) String() string {
	return fmt.Sprintf("-%s-", hex.EncodeToString(h[:]))
}

func (h *Hash) ReadString(s string) {
	t := hashRegEx.FindStringSubmatch(s) // drop the delimiters
	if t == nil || len(t) != 2 {
		HandleErrorf("Identity.ReadString(%v) failed", s)
		return
	}
	b, err := hex.DecodeString(t[1]) // decode the hash in hex
	n := len(b)
	if err != nil || n != sha256.Size {
		HandleErrorf("Identity.ReadString(%v) failed: %d %v", s, n, err)
	}
	copy(h[:], b[:])
}
