package election

import (
	"fmt"

	"math"

	"github.com/FactomProject/electiontesting/imessage"
	"github.com/FactomProject/electiontesting/messages"
	. "github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/volunteercontrol"
)

var _ = fmt.Println

type DepthLeaderLevel struct {
	Msg     *messages.LeaderLevelMessage
	VoteMsg *messages.VoteMessage
	Depth   int
}

func NewDepthLeaderLevel(ll *messages.LeaderLevelMessage, depth int) *DepthLeaderLevel {
	d := new(DepthLeaderLevel)
	d.Msg = ll
	d.Depth = depth

	return d
}

func (d *DepthLeaderLevel) Copy() *DepthLeaderLevel {
	b := new(DepthLeaderLevel)
	if d.Msg != nil {
		b.Msg = d.Msg.Copy()
	}
	if d.VoteMsg != nil {
		b.VoteMsg = d.VoteMsg.Copy()
	}
	b.Depth = d.Depth

	return b
}

type Election struct {
	// Level 0 volunteer votes map[vol]map[leader]msg
	VolunteerVotes map[Identity]map[Identity]*messages.VoteMessage

	// Indexed by volunteer
	VolunteerControls map[Identity]*volunteercontrol.VolunteerControl

	CurrentLevel int
	CurrentVote  messages.LeaderLevelMessage
	Self         Identity
	AuthSet
	ProcessListLocation

	MsgListIn  []*DepthLeaderLevel
	MsgListOut []*DepthLeaderLevel

	Display *Display
	// If I have committed to an answer and found enough to finish Election
	Committed bool

	// Some statistical info
	TotalMessages int

	// Each time I vote for the same vol in the next level
	CommitmentTally int
}

func NewElection(self Identity, authset AuthSet, loc ProcessListLocation) *Election {
	e := new(Election)
	e.VolunteerVotes = make(map[Identity]map[Identity]*messages.VoteMessage)
	e.VolunteerControls = make(map[Identity]*volunteercontrol.VolunteerControl)
	e.Self = self
	e.AuthSet = authset
	// Majority level starts at 1
	e.CurrentLevel = 1
	e.CurrentVote.Rank = -1
	e.CurrentVote.VolunteerPriority = -1
	e.MsgListIn = make([]*DepthLeaderLevel, 0)
	e.MsgListOut = make([]*DepthLeaderLevel, 0)

	// Used to determine volunteer priority
	e.ProcessListLocation = loc
	return e
}

func (e *Election) Execute(msg imessage.IMessage, depth int) (imessage.IMessage, bool) {
	//  ** Msg In Debug **
	if l, ok := msg.(*messages.LeaderLevelMessage); ok {
		e.MsgListIn = append(e.MsgListIn, NewDepthLeaderLevel(l, depth))
	}

	if v, ok := msg.(*messages.VoteMessage); ok {
		d := NewDepthLeaderLevel(nil, depth)
		d.VoteMsg = v
		e.MsgListIn = append(e.MsgListIn, d)
	}

	resp, c := e.execute(msg)

	//  ** Msg Out Debug **
	if l, ok := resp.(*messages.LeaderLevelMessage); ok {
		e.MsgListOut = append(e.MsgListOut, NewDepthLeaderLevel(l, depth))
	}

	if v, ok := resp.(*messages.VoteMessage); ok {
		d := NewDepthLeaderLevel(nil, depth)
		d.VoteMsg = v
		e.MsgListOut = append(e.MsgListOut, d)
	}

	return resp, c
}

// updateCurrentVote is called every time we send out a different vote
func (e *Election) updateCurrentVote(new *messages.LeaderLevelMessage) {
	//if e.CurrentVote.Rank == 0 {
	//	e.CommitmentTally = 0
	//	return
	//}
	if new.VolunteerPriority == e.CurrentVote.VolunteerPriority {
		if e.CurrentVote.Rank+1 == new.Rank {
			e.CommitmentTally++
			e.CurrentVote = *new
			if new.Rank == 0 {
				e.CommitmentTally = 0
			}
			return
		}
	}
	e.CommitmentTally = 1
	if new.Rank == 0 {
		e.CommitmentTally = 0
	}

	if e.CurrentVote.Rank >= 0 {
		prev := e.CurrentVote
		prev.Justification = []*messages.LeaderLevelMessage{}
		new.PreviousVote = &prev
	}
	e.CurrentVote = *new
	e.Display.Execute(new)
}

func (e *Election) execute(msg imessage.IMessage) (imessage.IMessage, bool) {
	e.TotalMessages++
	e.executeDisplay(msg)

	// We are done, never use anything else
	if e.Committed {
		return nil, false
	}

	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		return e.executeLeaderLevelMessage(msg.(*messages.LeaderLevelMessage))
	case *messages.VolunteerMessage:
		// Return a vote for this volunteer if we have not already
		vol := msg.(*messages.VolunteerMessage)
		if e.VolunteerVotes[vol.Signer] == nil {
			e.VolunteerVotes[vol.Signer] = make(map[Identity]*messages.VoteMessage)
		}

		// already voted
		if _, ok := e.VolunteerVotes[vol.Signer][e.Self]; ok {
			return nil, false
		}

		vote := messages.NewVoteMessage(*vol, e.Self)
		e.VolunteerVotes[vol.Signer][e.Self] = &vote

		msg, _ := e.execute(&vote)
		if ll, ok := msg.(*messages.LeaderLevelMessage); ok {
			for _, v := range e.VolunteerVotes[vol.Signer] {
				ll.VoteMessages = append(ll.VoteMessages, v)
			}
			return ll, true
		}

		e.Display.Execute(&vote)
		return &vote, true
	case *messages.VoteMessage:
		// Colecting these allows us to issue out 0.#
		vote := msg.(*messages.VoteMessage)

		change := e.addVote(vote)
		newll := e.getRank0Vote()
		if newll != nil {
			newll.Level = e.CurrentLevel
			e.CurrentLevel++
			if newll != nil {
				e.updateCurrentVote(newll)
			}
			e.addLeaderLevelMessage(newll)
			return newll, change
		}
		return nil, change
	}
	return nil, false
}

func (e *Election) addVote(vote *messages.VoteMessage) bool {
	if vote == nil {
		return false
	}
	vol := vote.Volunteer.Signer
	if e.VolunteerVotes[vol] == nil {
		e.VolunteerVotes[vol] = make(map[Identity]*messages.VoteMessage)
	}

	// Already seen this vote
	if _, ok := e.VolunteerVotes[vol][vote.Signer]; ok {
		return false
	}

	e.VolunteerVotes[vol][vote.Signer] = vote
	e.executeDisplay(vote)
	return true
}

func (e *Election) getRank0Vote() *messages.LeaderLevelMessage {
	// If we have a rank 1+, we will not issue a rank 0
	if e.CurrentVote.Rank > 0 {
		return nil
	}

	auds := e.GetAuds()
	for i := 1; i < len(auds); i++ {
		for j := 0; j < len(auds)-i; j++ {
			if auds[j] < auds[j+1] {
				auds[j], auds[j+1] = auds[j+1], auds[j]
			}
		}
	}

	// currentVote is -1. Updates when we get a rank 0
	for _, vol := range auds {
		if len(e.VolunteerVotes[vol]) >= e.Majority() {
			// We have a majority of level 0 votes and can issue a rank 0 LeaderLevel Message
			var volunteermsg messages.VolunteerMessage
			for _, vm := range e.VolunteerVotes[vol] {
				volunteermsg = vm.Volunteer
				break
			}

			ll := messages.NewLeaderLevelMessage(e.Self, 0, e.CurrentLevel, volunteermsg)
			ll.VolunteerPriority = e.getVolunteerPriority(vol)
			for _, v := range e.VolunteerVotes[vol] {
				ll.VoteMessages = append(ll.VoteMessages, v)
			}

			if e.CurrentVote.Less(&ll) {
				return &ll
			} else {
				return nil
			}
		}
	}
	return nil
}

func (e *Election) getVolunteerPriority(vol Identity) int {
	return e.GetVolunteerPriority(vol, e.ProcessListLocation)
}

func (e *Election) executeDisplay(msg imessage.IMessage) {
	if e.Display != nil {
		e.Display.Execute(msg)
	}
}

func (e *Election) executeLeaderLevelMessage(msg *messages.LeaderLevelMessage) (imessage.IMessage, bool) {
	// Add all messages to display and volunteer controllers. Then choose our best vote
	change := e.addLeaderLevelMessage(msg)
	for _, j := range msg.Justification {
		change = change || e.addLeaderLevelMessage(j)
	}

	// Best vote
	possibleVotes := []*messages.LeaderLevelMessage{}
	for _, vc := range e.VolunteerControls {
		vote := vc.CheckVoteCount()
		if vote != nil {
			vote.VolunteerPriority = e.getVolunteerPriority(vote.VolunteerMessage.Signer)
			possibleVotes = append(possibleVotes, vote)
		}
	}

	if len(possibleVotes) > 0 {
		BubbleSortLeaderLevelMsg(possibleVotes)
		// New Vote!
		vote := possibleVotes[0]
		if !e.CurrentVote.Less(vote) {
			return nil, change
		}

		if e.CurrentLevel > vote.Level {
			vote.Level = e.CurrentLevel
			e.CurrentLevel++
		} else {
			e.CurrentLevel = vote.Level + 1
		}

		e.updateCurrentVote(vote)

		vote = e.commitIfLast(vote)
		e.addLeaderLevelMessage(vote)
		if !e.Committed {
			resp, _ := e.execute(vote)
			if resp != nil {
				return resp, true
			}
		}

		return vote, true
	}

	// No best vote? Can we do a rank 0 with the new votes?

	return nil, change
}

func (e *Election) addLeaderLevelMessage(msg *messages.LeaderLevelMessage) bool {
	if msg.Level <= 0 {
		panic("level <= 0 should never happen")
	}

	// Volunteer Control keeps the highest votes for each volunteer for each leader
	if e.VolunteerControls[msg.VolunteerMessage.Signer] == nil {
		e.VolunteerControls[msg.VolunteerMessage.Signer] = volunteercontrol.NewVolunteerControl(e.Self, e.AuthSet)
	}

	change := false
	if msg.PreviousVote != nil {
		change = e.addLeaderLevelMessage(msg.PreviousVote)
	}

	e.Display.Execute(msg)
	change = change || e.VolunteerControls[msg.VolunteerMessage.Signer].AddVote(msg)

	voteChange := false
	// Votes exist, so we can add these to our vote map
	if len(msg.VoteMessages) > 0 {
		for _, v := range msg.VoteMessages {
			// Add vote to maps and display
			voteChange = voteChange || e.addVote(v)
			e.Display.Execute(v)
		}
	}

	return change
}

func (e *Election) commitIfLast(msg *messages.LeaderLevelMessage) *messages.LeaderLevelMessage {
	// If commit is true, then we are done. Return the EOM
	// commit := e.CommitmentIndicator.ShouldICommit(msg)
	if e.CommitmentTally > 2 { //commit {
		e.Committed = true
		msg.Committed = true
		e.Display.Execute(msg)
		return msg
	}
	return msg
}

// ***************
//
//  Display Stuff
//
// ***************

func (e *Election) PrintMessages() string {
	str := fmt.Sprintf("-- In -- (%p)\n", e.MsgListIn)
	for i, m := range e.MsgListIn {
		if m.Msg != nil {
			str += fmt.Sprintf("%d Depth:%d %s\n", i, m.Depth, e.Display.FormatMessage(m.Msg))
		} else if m.VoteMsg != nil {
			str += fmt.Sprintf("%d Depth:%d %s\n", i, m.Depth, e.Display.FormatMessage(m.VoteMsg))
		}
	}
	str += fmt.Sprintf("-- Out -- (%p)\n", e.MsgListOut)
	for i, m := range e.MsgListOut {
		if m.Msg != nil {
			str += fmt.Sprintf("%d Depth:%d %s\n", i, m.Depth, e.Display.FormatMessage(m.Msg))
		} else if m.VoteMsg != nil {
			str += fmt.Sprintf("%d Depth:%d %s\n", i, m.Depth, e.Display.FormatMessage(m.VoteMsg))
		}
	}
	return str
}

func (e *Election) VolunteerControlString() string {
	str := "VolunteerControls\n"
	if e.Display == nil {
		return "No display\n"
	}

	vcs := make([]string, 0)

	for i, v := range e.VolunteerControls {
		line := fmt.Sprintf("(%d) ", e.getVolunteerPriority(i))
		if e.VolunteerControls[i] == nil {
			line += "nil"
		} else {
			votes := ""
			sep := ""
			for _, vo := range v.Votes {
				votes += sep + e.Display.FormatMessage(vo)
				sep = ","
			}
			line += votes
		}

		vcs = append(vcs, line)
	}

	for _, l := range vcs {
		str += l + "\n"
	}

	return str
}

// AddDisplay takes a global tracker. Send nil if you don't care about
// a global state
func (e *Election) AddDisplay(global *Display) *Display {
	e.Display = NewDisplay(e, global)
	return e.Display
}

func (a *Election) Copy() *Election {
	b := NewElection(a.Self, a.AuthSet.Copy(), a.ProcessListLocation)
	b.TotalMessages = a.TotalMessages
	b.CommitmentTally = a.CommitmentTally

	for k, _ := range a.VolunteerVotes {
		b.VolunteerVotes[k] = make(map[Identity]*messages.VoteMessage)
		for k2, v2 := range a.VolunteerVotes[k] {
			b.VolunteerVotes[k][k2] = v2.Copy()
		}
	}

	for k, v := range a.VolunteerControls {
		b.VolunteerControls[k] = v.Copy()
	}

	b.CurrentLevel = a.CurrentLevel
	b.CurrentVote = *(a.CurrentVote.Copy())
	if a.Display == nil {
		b.Display = nil
	} else {
		b.Display = a.Display.Copy(b)
		b.Display.Global = a.Display.Global.Copy(b)
	}
	b.Committed = a.Committed
	b.MsgListIn = make([]*DepthLeaderLevel, len(a.MsgListIn))

	for i, v := range a.MsgListIn {
		b.MsgListIn[i] = v.Copy()
	}

	b.MsgListOut = make([]*DepthLeaderLevel, len(a.MsgListOut))
	for i, v := range a.MsgListOut {
		b.MsgListOut[i] = v.Copy()
	}

	return b
}

func (e *Election) StateString() []byte {
	return e.stateString(0)
}

func (e *Election) stateString(decrement int) []byte {
	vc := e.StateVCDataset()
	c := e.CurrentVote
	votes := e.StateVotes()

	// Combine into a string
	str := fmt.Sprintf("Current: (%d)%d.%d\n",
		c.Level-decrement, c.Rank-decrement, c.VolunteerPriority)

	vcstr := ""
	for vol, v := range vc {
		vcstr += fmt.Sprintf("%d->", vol)
		sep := ""
		for _, l := range v {
			vcstr += sep + fmt.Sprintf("(%d)%d.%d", l.Level-decrement, l.Rank-decrement, l.VolunteerPriority)
			sep = ","
		}
		vcstr += "\n"
	}

	votestr := ""
	sep := ""
	for vol, v := range votes {
		votestr += sep + fmt.Sprintf(" %d:%d", vol, v)
		sep = ","
	}
	votestr += "\n"

	return []byte(str + vcstr + votestr)
}

func (e *Election) NormalizedString() []byte {
	vc := e.StateVCDataset()
	lowest := math.MaxInt32
	for _, v := range vc {
		for _, v2 := range v {
			if v2.Rank < lowest {
				lowest = v2.Rank
			}
		}
	}
	decrement := (lowest - 1)
	return e.stateString(decrement)
}

func (e *Election) StateVotes() []int {
	var votearr []int
	votearr = make([]int, len(e.GetAuds()))
	for vol, votes := range e.VolunteerVotes {
		votearr[e.getVolunteerPriority(vol)] = len(votes)
	}
	return votearr
}

func (e *Election) StateVCDataset() [][]*messages.LeaderLevelMessage {
	// Loop through volunteers, and record only those that are above current vote

	var vcarray [][]*messages.LeaderLevelMessage
	vcarray = make([][]*messages.LeaderLevelMessage, len(e.GetAuds()))

	for vol, m := range e.VolunteerControls {
		var volarray []*messages.LeaderLevelMessage
		// vol is the volunteer
		for _, vote := range m.Votes {
			if e.CurrentVote.Level > 0 && e.CurrentVote.Rank > vote.Level {
				continue
			}
			volarray = append(volarray, vote)
		}

		vcarray[e.getVolunteerPriority(vol)] = bubbleSortLeaderLevelMsgByRank(volarray)
	}

	return vcarray
}

func BubbleSortLeaderLevelMsg(arr []*messages.LeaderLevelMessage) {
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if arr[j].Less(arr[j+1]) {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

func bubbleSortLeaderLevelMsgByRank(arr []*messages.LeaderLevelMessage) []*messages.LeaderLevelMessage {
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if arr[j].Rank > arr[j+1].Rank {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
	return arr
}
