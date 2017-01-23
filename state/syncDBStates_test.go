package state_test

import (
	"testing"

	. "github.com/FactomProject/factomd/state"
	"fmt"
)

func TestAskStuff(t *testing.T) {
	state			 := new(State)
	histories  := new(AskHistories)
	fmt.Println("Highest Saved",state.HighestSaved)
	fmt.Println("Highest Known",state.HighestKnown)
	state.HighestSaved=100
	state.HighestKnown=200
	fmt.Println("Get(0)",histories.Get(0))
	histories.Trim(state.HighestKnown,state.HighestKnown)
	begin := FindBegin(state,histories,0)
	if begin > 0 {
		end := FindEnd(state, histories, uint32(begin))
		fmt.Println("begin/end",begin,end)
	}
}
