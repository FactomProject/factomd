package controller_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/electionsCore/controller"
	. "github.com/FactomProject/factomd/electionsCore/errorhandling"
)

var _ = fmt.Println

/*
Scenario found
(Global)
 Lvl    L0    L1    L2
  0:   012   012   012
  1:         0.0   0.0
  2:
  3:
  4:         1.0

(Leader 0, ID: 1)
 Lvl    L0    L1    L2
  0:   012     0
  1:   0.0

(Leader 1, ID: 2)
 Lvl    L0    L1    L2
  0:     0   012   120
  1:         0.0   0.0
  2:         0.1
  3:         0.2
  4:         1.0

(Leader 2, ID: 3)
 Lvl    L0    L1    L2
  0:     0    12   012
  1:               0.0
  2:               0.1
  3:               0.2

Leader 0
-- In -- (0x629798)
-- Out -- (0xc420096d60)
0 Depth:13 L0:1]0.0

Leader 1
-- In -- (0xc420096e00)
0 Depth:12 L2:1]0.0
-- Out -- (0xc420416e10)
0 Depth:10 L1:1]0.0
1 Depth:13 L1:2]0.1
2 Depth:13 L1:3]0.2
3 Depth:12 L1:4]1.0

Leader 2
-- In -- (0x629798)
-- Out -- (0xc4203ecea0)
0 Depth:11 L2:1]0.0
1 Depth:14 L2:2]0.1
2 Depth:14 L2:3]0.2
*/

func TestElectionScenario01(t *testing.T) {
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2}
	con.SendOutputsToRouter(true)
	con.RouteVolunteerMessage(0, all)
	con.RouteVolunteerMessage(1, all)
	con.RouteVolunteerMessage(2, all)

	con.RouteLeaderSetVoteMessage(all, 0, []int{0, 2})
	con.RouteLeaderSetVoteMessage(all, 0, []int{1})
	con.RouteLeaderSetVoteMessage(all, 1, []int{1})
	con.RouteLeaderSetVoteMessage(all, 2, []int{1})

	con.RouteLeaderSetLevelMessage([]int{2}, 1, []int{1})

	t.Log(con.ElectionStatus(1))

	// con.RouteLeaderSetVoteMessage(all, 1, all)
	// con.RouteLeaderSetVoteMessage(all, 1, all)
	// con.RouteLeaderSetVoteMessage(all, 1, all)

	// con.RouteVolunteerMessage(1, all)
	// con.RouteLeaderSetVoteMessage(all, 1, all)
	// con.RouteLeaderSetLevelMessage(all, 1, all)
	// con.RouteLeaderSetLevelMessage(all, 2, all)

	// loop := con.GlobalDisplay.DetectVerticalLoop(con.Elections[0].Self)
	// if loop {
	// 	t.Errorf("Detected a vert loop when there is not one")
	// }

	// runToComplete(con, t)
}
