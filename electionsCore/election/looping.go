package election

import (
	"strconv"
	"strings"

	"fmt"

	"github.com/PaulSnow/factom2d/electionsCore/primitives"
)

func (d *Display) DetectIllegalVotes() (loops int) {
	for _, f := range d.GetFeds() {
		if d.DetectIllegalVote(f) {
			loops++
		}
	}
	return
}

// Detecting Looping
func (d *Display) DetectLoops() (loops int) {
	for _, f := range d.GetFeds() {
		if d.DetectLoopForLeader(f) {
			loops++
		}
	}
	return
}

// DetectIllegalVote will detect if the vote sequence is valid
//		true ==> Illegal vote
func (d *Display) DetectIllegalVote(leader primitives.Identity) bool {
	myVotes := d.getLeaderVotes(leader)
	rnk, vol := 0, 0
	for i, v := range myVotes {
		nxtrnk, nxtvol := parseVote(v)
		if nxtrnk == -1 {
			continue
		}
		if i > 0 {
			if nxtrnk < rnk {
				fmt.Println("r", nxtrnk, rnk)
				return true
			}
			if nxtrnk == rnk {
				if nxtvol < vol {
					fmt.Println(nxtvol, vol)
					return true
				}
			}
		}
		vol = nxtvol
		rnk = nxtrnk
	}

	return false
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
	rnk := -1
	tally := 0
	for i, v := range last3 {
		nxtrnk, nxtvol := parseVote(v)
		if i > 0 {
			if nxtvol != vol && nxtrnk == rnk+1 {
				tally++
			}
		}
		vol = nxtvol
		rnk = nxtrnk
	}

	return tally >= 2
}

func (d *Display) getLeaderVotes(leader primitives.Identity) (myvotes []string) {
	for i := 0; i < len(d.Votes); i++ {
		if d.FedIDtoIndex(leader) == -1 {
			panic("Leader was -1, but it should never be")
		}
		myvotes = append(myvotes, d.Votes[i][d.FedIDtoIndex(leader)])
	}
	// Now trim the end
	for i := len(myvotes) - 1; i >= 0; i-- {
		if myvotes[i] == "" {
			myvotes = myvotes[:i]
		} else {
			break
		}
	}

	return
}

func parseVote(vote string) (rank int, vol int) {
	strs := strings.Split(vote, ".")
	if len(strs) != 2 {
		return -1, -1
	}

	rank, _ = strconv.Atoi(strs[0])
	vol, _ = strconv.Atoi(strs[1])
	return rank, vol
}
