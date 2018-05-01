package election

import (
	"fmt"
	"strings"

	"sync"

	"github.com/FactomProject/factomd/electionsCore/imessage"
	"github.com/FactomProject/factomd/electionsCore/messages"
	"github.com/FactomProject/factomd/electionsCore/primitives"
)

// Display is a 2D array containing all the votes seen by all leaders
type Display struct {
	Identifier string

	Votes [][]string

	FedList []primitives.Identity

	primitives.AuthSet

	Global *Display

	// Used to control display locking
	lock sync.RWMutex
}

func NewDisplay(ele *Election, global *Display) *Display {
	d := new(Display)
	d.Votes = make([][]string, 0)
	d.AuthSet = ele.AuthSet
	d.FedList = d.GetFeds()
	d.Global = global

	d.Identifier = fmt.Sprintf("Leader %d, ID: %x", d.getColumn(ele.Self), ele.Self[:8])
	if ele.Observer {
		d.Identifier = fmt.Sprintf("Observer ID: %x", ele.Self[:8])
	}

	return d
}

func (d *Display) ResetIdentifier(ele *Election) {
	d.Identifier = fmt.Sprintf("Leader %d, ID: %x", d.getColumn(ele.Self), ele.Self[:8])
	if ele.Observer {
		d.Identifier = fmt.Sprintf("Observer ID: %x", ele.Self[:8])
	}
}

func (a *Display) Copy(be *Election) *Display {
	b := NewDisplay(be, nil)
	b.Votes = make([][]string, len(a.Votes))
	b.Identifier = a.Identifier
	for i, v := range a.Votes {
		b.Votes[i] = make([]string, len(v))
		for i2, v2 := range a.Votes[i] {
			b.Votes[i][i2] = v2
		}
	}

	for i, f := range a.FedList {
		b.FedList[i] = f
	}

	return b
}

func (d *Display) Execute(msg imessage.IMessage) {
	if d.Global != nil {
		d.Global.Execute(msg)
	}
	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		ll := msg.(*messages.LeaderLevelMessage)
		d.insertLeaderLevelMessage(ll)
	case *messages.VoteMessage:
		v := msg.(*messages.VoteMessage)
		d.insertVote0Message(v)
	}
}

func (d *Display) pad(row int) string {
	return fmt.Sprintf(" %3s", fmt.Sprintf("%d:", row))
}

func (d *Display) stringHeader() string {
	str := fmt.Sprintf("(%s)\n", d.Identifier)
	// 3 spaces | L# centered 4 slots
	str += fmt.Sprintf(" %3s", "Lvl")
	for f := range d.FedList {
		headerVal := fmt.Sprintf("L%d", f)
		str += center(headerVal)
	}
	return str
}

func (d *Display) String() string {
	d.RLock()
	defer d.RUnlock()
	str := d.stringHeader() + "\n"

	for r := range d.Votes {
		var _ = r
		str += d.pad(r)
		for c := range d.Votes[r] {
			str += center(d.Votes[r][c])
		}
		str += "\n"
	}
	return str
}

//var w = 18

func center(str string) string {
	//return fmt.Sprintf(fmt.Sprintf("%%-%ds", w/2), fmt.Sprintf(fmt.Sprintf("%%%ds", w/2), str))
	return fmt.Sprintf("%-10s", fmt.Sprintf("%10s", str))
}

func (d *Display) insertVote0Message(msg *messages.VoteMessage) {
	col := d.getColumn(msg.Signer)
	if col == -1 {
		// Error?
		return
	}
	row := 0

	d.Lock()
	defer d.Unlock()
	// Make row will just ensure the row exists
	d.makeRow(row)

	vol := d.getVolunteerPriority(msg.Volunteer.Signer)
	vote0 := fmt.Sprintf("%d", vol)
	if strings.Contains(d.Votes[row][col], vote0) {
		return
	}

	d.Votes[row][col] += vote0
}

func (d *Display) getVolunteerPriority(id primitives.Identity) int {
	return d.AuthSet.GetVolunteerPriority(id)
}

func (d *Display) insertLeaderLevelMessage(msg *messages.LeaderLevelMessage) {
	if msg == nil {
		return
	}
	col := d.getColumn(msg.Signer)
	if col == -1 {
		// Error?
		return
	}
	row := msg.Level

	d.Lock()
	// Make row will just ensure the row exists
	d.makeRow(row)

	d.Votes[row][col] = d.FormatLeaderLevelMsgShort(msg)
	d.Unlock()
}

func (d *Display) FormatMessage(msg imessage.IMessage) string {
	if msg == nil {
		return "nil"
	}
	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		return d.FormatLeaderLevelMsg(msg.(*messages.LeaderLevelMessage))
	case *messages.VolunteerMessage:
		return d.FormatVolunteerMsg(msg.(*messages.VolunteerMessage))
	case *messages.VoteMessage:
		return d.FormatVoteMsg(msg.(*messages.VoteMessage))
	default:
		return "na"
	}
}

func (d *Display) FormatVoteMsg(msg *messages.VoteMessage) string {
	return fmt.Sprintf("L%d:V%d", d.AuthSet.FedIDtoIndex(msg.Signer), d.getVolunteerPriority(msg.Volunteer.Signer))
}

func (d *Display) FormatVolunteerMsg(msg *messages.VolunteerMessage) string {
	return fmt.Sprintf("V%d", d.getVolunteerPriority(msg.Signer))
}

func (d *Display) FormatLeaderLevelMsg(msg *messages.LeaderLevelMessage) string {
	return fmt.Sprintf("L%d:%d]%s", d.AuthSet.FedIDtoIndex(msg.Signer), msg.Level, d.FormatLeaderLevelMsgShort(msg))
}

func (d *Display) FormatLeaderLevelMsgShort(msg *messages.LeaderLevelMessage) string {
	if msg.Committed {
		if msg.EOMFrom == msg.Signer {
			return fmt.Sprintf("S-EOM%d", msg.VolunteerPriority)
		}
		return fmt.Sprintf("%d-EOM%d", d.FedIDtoIndex(msg.EOMFrom), msg.VolunteerPriority)
	}
	return fmt.Sprintf("%d.%d", msg.Rank, msg.VolunteerPriority)
}

// newRow will take a level and add a row for it.
func (d *Display) makeRow(level int) {
	if len(d.Votes) > level {
		// All good!
		return
	}
	for len(d.Votes) <= level {
		//if len(d.Votes) <= level {
		// Need to add rows to get to level
		d.Votes = append(d.Votes, make([]string, len(d.FedList)))
		//}
	}
}

func (d *Display) getColumn(id primitives.Identity) int {
	for i, f := range d.FedList {
		if f == id {
			return i
		}
	}
	return -1
}

// Lock functions (Put here so we can control the lock actions in 1 place)

func (d *Display) Lock() {
	d.lock.Lock()
}

func (d *Display) Unlock() {
	d.lock.Unlock()
}

func (d *Display) RLock() {
	d.lock.RLock()
}

func (d *Display) RUnlock() {
	d.lock.RUnlock()
}
