package election

import (
	"fmt"
	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
)

// Display is a 2D array containing all the votes seen by all leaders
type Display struct {
	Identifier string

	Votes [][]string

	election *Election

	fedList []primitives.Identity

	primitives.AuthSet

	Global *Display
}

func NewDisplay(ele *Election, global *Display) *Display {
	d := new(Display)
	d.Votes = make([][]string, 0)
	d.election = ele
	d.AuthSet = ele.AuthSet
	d.fedList = d.GetFeds()
	d.Global = global
	d.Identifier = fmt.Sprintf("Leader %d", d.getColumn(ele.Self))

	return d
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
	for f, _ := range d.fedList {
		headerVal := fmt.Sprintf("L%d", f)
		str += center(headerVal)
	}
	return str
}

func (d *Display) String() string {
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

func center(str string) string {
	return fmt.Sprintf("%-4s", fmt.Sprintf("%4s", str))
}

func (d *Display) insertVote0Message(msg *messages.VoteMessage) {
	col := d.getColumn(msg.Signer)
	if col == -1 {
		// Error?
		return
	}
	row := 0

	// Make row will just ensure the row exists
	d.makeRow(row)

	vol := d.election.getVolunteerPriority(msg.Volunteer.Signer)
	d.Votes[row][col] = fmt.Sprintf("%d", vol)
}

func (d *Display) insertLeaderLevelMessage(msg *messages.LeaderLevelMessage) {
	col := d.getColumn(msg.Signer)
	if col == -1 {
		// Error?
		return
	}
	row := msg.Level

	// Make row will just ensure the row exists
	d.makeRow(row)

	d.Votes[row][col] = d.formatLeaderLevelMsg(msg)
}

func (d *Display) formatLeaderLevelMsg(msg *messages.LeaderLevelMessage) string {
	if msg.Committed {
		return "EOM"
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
		d.Votes = append(d.Votes, make([]string, len(d.fedList)))
		//}
	}
}

func (d *Display) getColumn(id primitives.Identity) int {
	for i, f := range d.fedList {
		if f == id {
			return i
		}
	}
	return -1
}
