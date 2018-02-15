package election

import (
	"strconv"
	"strings"
)

// Detecting Looping

func (d *Display) DetectLoop(leader int) bool {
	return d.DetectVerticalLoop(leader)
}

// detectVerticalLoop detects if leader # is looping vertically
func (d *Display) DetectVerticalLoop(leader int) bool {
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

func (d *Display) getLeaderVotes(leader int) (myvotes []string) {
	for i := 0; i < len(d.Votes); i++ {
		myvotes = append(myvotes, d.Votes[i][leader])
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
