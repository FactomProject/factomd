package controller_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/electiontesting/controller"
	. "github.com/FactomProject/electiontesting/errorhandling"
)

func TestSimpleController(t *testing.T) {
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2}
	con.RouteVolunteerMessage(1, all)

}

func TestElectionDisplay(t *testing.T) {
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2}
	con.RouteVolunteerMessage(1, all)
	fmt.Println(con.GobalDisplay.String())

	ExpMsg(con.RouteLeaderSetLevelMessage(all, 0, all))
	fmt.Println(con.GobalDisplay.String())

	fmt.Println(con.Elections[0].Display.String())
	//
	//ExpMsg(con.RouteLeaderLevelMessage(1, 0, all))
	//ExpMsg(con.RouteLeaderLevelMessage(2, 0, all))
	//
	//ExpMsg(con.RouteLeaderLevelMessage(0, 2, all))

	//fmt.Println(con.GobalDisplay.String())
}
