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


type Election struct {
	// Level 0 volunteer votes map[vol]map[leader]msg
	VolunteerVotes map[Identity]map[Identity]*messages.VoteMessage

	// Indexed by volunteer
	VolunteerControls map[Identity]*volunteercontrol.VolunteerControl

	CurrentLevel int
	CurrentVote  messages.LeaderLevelMessage
	Self         Identity
	AuthSet

	// Used for debugging and seeing history
	MsgListIn  []*DepthLeaderLevel
	MsgListOut []*DepthLeaderLevel

	Display *Display
	// If I have committed to an answer and found enough to finish Election
	Committed bool

	// Some statistical info
	TotalMessages int

	// Each time I vote for the same vol in the next level
	CommitmentTally int
	// An observer never participates in an election, but can watch (audit or follower)
	Observer bool
}

func NewElection(self Identity, authset AuthSet) *Election {
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

	return e
}

func (e *Election) SetObserver(o bool) {
	e.Observer = o
	e.Display.ResetIdentifier(e)
}

func (e *Election) Execute(msg imessage.IMessage, depth int) (imessage.IMessage, bool) {
	if e.Observer {
		e.executeObserver(msg)
		return nil, false
	}

	//  ** Msg In Debug **
	if l, ok := msg.(*messages.LeaderLevelMessage); ok {
		e.MsgListIn = append(e.MsgListIn, NewDepthLeaderLevel(l, depth))
	}

	if v, ok := msg.(*messages.VoteMessage); ok {
		d := NewDepthLeaderLevel(nil, depth)
		d.VoteMsg = v
		e.MsgListIn = append(e.MsgListIn, d)
	}

	resp, c := e.execute(msg) // <-- The only non-debug code in this function

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

func (e *Election) executeObserver(msg imessage.IMessage) {
	e.TotalMessages++
	e.executeDisplay(msg)
	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		e.addLeaderLevelMessage(msg.(*messages.LeaderLevelMessage))
	case *messages.VolunteerMessage:
	case *messages.VoteMessage:
		e.addVote(msg.(*messages.VoteMessage))
	}
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
		if e.VolunteerVotes[vol.Signer] == nil { // Making sure no nil's in map when we are using it
			e.VolunteerVotes[vol.Signer] = make(map[Identity]*messages.VoteMessage)
		}

		// already voted, no need to vote again
		if _, ok := e.VolunteerVotes[vol.Signer][e.Self]; ok {
			return nil, false
		}

		// Cast our vote, this is at level 0, so we can cast as many as there are volunteers
		vote := messages.NewVoteMessage(*vol, e.Self)
		e.VolunteerVotes[vol.Signer][e.Self] = &vote

		e.executeDisplay(&vote) // Update display

		// We might be able to get a better vote from this, execute and look for response
		resp, _ := e.execute(&vote)
		if resp != nil {
			return resp, true
		}

		return &vote, true
	case *messages.VoteMessage:
		// Collecting these allows us to issue out 0.#
		vote := msg.(*messages.VoteMessage)

		change := e.addVote(vote)
		newll := e.getRank0Vote()
		if newll != nil { // got a rank0 vote. Check if we can get anything better first
				e.updateCurrentVote(newll)
			resp, _ := e.execute(newll)
			if resp != nil {
				return resp, true
			}
			// Rank0 is best
			return newll, true
		}
		return nil, change
	}
	return nil, false
}

// updateCurrentVote is called every time we send out a different vote
func (e *Election) updateCurrentVote(new *messages.LeaderLevelMessage) {
	new.VolunteerPriority = e.GetVolunteerPriority(new.VolunteerMessage.Signer)
	// Add to display, add the previous vote
	e.executeDisplay(new)
	if e.CurrentVote.Rank >= 0 {
		prev := e.CurrentVote.Copy()
		prev.Justification = []messages.LeaderLevelMessage{}
		prev.PreviousVote = nil
		new.PreviousVote = prev
	}

	/**** Commitment checking ****/
	// Check if this is sequential
	if new.VolunteerPriority == e.CurrentVote.VolunteerPriority &&
		(e.CurrentVote.Rank+1 == new.Rank && e.CurrentVote.Level+1 == new.Level) {
		e.CommitmentTally++
	} else {
		// Resetting the tally
		e.CommitmentTally = 1
	}

	if new.Rank == 0 {
		// Rank 0 doesn't count towards the tally
		e.CommitmentTally = 0
	}
	e.CurrentVote = *new

}

// addVote adds the proposal vote without acting upon it.
//		Returns true if it is new
func (e *Election) addVote(vote *messages.VoteMessage) bool {
	if vote == nil {
		// Can never be too sure
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

// getRank0Vote will return a rank 0 vote if one can be cast and is better than the current
func (e *Election) getRank0Vote() *messages.LeaderLevelMessage {
	// If we have a rank 1+, we will not issue a rank 0
	if e.CurrentVote.Rank > 0 {
		return nil
	}

	// Assume sorted
	auds := e.GetAuds()

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
				e.CurrentLevel++
				return &ll
			} else {
				return nil
			}
		}
	}
	return nil
}

func (e *Election) getVolunteerPriority(vol Identity) int {
	return e.GetVolunteerPriority(vol)
}

// executeDisplay will add to the display if the display is not nil
func (e *Election) executeDisplay(msg imessage.IMessage) {
	if e.Display != nil {
		e.Display.Execute(msg)
	}
}

// executeLeaderLevelMessage will add all the messages contained, then act upon the messge and come up
// with a response
func (e *Election) executeLeaderLevelMessage(msg *messages.LeaderLevelMessage) (imessage.IMessage, bool) {
	// Add all messages to display and volunteer controllers. Then choose our best vote
	change := e.addLeaderLevelMessage(msg)
	for _, j := range msg.Justification {
		change = e.addLeaderLevelMessage(&j) || change
	}

	// Special case. If the msg we get is an EOM, we can immediately issue an EOM and say we are done
	if msg.Committed {
		ll := messages.NewLeaderLevelMessage(e.Self, msg.Rank, e.CurrentLevel, msg.VolunteerMessage)
		skipped := &ll

		skipped.Level = e.CurrentLevel
		e.CurrentLevel++
		e.updateCurrentVote(skipped)
		e.CommitmentTally = 4 // We can commit
		skipped = e.commitIfLast(skipped)
		skipped.EOMFrom = msg.Signer
		// We need to set the EOMFrom and update display
		e.executeDisplay(skipped)

		return skipped, true
	}

	// All possible votes that we can choose from
	possibleVotes := []*messages.LeaderLevelMessage{}
	for _, vc := range e.VolunteerControls {
		vote := vc.CheckVoteCount()
		if vote != nil {
			vote.VolunteerPriority = e.getVolunteerPriority(vote.VolunteerMessage.Signer)
			possibleVotes = append(possibleVotes, vote)
		}
	}

	// Sort the votes such that the highest ranked vote is first
	if len(possibleVotes) > 0 {
		// TODO: Use select/scan
		BubbleSortLeaderLevelMsg(possibleVotes)
		// New Vote!
		vote := possibleVotes[0]
		// If our current vote is equal or greater, then don't cast it
		if !e.CurrentVote.Less(vote) {
			return nil, change
		}

		if vote.Rank > vote.Level {
			panic(fmt.Sprintf("Vote rank should never be greater than level: %v", vote))
		}

		// Set the level on the vote
		if e.CurrentLevel >= vote.Level {
			vote.Level = e.CurrentLevel
			e.CurrentLevel++
		} else {
			e.CurrentLevel = vote.Level
				e.CurrentLevel = vote.Level + 1
			}

		// Update our last vote
		e.updateCurrentVote(vote)

		// If it's the last vote, we are done
		vote = e.commitIfLast(vote)
		e.addLeaderLevelMessage(vote)
		if !e.Committed {
			// Not the last vote means we might be able to make a better
			// vote off of our current. Execute it and see if there is a response, meaning something better.
			resp, _ := e.execute(vote)
			if resp != nil {
				return resp, true
			}
		}

		return vote, true
	}

	// No best vote? Can we do a rank 0 with the new votes?

	rank0 := e.getRank0Vote()
	if rank0 != nil {
		e.updateCurrentVote(rank0)
		e.addLeaderLevelMessage(rank0)
		better, _ := e.execute(rank0)
		if better != nil {
			return better, true
		}
		return rank0, true
	}
	return nil, change
}

// addLeaderLevelMessage will add the msg to our lists, but not act upon it.
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

	e.executeDisplay(msg)
	change = e.VolunteerControls[msg.VolunteerMessage.Signer].AddVote(*msg) || change

	voteChange := false
	// Votes exist, so we can add these to our vote map
		for _, v := range msg.VoteMessages {
			// Add vote to maps and display
			voteChange = e.addVote(v) || voteChange
	}

	return change || voteChange
}

// commitIfLast will mark the messages as committed if the correct criteria are met
func (e *Election) commitIfLast(msg *messages.LeaderLevelMessage) *messages.LeaderLevelMessage {
	// If commit is true, then we are done. Return the EOM
	// commit := e.CommitmentIndicator.ShouldICommit(msg)
	if e.CommitmentTally > 3 { //commit {
		e.Committed = true
		msg.Committed = true
		msg.EOMFrom = e.Self
		e.executeDisplay(msg)
	}
	return msg
}

// ***************
//
//  Display & Normalization Stuff
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
	str := ""
	str += "Leaders\n"
	for i, l := range e.GetFeds() {
		str += fmt.Sprintf("%d: [%x]\n", i, l[:8])
	}

	str += "\n"
	str += "VolunteerControls\n"
	if e.Display == nil {
		return "No display\n"
	}

	vcs := make([]string, 0)

	for i, v := range e.VolunteerControls {
		line := fmt.Sprintf("[%x](%d) ", i[:8], e.getVolunteerPriority(i))
		if e.VolunteerControls[i] == nil {
			line += "nil"
		} else {
			votes := ""
			sep := ""
			arr := make([]messages.LeaderLevelMessage, len(v.Votes))
			i := 0
			for _, vo := range v.Votes {
				arr[i] = vo
				i++
			}
			arr = bubbleSortLeaderLevelMsgWithLevel(arr)

			for _, vo := range arr {
				votes += sep + e.Display.FormatMessage(&vo)
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
	b := NewElection(a.Self, a.AuthSet.Copy())
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

func (e *Election) StateVCDataset() [][]messages.LeaderLevelMessage {
	// Loop through volunteers, and record only those that are above current vote

	var vcarray [][]messages.LeaderLevelMessage
	vcarray = make([][]messages.LeaderLevelMessage, len(e.GetAuds()))

	for vol, m := range e.VolunteerControls {
		var volarray []messages.LeaderLevelMessage
		// vol is the volunteer
		for _, vote := range m.Votes {
			if e.CurrentVote.Level > 0 && e.CurrentVote.Rank > vote.Level {
				continue
			}
			volarray = append(volarray, vote)
		}

		vcarray[e.getVolunteerPriority(vol)] = bubbleSortLeaderLevelMsgWithLevel(volarray)
	}

	return vcarray
}

/****************
 ****************/
func BubbleSortLeaderLevelMsg(arr []*messages.LeaderLevelMessage) {
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if arr[j].Less(arr[j+1]) {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

func levelMessageLessWithLevel(a messages.LeaderLevelMessage, b messages.LeaderLevelMessage) bool {
	if a.Level == b.Level {
		return a.Less(&b)
	}
	return a.Level < b.Level
}

func bubbleSortLeaderLevelMsgWithLevel(arr []messages.LeaderLevelMessage) []messages.LeaderLevelMessage {
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if levelMessageLessWithLevel(arr[j], (arr[j+1])) {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
	return arr
}

func bubbleSortLeaderLevelMsgByRank(arr []messages.LeaderLevelMessage) []messages.LeaderLevelMessage {
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if arr[j].Rank > arr[j+1].Rank {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
	return arr
}

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
