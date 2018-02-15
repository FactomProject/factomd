package election

import (
	"fmt"
	"github.com/FactomProject/electiontesting/primitives"
	"strconv"
	"strings"
)

// Detecting Looping

func (d *Display) DetectLoops() (loops int) {
	for _, f := range d.GetFeds() {
		if d.DetectLoopForLeader(f) {
			loops++
		}
	}
	return
}

func (d *Display) DetectLoopForLeader(leader primitives.Identity) bool {
	return d.DetectVerticalLoop(leader)
}

// detectVerticalLoop detects if leader # is looping vertically
func (d *Display) DetectVerticalLoop(leader primitives.Identity) bool {
	myVotes := d.getLeaderVotes(leader)
	if len(myVotes) < 5 {
		return false
	}
	last3 := myVotes[len(myVotes)-3:]
	vol := -1
	tally := 0
	for i, v := range last3 {
		_, nxtvol := parseVote(v)
		if i > 0 {
			if nxtvol == vol+1 {
				tally++
			}
		}
		vol = nxtvol
	}

	return tally == 2
}

func (d *Display) getLeaderVotes(leader primitives.Identity) (myvotes []string) {
	for i := 0; i < len(d.Votes); i++ {
		if d.FedIDtoIndex(leader) >= len(d.Votes[i]) {
			fmt.Print()
		}
		myvotes = append(myvotes, d.Votes[i][d.FedIDtoIndex(leader)])
	}
	return
}

func parseVote(vote string) (int, int) {
	strs := strings.Split(vote, ".")
	if len(strs) != 2 {
		return -1, -1
	}

	rank, _ := strconv.Atoi(strs[0])
	vol, _ := strconv.Atoi(strs[1])
	return rank, vol
}
