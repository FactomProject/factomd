package election

import (
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	"sort"
)

const NumberOfSequential int = 2

// DiamondShop is the place you go too looking for commitment. It will determine
// when you can commit an EOM and end an election.
// AKA: When it has a diamond ring that is affordable
type DiamondShop struct {
	VoteHistories map[primitives.Identity]*LeaderVoteHistory
	Commitment    map[primitives.Identity]int
	primitives.AuthSet
}

func NewDiamondShop(authset primitives.AuthSet) *DiamondShop {
	d := new(DiamondShop)
	d.VoteHistories = make(map[primitives.Identity]*LeaderVoteHistory)
	d.AuthSet = authset
	d.Commitment = make(map[primitives.Identity]int)
	for _, f := range d.GetFeds() {
		d.Commitment[f] = -1
		d.VoteHistories[f] = NewLeaderVoteHistory()
	}

	return d
}

// ShouldICommit will return a bool that tells you if you can commit to the election results.
// True --> Use the EOM, we are done
func (d *DiamondShop) ShouldICommit(msg *messages.LeaderLevelMessage) bool {
	c := d.VoteHistories[msg.Signer].Add(msg)
	d.Commitment[msg.Signer] = c

	tally := 0
	if c != -1 {
		for _, v := range d.Commitment {
			if v == c {
				tally++
			}
		}
	}

	return tally > d.Majority()
}

type LeaderVoteHistory struct {
	// [0].level > [1].level
	Votes []*messages.LeaderLevelMessage
}

func NewLeaderVoteHistory() *LeaderVoteHistory {
	h := new(LeaderVoteHistory)
	h.Votes = make([]*messages.LeaderLevelMessage, NumberOfSequential, NumberOfSequential)

	return h
}

// Add returns true if the leader is good to commit. Only returns true
// if we get 2 sequential levels with the same volunteer decision
func (h *LeaderVoteHistory) Add(l *messages.LeaderLevelMessage) int {
	place := -1

	// Check for nil. If a nil, we have an open spot
	for i, v := range h.Votes {
		// Found a spot
		if v == nil && place == -1 {
			place = i
		}
		// Need to check that we don't already have this vote
		if v.Level == l.Level {
			return -1
		}
	}

	if place != -1 {
		h.Votes[place] = l
		h.sort()
		return -1
	}

	// No nils, no duplicates, replace the lowest one
	h.Votes[len(h.Votes)-1] = l
	h.sort()
	return h.checkForComplete()
}

func (h *LeaderVoteHistory) checkForComplete() int {
	var vol primitives.Identity

	for i := 0; i < len(h.Votes)-1; i++ {
		// Check levels are sequential
		if h.Votes[i].Level != h.Votes[i+1].Level-1 {
			return -1
		}

		// Check auds are the same
		if h.Votes[i].VolunteerMessage.Signer != h.Votes[i+1].VolunteerMessage.Signer {
			return -1
		}
		vol = h.Votes[i].VolunteerMessage.Signer
	}

	return int(vol)
}

func (h *LeaderVoteHistory) sort() {
	sort.Sort(SortByLeaderLevel(h.Votes))
}

type SortByLeaderLevel []*messages.LeaderLevelMessage

func (s SortByLeaderLevel) Len() int {
	return len(s)
}
func (s SortByLeaderLevel) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less == better
func (s SortByLeaderLevel) Less(i, j int) bool {
	if s[i] == nil && s[j] == nil {
		return false
	}

	if s[j] == nil {
		return true
	}
	return s[i].Level > s[j].Level
}
