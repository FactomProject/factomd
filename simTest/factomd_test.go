package simtest

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/util/atomic"
	"github.com/FactomProject/factomd/wsapi"
)

func TestOne(t *testing.T) {
	if RanSimTest {
		return
	}
	state.MMR_enable = false // No MMR for you!

	RanSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LF", map[string]string{"--fastsaverate": "5"}, 12, 0, 0, t)

	RunCmd("0")   // select 2
	RunCmd("R30") // Feed load
	WaitBlocks(state0, 5)
	RunCmd("R0") // Stop load
	WaitBlocks(state0, 2)
	ShutDownEverything(t)
} // testOne(){...}

func TestDualElections(t *testing.T) {
	if RanSimTest {
		return
	}
	state.MMR_enable = false // No MMR for you!

	RanSimTest = true

	// 							  01234567
	state0 := SetupSim("LALLLALFFLLFFFF", map[string]string{"--debuglog": ".", "--blktime": "20"}, 12, 0, 0, t)

	WaitMinutes(state0, 8)
	RunCmd("2")            // select 2
	RunCmd("x")            // off the net
	RunCmd("6")            // select 6
	RunCmd("x")            // off the net
	WaitMinutes(state0, 2) // wait for elections
	RunCmd("2")            // select 2
	RunCmd("x")            // on the net
	RunCmd("6")            // select 6
	RunCmd("x")            // on the net
	WaitBlocks(state0, 2)  // wait till nodes should have updated by dbstate
	WaitForAllNodes(state0)
	ShutDownEverything(t)
} // TestDualElections(){...}

func TestLoad(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLLLFFFF", map[string]string{"--debuglog": ".", "--blktime": "30"}, 15, 0, 0, t)

	RunCmd("2")    // select 2
	RunCmd("w")    // feed load into follower
	RunCmd("F200") // delay messages
	RunCmd("R25")  // Feed load
	WaitBlocks(state0, 3)
	RunCmd("R0") // Stop load
	for state0.Hold.GetSize() > 10 || len(state0.Holding) > 10 {
		WaitBlocks(state0, 1)
	}
	ShutDownEverything(t)
} // testLoad(){...}

// Test replicates a savestate restore bug when run twice. First run must complete 10 blocks.
func TestErr(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true
	state0 := SetupSim("LF", map[string]string{"--debuglog": ".", "--db": "LDB", "--controlpanelsetting": "readwrite",
		"--network": "LOCAL", "--fastsaverate": "4", "--checkheads": "false", "--net": "alot",
		"--blktime": "15", "--faulttimeout": "120000", "--enablenet": "false", "--startdelay": "1"},
		150, 0, 0, t)

	RunCmd("2")    // select 2
	RunCmd("w")    // feed load into follower
	RunCmd("F200") // delay messages
	RunCmd("R0")   // Feed load
	WaitBlocks(state0, 5)
	RunCmd("R0") // Stop load
	WaitBlocks(state0, 5)
	// should check holding and queues cleared out
	ShutDownEverything(t)
} //TestErr(){...}
func TestCatchup(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LF", map[string]string{}, 15, 0, 0, t)
	state1 := GetFnodes()[1].State

	RunCmd("1") // select 1
	RunCmd("x")
	RunCmd("R5") // Feed load
	WaitBlocks(state0, 5)
	RunCmd("R0")          // Stop load
	RunCmd("x")           // back online
	WaitBlocks(state0, 3) // give him a few blocks to catch back up
	//todo: check that the node01 caught up and finished 2nd pass sync
	dbht0 := state0.GetLLeaderHeight()
	dbht1 := state1.GetLLeaderHeight()

	if dbht0 != dbht1 {
		t.Fatalf("Node 0 was at dbheight %d which didn't match Node 1 at dbheight %d", dbht0, dbht1)
	}

	ShutDownEverything(t)
} //TestCatchup(){...}

// Test that we don't put invalid TX into a block.  This is done by creating transactions that are just outside
// the time for the block, and we let the block catch up.  The code should validate against the block time of the
// block to ensure that we don't record an invalid transaction in the block relative to the block time.
func TestTXTimestampsAndBlocks(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	go RunCmd("Re") // Turn on tight allocation of EC as soon as the simulator is up and running
	state0 := SetupSim("LLLAAAFFF", map[string]string{}, 24, 0, 0, t)
	StatusEveryMinute(state0)

	RunCmd("7") // select node 7
	RunCmd("x") // take out 7 from the network
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)
	RunCmd("Rt60") // Offset FCT transaction into the future by 60 minutes
	RunCmd("R.5")  // turn down the load
	WaitBlocks(state0, 2)
	RunCmd("x")
	RunCmd("R0") // turn off the load
}
func TestLoad2(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// use tree node setup so messages get reordered
	go RunCmd("Re") // Turn on tight allocation of EC as soon as the simulator is up and running
	state0 := SetupSim("LLLAF", map[string]string{"--blktime": "20", "--net": "tree"}, 24, 0, 0, t)
	StatusEveryMinute(state0)

	RunCmd("4") // select node 4
	RunCmd("x") // take out 4 from the network
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)

	RunCmd("R20") // Feed load
	WaitBlocks(state0, 3)
	RunCmd("Rt60")
	RunCmd("T20")
	RunCmd("R.5")
	WaitBlocks(state0, 2)
	RunCmd("x")
	RunCmd("R0")

	WaitBlocks(state0, 3)
	WaitMinutes(state0, 3)

	ht1 := GetFnodes()[1].State.GetLLeaderHeight()
	ht4 := GetFnodes()[4].State.GetLLeaderHeight()

	if ht1 != ht4 {
		t.Fatalf("Node 1 was at dbheight %d which didn't match Node 4 at dbheight %d", ht1, ht4)
	}
	ShutDownEverything(t)
} //TestLoad2(){...}
// The intention of this test is to detect the EC overspend/duplicate commits (FD-566) bug.
// the bug happened when the FCT transaction and the commits arrived in different orders on followers vs the leader.
// Using a message delay, drop and tree network makes this likely
//
func TestLoadScrambled(t *testing.T) {
	if RanSimTest {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestLoadScrambled: %v", r)
		}
	}()

	RanSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLFFFFFF", map[string]string{"--net": "tree"}, 32, 0, 0, t)
	//TODO: Why does this run longer than expected?

	RunCmd("2")     // select 2
	RunCmd("F1000") // set the message delay
	RunCmd("S10")   // delete 1% of the messages
	RunCmd("r")     // rotate the load around the network
	RunCmd("R3")    // Feed load
	WaitBlocks(state0, 10)
	RunCmd("R0") // Stop load
	WaitBlocks(state0, 1)

	ShutDownEverything(t)
} //TestLoadScrambled(){...}

func TestMinute9Election(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLAL", map[string]string{"--net": "line"}, 10, 1, 1, t)
	state3 := GetFnodes()[3].State

	WaitForMinute(state3, 9)
	RunCmd("3")
	RunCmd("x")
	WaitMinutes(state0, 1)
	RunCmd("x")
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)

	WaitForAllNodes(state0)
	ShutDownEverything(t)
} //TestMinute9Election(){...}

func TestMakeALeader(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LF", map[string]string{}, 5, 0, 0, t)
	RunCmd("g1")
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)

	RunCmd("1") // select node 1
	RunCmd("l") // make him a leader
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	// Adjust expectations
	Leaders++
	Followers--
	ShutDownEverything(t)
}

//func TestActivationHeightElection(t *testing.T) {
//	if RanSimTest {
//		return
//	}
//
//	RanSimTest = true
//
//	state0 := SetupSim("LLLLLAAF", map[string]string{}, 16, 2, 2, t)
//
//	// Kill the last two leader to cause a double election
//	RunCmd("3")
//	RunCmd("x")
//	RunCmd("4")
//	RunCmd("x")
//
//	WaitMinutes(state0, 2) // make sure they get faulted
//
//	// bring them back
//	RunCmd("3")
//	RunCmd("x")
//	RunCmd("4")
//	RunCmd("x")
//	WaitBlocks(state0, 2)
//	WaitMinutes(state0, 1)
//	WaitForAllNodes(state0)
//	CheckAuthoritySet(t)
//
//	if GetFnodes()[3].State.Leader {
//		t.Fatalf("Node 3 should not be a leader")
//	}
//	if GetFnodes()[4].State.Leader {
//		t.Fatalf("Node 4 should not be a leader")
//	}
//	if !GetFnodes()[5].State.Leader {
//		t.Fatalf("Node 5 should be a leader")
//	}
//	if !GetFnodes()[6].State.Leader {
//		t.Fatalf("Node 6 should be a leader")
//	}
//
//	CheckAuthoritySet(t)
//
//	if state0.IsActive(activations.ELECTION_NO_SORT) {
//		t.Fatalf("ELECTION_NO_SORT active too early")
//	}
//
//	for !state0.IsActive(activations.ELECTION_NO_SORT) {
//		WaitBlocks(state0, 1)
//	}
//
//	WaitForMinute(state0, 2) // Don't Fault at the end of a block
//
//	// Cause a new double elections by killing the new leaders
//	RunCmd("5")
//	RunCmd("x")
//	RunCmd("6")
//	RunCmd("x")
//	WaitMinutes(state0, 2) // make sure they get faulted
//	// bring them back
//	RunCmd("5")
//	RunCmd("x")
//	RunCmd("6")
//	RunCmd("x")
//	WaitBlocks(state0, 3)
//	WaitMinutes(state0, 1)
//	WaitForAllNodes(state0)
//	CheckAuthoritySet(t)
//
//	if GetFnodes()[5].State.Leader {
//		t.Fatalf("Node 5 should not be a leader")
//	}
//	if GetFnodes()[6].State.Leader {
//		t.Fatalf("Node 6 should not be a leader")
//	}
//	if !GetFnodes()[3].State.Leader {
//		t.Fatalf("Node 3 should be a leader")
//	}
//	if !GetFnodes()[4].State.Leader {
//		t.Fatalf("Node 4 should be a leader")
//	}
//
//	ShutDownEverything(t)
//}

func TestAnElection(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLAAF", map[string]string{"--blktime": "15"}, 9, 1, 1, t)

	StatusEveryMinute(state0)
	WaitMinutes(state0, 2)

	RunCmd("2")
	RunCmd("w") // point the control panel at 2

	// remove the last leader
	RunCmd("2")
	RunCmd("x")
	// wait for the election
	WaitMinutes(state0, 2)
	//bring him back
	RunCmd("x")

	// wait for him to update via dbstate and become an audit
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)

	// PrintOneStatus(0, 0)
	if GetFnodes()[2].State.Leader {
		t.Fatalf("Node 2 should not be a leader")
	}
	if !GetFnodes()[3].State.Leader && !GetFnodes()[4].State.Leader {
		t.Fatalf("Node 3 or 4  should be a leader")
	}

	WaitForAllNodes(state0)
	ShutDownEverything(t)

}

func TestDBsigEOMElection(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 9, 4, 4, t)

	// get status from FNode02 because he is not involved in the elections
	state2 := GetFnodes()[2].State
	StatusEveryMinute(state2)

	var wait sync.WaitGroup
	wait.Add(2)

	// wait till after EOM 9 but before DBSIG
	stop0 := func() {
		s := GetFnodes()[0].State
		// wait till minute flips
		for s.CurrentMinute != 0 {
			runtime.Gosched()
		}
		s.SetNetStateOff(true)
		wait.Done()
		fmt.Println("Stopped FNode0")
	}

	// wait for after DBSIG is sent but before EOM0
	stop1 := func() {
		s := GetFnodes()[1].State
		for s.CurrentMinute != 0 {
			runtime.Gosched()
		}
		pl := s.ProcessLists.Get(s.LLeaderHeight)
		vm := pl.VMs[s.LeaderVMIndex]
		for s.CurrentMinute == 0 && vm.Height == 0 {
			runtime.Gosched()
		}
		s.SetNetStateOff(true)
		wait.Done()
		fmt.Println("Stopped FNode01")
	}

	go stop0()
	go stop1()
	wait.Wait()
	fmt.Println("Caused Elections")

	WaitMinutes(state2, 1)
	// bring them back
	RunCmd("0")
	RunCmd("x")
	RunCmd("1")
	RunCmd("x")
	// wait for him to update via dbstate and become an audit
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

func TestMultiple2Election(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 7, 2, 2, t)

	WaitForMinute(state0, 2)

	RunCmd("1")
	RunCmd("x")
	RunCmd("2")
	RunCmd("x")
	WaitForMinute(state0, 1)
	RunCmd("1")
	RunCmd("x")
	RunCmd("2")
	RunCmd("x")

	WaitBlocks(state0, 2)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	ShutDownEverything(t)

}

func TestMultiple3Election(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLLLLLAAAAF", map[string]string{}, 9, 3, 3, t)

	RunCmd("1")
	RunCmd("x")
	RunCmd("2")
	RunCmd("x")
	RunCmd("3")
	RunCmd("x")
	RunCmd("0")
	WaitMinutes(state0, 1)
	RunCmd("3")
	RunCmd("x")
	RunCmd("1")
	RunCmd("x")
	RunCmd("2")
	RunCmd("x")
	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 3)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	ShutDownEverything(t)

}

func TestSimCtrl(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	type walletcallHelper struct {
		Status string `json:"status"`
	}
	type walletcall struct {
		Jsonrpc string           `json:"jsonrps"`
		Id      int              `json:"id"`
		Result  walletcallHelper `json:"result"`
	}

	apiCall := func(state0 *state.State, cmd string) {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/debug"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method": "sim-ctrl", "params":{"commands": ["` + cmd + `"]}}`)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")
		if err != nil {
			t.Error(err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(walletcall)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		fmt.Println("resp2: ", resp2)
	}

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 8, 2, 2, t)

	WaitForMinute(state0, 2)
	apiCall(state0, "1")
	apiCall(state0, "x")
	apiCall(state0, "2")
	apiCall(state0, "x")
	WaitForMinute(state0, 1)
	apiCall(state0, "1")
	apiCall(state0, "x")
	apiCall(state0, "2")
	apiCall(state0, "x")

	apiCall(state0, "E")
	apiCall(state0, "F")
	apiCall(state0, "0")
	apiCall(state0, "p")

	WaitBlocks(state0, 2)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

func TestMultipleFTAccountsAPI(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true
	// only have one leader because if you are not the leader responcible for the FCT transaction then
	// you will return transACK before teh balance is updated which will make thsi test fail.
	state0 := SetupSim("LAF", map[string]string{}, 6, 0, 0, t)
	WaitForMinute(state0, 1)

	type walletcallHelper struct {
		CurrentHeight   uint32        `json:"currentheight"`
		LastSavedHeight uint          `json:"lastsavedheight"`
		Balances        []interface{} `json:"balances"`
	}
	type walletcall struct {
		Jsonrpc string           `json:"jsonrps"`
		Id      int              `json:"id"`
		Result  walletcallHelper `json:"result"`
	}

	type ackHelp struct {
		Jsonrpc string                       `json:"jsonrps"`
		Id      int                          `json:"id"`
		Result  wsapi.GeneralTransactionData `json:"result"`
	}

	apiCall := func(state0 *state.State, arrayOfFactoidAccounts []string) *walletcall {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method": "multiple-fct-balances", "params":{"addresses":["` + strings.Join(arrayOfFactoidAccounts, `", "`) + `"]}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(walletcall)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		return resp2
	}

	arrayOfFactoidAccounts := []string{"FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "FA3Y1tBWnFpyoZUPr9ZH51R1gSC8r5x5kqvkXL3wy4uRvzFnuWLB", "FA3Fsy2WPkR5z7qjpL8H1G51RvZLCiLDWASS6mByeQmHSwAws8K7"}
	resp2 := apiCall(state0, arrayOfFactoidAccounts)

	// To check if the balances returned from the API are right
	for i, a := range arrayOfFactoidAccounts {
		fmt.Println("state0.LLeaderHeight ", state0.LLeaderHeight)
		fmt.Println("state0.GetHighestSavedBlk() ", state0.GetHighestSavedBlk())
		currentHeight := state0.LLeaderHeight
		heighestSavedHeight := state0.GetHighestSavedBlk()
		errNotAcc := ""

		byteAcc := [32]byte{}
		copy(byteAcc[:], primitives.ConvertUserStrToAddress(a))

		PermBalance, pok := state0.FactoidBalancesP[byteAcc] // Gets the Balance of the Factoid address

		if state0.FactoidBalancesPapi != nil {
			if savedBal, ok := state0.FactoidBalancesPapi[byteAcc]; ok {
				PermBalance = savedBal
			}
		}

		pl := state0.ProcessLists.Get(currentHeight)
		pl.FactoidBalancesTMutex.Lock()
		// Gets the Temp Balance of the Factoid address
		TempBalance, tok := pl.FactoidBalancesT[byteAcc]
		pl.FactoidBalancesTMutex.Unlock()

		if tok != true && pok != true {
			TempBalance = 0
			PermBalance = 0
			errNotAcc = "Address has not had a transaction"
		} else if tok == true && pok == false {
			PermBalance = 0
			errNotAcc = ""
		} else if tok == false && pok == true {
			plLastHeight := state0.ProcessLists.Get(currentHeight - 1)
			plLastHeight.FactoidBalancesTMutex.Lock()
			TempBalanceLastHeight, tokLastHeight := plLastHeight.FactoidBalancesT[byteAcc] // Gets the Temp Balance of the Factoid address
			plLastHeight.FactoidBalancesTMutex.Unlock()
			if tokLastHeight == false {
				TempBalance = PermBalance
			} else {
				TempBalance = TempBalanceLastHeight
			}
		}

		x, ok := resp2.Result.Balances[i].(map[string]interface{})
		if ok != true {
			fmt.Println(x)
		}
		if resp2.Result.CurrentHeight != currentHeight || string(resp2.Result.LastSavedHeight) != string(heighestSavedHeight) {
			t.Fatalf("Who wrote this trash code?... Expected a current height of " + fmt.Sprint(currentHeight) + " and a saved height of " + fmt.Sprint(heighestSavedHeight) + " but got " + fmt.Sprint(resp2.Result.CurrentHeight) + ", " + fmt.Sprint(resp2.Result.LastSavedHeight))
		}

		if x["err"].(string) != errNotAcc {
			t.Fatalf("Expected err = \"%s\" but got \"%s\"", x["err"], errNotAcc)
		}
		if int64(x["ack"].(float64)) != TempBalance {
			t.Fatalf("Expected temp[%d] but got X[%d]<%f> ", TempBalance, int64(x["ack"].(float64)), x["ack"].(float64))
		}
		if int64(x["saved"].(float64)) != PermBalance {
			t.Fatalf("Expected perm[%d] but got X[%d]<%f>", PermBalance, int64(x["saved"].(float64)), x["saved"].(float64))
		}
	}
	TimeNow(state0)
	ToTestPermAndTempBetweenBlocks := []string{"FA3EPZYqodgyEGXNMbiZKE5TS2x2J9wF8J9MvPZb52iGR78xMgCb", "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q"}
	resp3 := apiCall(state0, ToTestPermAndTempBetweenBlocks)
	x, ok := resp3.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	//if x["ack"] != x["saved"] {
	//	t.Fatalf("Expected acknowledged and saved balances to be the same")
	//}
	if int64(x["ack"].(float64)) != int64(x["saved"].(float64)) {
		t.Fatalf("Expected  temp[%d] to match perm[%d]", int64(x["ack"].(float64)), int64(x["saved"].(float64)))
	}

	TimeNow(state0)

	_, str := FundWallet(state0, uint64(200*5e7))

	// a while loop to find when the transaction made FundWallet ^^Above^^ has been acknowledged
	thisShouldNotBeUnknownAtSomePoint := "Unknown"
	for thisShouldNotBeUnknownAtSomePoint != "TransactionACK" {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method":"factoid-ack", "params":{"txid":"` + str + `"}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(ackHelp)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		if resp2.Result.Status == "TransactionACK" {
			thisShouldNotBeUnknownAtSomePoint = resp2.Result.Status
		}
	}

	// This call should show a different acknowledged balance than the Saved Balance
	resp_5 := apiCall(state0, ToTestPermAndTempBetweenBlocks)
	x, ok = resp_5.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}

	//
	//if x["ack"] == x["saved"] {
	//	t.Fatalf("Expected acknowledged and saved balances to be different.")
	//}
	if int64(x["ack"].(float64)) == int64(x["saved"].(float64)) {
		t.Fatalf("Expected  temp[%d] to not match perm[%d]", int64(x["ack"].(float64)), int64(x["saved"].(float64)))
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

	resp_6 := apiCall(state0, ToTestPermAndTempBetweenBlocks)
	x, ok = resp_6.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	if x["ack"] != x["saved"] {
		t.Fatalf("Expected acknowledged and saved balances to be the same")
	}
	ShutDownEverything(t)
}

func TestMultipleECAccountsAPI(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// only have one leader because if you are not the leader responcible for the FCT transaction then
	// you will return transACK before teh balance is updated which will make thsi test fail.
	state0 := SetupSim("LAF", map[string]string{}, 6, 0, 0, t)
	WaitForMinute(state0, 1)

	type walletcallHelper struct {
		CurrentHeight   uint32        `json:"currentheight"`
		LastSavedHeight uint          `json:"lastsavedheight"`
		Balances        []interface{} `json:"balances"`
	}
	type walletcall struct {
		Jsonrpc string           `json:"jsonrps"`
		Id      int              `json:"id"`
		Result  walletcallHelper `json:"result"`
	}

	type GeneralTransactionData struct {
		Transid               string `json:"txis"`
		TransactionDate       int64  `json:"transactiondate,omitempty"`       //Unix time
		TransactionDateString string `json:"transactiondatestring,omitempty"` //ISO8601 time
		BlockDate             int64  `json:"blockdate,omitempty"`             //Unix time
		BlockDateString       string `json:"blockdatestring,omitempty"`       //ISO8601 time

		//Malleated *Malleated `json:"malleated,omitempty"`
		Status string `json:"status"`
	}

	type ackHelp struct {
		Jsonrpc string                 `json:"jsonrps"`
		Id      int                    `json:"id"`
		Result  GeneralTransactionData `json:"result"`
	}

	type ackHelpEC struct {
		Jsonrpc string            `json:"jsonrps"`
		Id      int               `json:"id"`
		Result  wsapi.EntryStatus `json:"result"`
	}

	apiCall := func(state0 *state.State, arrayOfECAccounts []string) *walletcall {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method": "multiple-ec-balances", "params":{"addresses":["` + strings.Join(arrayOfECAccounts, `", "`) + `"]}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(walletcall)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		return resp2
	}

	arrayOfECAccounts := []string{"EC1zGzM78psHhs5xVdv6jgVGmswvUaN6R3VgmTquGsdyx9W67Cqy", "EC1zGzM78psHhs5xVdv6jgVGmswvUaN6R3VgmTquGsdyx9W67Cqy"}
	resp2 := apiCall(state0, arrayOfECAccounts)

	// To check if the balances returned from the API are right
	for i, a := range arrayOfECAccounts {
		currentHeight := state0.LLeaderHeight
		heighestSavedHeight := state0.GetHighestSavedBlk()
		errNotAcc := ""

		byteAcc := [32]byte{}
		copy(byteAcc[:], primitives.ConvertUserStrToAddress(a))

		PermBalance, pok := state0.ECBalancesP[byteAcc] // Gets the Balance of the EC address

		if state0.ECBalancesPapi != nil {
			if savedBal, ok := state0.ECBalancesPapi[byteAcc]; ok {
				PermBalance = savedBal
			}
		}

		pl := state0.ProcessLists.Get(currentHeight)
		pl.ECBalancesTMutex.Lock()
		// Gets the Temp Balance of the Entry Credit address
		TempBalance, tok := pl.ECBalancesT[byteAcc]
		pl.ECBalancesTMutex.Unlock()

		if tok != true && pok != true {
			TempBalance = 0
			PermBalance = 0
			errNotAcc = "Address has not had a transaction"
		} else if tok == true && pok == false {
			PermBalance = 0
			errNotAcc = ""
		} else if tok == false && pok == true {
			plLastHeight := state0.ProcessLists.Get(currentHeight - 1)
			plLastHeight.FactoidBalancesTMutex.Lock()
			TempBalanceLastHeight, tokLastHeight := plLastHeight.FactoidBalancesT[byteAcc] // Gets the Temp Balance of the Factoid address
			plLastHeight.FactoidBalancesTMutex.Unlock()
			if tokLastHeight == false {
				TempBalance = PermBalance
			} else {
				TempBalance = TempBalanceLastHeight
			}
		}

		x, ok := resp2.Result.Balances[i].(map[string]interface{})
		if ok != true {
			fmt.Println(x)
		}

		if resp2.Result.CurrentHeight != currentHeight || string(resp2.Result.LastSavedHeight) != string(heighestSavedHeight) {
			t.Fatalf("Who wrote this trash code?... Expected a current height of " + fmt.Sprint(currentHeight) + " and a saved height of " + fmt.Sprint(heighestSavedHeight) + " but got " + fmt.Sprint(resp2.Result.CurrentHeight) + ", " + fmt.Sprint(resp2.Result.LastSavedHeight))
		}

		//for i := range x {
		//	fmt.Printf("%s: %v %T\n", i, x[i], x[i])
		//}

		if x["err"].(string) != errNotAcc {
			t.Fatalf("Expected err = \"%s\" but got \"%s\"", x["err"], errNotAcc)
		}
		if int64(x["ack"].(float64)) != TempBalance {
			t.Fatalf("Expected temp[%d] but got X[%d]<%f> ", TempBalance, int64(x["ack"].(float64)), x["ack"].(float64))
		}
		if int64(x["saved"].(float64)) != PermBalance {
			t.Fatalf("Expected perm[%d] but got X[%d]<%f>", PermBalance, int64(x["saved"].(float64)), x["saved"].(float64))
		}
	}
	TimeNow(state0)
	ToTestPermAndTempBetweenBlocks := []string{"EC1zGzM78psHhs5xVdv6jgVGmswvUaN6R3VgmTquGsdyx9W67Cqy", "EC3Eh7yQKShgjkUSFrPbnQpboykCzf4kw9QHxi47GGz5P2k3dbab"}
	resp3 := apiCall(state0, ToTestPermAndTempBetweenBlocks)
	x, ok := resp3.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}

	if int64(x["ack"].(float64)) != int64(x["saved"].(float64)) {
		t.Fatalf("Expected  temp[%d] to match perm[%d]", int64(x["ack"].(float64)), int64(x["saved"].(float64)))
	}

	TimeNow(state0)

	_, str := FundWallet(state0, 20000000)

	// a while loop to find when the transaction made FundWallet ^^Above^^ has been acknowledged
	for {
		url := "http://localhost:" + fmt.Sprint(state0.GetPort()) + "/v2"
		var jsonStr = []byte(`{"jsonrpc": "2.0", "id": 0, "method":"factoid-ack", "params":{"txid":"` + str + `"}}  `)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("content-type", "text/plain;")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		resp2 := new(ackHelp)
		err1 := json.Unmarshal([]byte(body), &resp2)
		if err1 != nil {
			t.Error(err1)
		}

		if resp2.Result.Status == "TransactionACK" {
			break
		}
	}

	// This call should show a different acknowledged balance than the Saved Balance
	resp_5 := apiCall(state0, ToTestPermAndTempBetweenBlocks)
	x, ok = resp_5.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	// looking at EC3Eh7yQKShgjkUSFrPbnQpboykCzf4kw9QHxi47GGz5P2k3dbab
	if int64(x["ack"].(float64)) == int64(x["saved"].(float64)) {
		t.Fatalf("Expected  temp[%d] to not match perm[%d]", int64(x["ack"].(float64)), int64(x["saved"].(float64)))
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

	resp_6 := apiCall(state0, ToTestPermAndTempBetweenBlocks)
	x, ok = resp_6.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	if x["ack"] != x["saved"] {
		t.Fatalf("Expected " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]) + " but got " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]))
	}
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

func TestDBSigElection(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LLLAF", map[string]string{"--faulttimeout": "10"}, 8, 1, 1, t)

	s := GetFnodes()[2].State
	if !s.IsLeader() {
		panic("Can't kill a audit and cause an election")
	}
	WaitForMinute(s, 9) // wait till the victim is at minute 9
	// wait till minute flips
	for s.CurrentMinute != 0 {
		runtime.Gosched()
	}
	s.SetNetStateOff(true) // kill the victim
	s.LogPrintf("faulting", "Stopped %s\n", s.FactomNodeName)
	WaitForMinute(state0, 2) // Wait till FNode0 move ahead a minute (the election is over)
	s.LogPrintf("faulting", "Start %s\n", s.FactomNodeName)
	s.SetNetStateOff(false) // resurrect the victim

	WaitBlocks(state0, 2)    // wait till the victim is back as the audit server
	WaitForMinute(state0, 1) // Wait till ablock is loaded
	WaitForAllNodes(state0)

	ShutDownEverything(t)
}

func TestCoinbaseCancel(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LFFFFF", map[string]string{"-blktime": "5"}, 30, 0, 0, t)
	// Make it quicker
	constants.COINBASE_PAYOUT_FREQUENCY = 2
	constants.COINBASE_DECLARATION = constants.COINBASE_PAYOUT_FREQUENCY * 2

	WaitMinutes(state0, 2)
	RunCmd("g10") // Adds 10 identities to your identity pool.
	WaitBlocks(state0, 2)
	// Assign identities
	RunCmd("1")
	RunCmd("t")
	RunCmd("2")
	RunCmd("t")
	RunCmd("3")
	RunCmd("t")
	RunCmd("4")
	RunCmd("t")
	RunCmd("5")
	RunCmd("t")

	WaitBlocks(state0, 2)
	// Promotions, create 3 feds and 3 audits
	RunCmd("1")
	RunCmd("l")
	RunCmd("2")
	RunCmd("l")
	RunCmd("3")
	RunCmd("o")
	RunCmd("4")
	RunCmd("o")
	RunCmd("5")
	RunCmd("o")

	WaitForBlock(state0, 15)
	WaitMinutes(state0, 1)
	// Cancel coinbase of 18 (14+ delay of 4) with a majority of the authority set, should succeed
	RunCmd("1")
	RunCmd("L14.1")
	RunCmd("2")
	RunCmd("L14.1")
	RunCmd("3")
	RunCmd("L14.1")
	RunCmd("4")
	RunCmd("L14.1")
	WaitForBlock(state0, 17)
	WaitMinutes(state0, 1)

	// attempt cancel coinbase of  20 (16+ delay of 4) without a majority of the authority set.  Should fail
	// This tests 3 of 6 canceling, which is not a majority (but almost is)
	// all feds
	RunCmd("0")
	RunCmd("L16.1")
	RunCmd("1")
	RunCmd("L16.1")
	RunCmd("2")
	RunCmd("L16.1")
	WaitForBlock(state0, 21)
	WaitForMinute(state0, 9)

	// attempt cancel coinbase of  22 (18+ delay of 4) without a majority of the authority set.  Should fail
	// This tests 3 of 6 canceling, which is not a majority (but almost is)
	// all audits
	RunCmd("3")
	RunCmd("L18.1")
	RunCmd("4")
	RunCmd("L18.1")
	RunCmd("5")
	RunCmd("L18.1")
	WaitForBlock(state0, 23)
	WaitForMinute(state0, 2)

	// attempt cancel coinbase of  24 (20+ delay of 4) without a majority of the authority set.  Should fail
	// This tests 3 of 6 canceling, which is not a majority (but almost is)
	// 2 audit 1 fed
	RunCmd("2")
	RunCmd("L20.1")
	RunCmd("4")
	RunCmd("L20.1")
	RunCmd("5")
	RunCmd("L20.1")
	WaitForBlock(state0, 25)
	WaitForMinute(state0, 2)

	// Check the coinbase blocks for correct number of outputs, indicating a successful (or correctly ignored) coinbase cancels

	hei := 18
	expected := 4
	f, err := state0.DB.FetchFBlockByHeight(uint32(hei))
	if err != nil {
		panic(fmt.Sprintf("Missing coinbase, admin block at height %d could not be retrieved", hei))
	}
	c := len(f.GetTransactions()[0].GetOutputs())
	if c != expected {
		t.Fatalf("Coinbase at height %d improperly cancelled.  should have %d outputs, but found %d", hei, expected, c)
	}

	hei = 20
	expected = 5
	f, err = state0.DB.FetchFBlockByHeight(uint32(hei))
	if err != nil {
		panic(fmt.Sprintf("Missing coinbase, admin block at height %d could not be retrieved", hei))
	}
	c = len(f.GetTransactions()[0].GetOutputs())
	if c != expected {
		t.Fatalf("Coinbase at height %d improperly cancelled.  should have %d outputs, but found %d", hei, expected, c)
	}

	hei = 22
	expected = 5
	f, err = state0.DB.FetchFBlockByHeight(uint32(hei))
	if err != nil {
		panic(fmt.Sprintf("Missing coinbase, admin block at height %d could not be retrieved", hei))
	}
	c = len(f.GetTransactions()[0].GetOutputs())
	if c != expected {
		t.Fatalf("Coinbase at height %d improperly cancelled.  should have %d outputs, but found %d", hei, expected, c)
	}

	hei = 24
	expected = 5
	f, err = state0.DB.FetchFBlockByHeight(uint32(hei))
	if err != nil {
		panic(fmt.Sprintf("Missing coinbase, admin block at height %d could not be retrieved", hei))
	}
	c = len(f.GetTransactions()[0].GetOutputs())
	if c != expected {
		t.Fatalf("Coinbase at height %d improperly cancelled.  should have %d outputs, but found %d", hei, expected, c)
	}

	//ShutDownEverythingWithoutAuthCheck(t)  see 9cf214e9140d767ea172b06a6e4b748475a9c494 for ShutDownEverythingWithoutAuthCheck()

}

func TestElection9(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LLAL", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 8, 1, 1, t)
	StatusEveryMinute(state0)
	CheckAuthoritySet(t)

	state3 := GetFnodes()[3].State
	if !state3.IsLeader() {
		panic("Can't kill a audit and cause an election")
	}
	RunCmd("3")
	WaitForMinute(state3, 9) // wait till the victim is at minute 9
	RunCmd("x")
	WaitMinutes(state0, 2) // Wait till fault completes
	RunCmd("x")

	WaitBlocks(state0, 2)    // wait till the victim is back as the audit server
	WaitForMinute(state0, 1) // Wait till ablock is loaded
	WaitForAllNodes(state0)
	WaitForMinute(state3, 1) // Wait till node 3 is following by minutes

	WaitForAllNodes(state0)
	ShutDownEverything(t)

}

func TestBadDBStateUnderflow(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true
	state0 := SetupSim("LF", map[string]string{}, 6, 0, 0, t)
	RunCmd("g1")
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)

	msg, err := state0.LoadDBState(state0.GetDBHeightComplete() - 1)
	if err != nil {
		panic(err)
	}
	dbs := msg.(*messages.DBStateMsg)
	dbs.DirectoryBlock.GetHeader().(*directoryBlock.DBlockHeader).DBHeight += 2
	m_dbs, err := dbs.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// replace the length of transaction in the marshaled datta with 0xdeadbeef!
	m_dbs = append(append(m_dbs[:659], []byte{0xde, 0xad, 0xbe, 0xef}...), m_dbs[663:]...)

	// i := 659
	// fmt.Printf("---%x---\n", m_dbs[i:i+4])

	s := hex.EncodeToString(m_dbs)
	wsapi.HandleV2SendRawMessage(state0, map[string]string{"message": s})

	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}
func TestFactoidDBState(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LAF", map[string]string{"--faulttimeout": "10", "--blktime": "5"}, 120, 0, 0, t)
	WaitBlocks(state0, 1)

	go func() {
		for i := 0; i <= 1000; i++ {
			FundWallet(state0, 10000)
			time.Sleep(time.Duration(random.RandIntBetween(250, 1250)) * time.Millisecond)
		}
	}()

	RunCmd("2")
	for i := 0; i < 20; i++ {
		WaitMinutes(state0, i)
		RunCmd("x")
		WaitMinutes(state0, 1+i)
		RunCmd("x")
		WaitBlocks(state0, 2)
	}
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

func TestNoMMR(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LLLAAFFFFF", map[string]string{}, 10, 0, 0, t)
	state.MMR_enable = false // turn off MMR processing
	StatusEveryMinute(state0)
	RunCmd("R10") // turn on some load
	WaitBlocks(state0, 5)
	RunCmd("R0") // turn off load
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

func TestDBStateCatchup(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LFF", map[string]string{}, 100, 0, 0, t)
	state.MMR_enable = false // turn off MMR processing
	state1 := GetFnodes()[1].State
	StatusEveryMinute(state1)

	WaitMinutes(state0, 2)

	RunCmd("1")
	RunCmd("x") // knock the follower offline

	RunCmd("R10") // turn on some load

	WaitBlocks(state0, 5)
	RunCmd("R0") // turn off load
	WaitMinutes(state0, 2)
	RunCmd("x") // bring the follower online
	WaitBlocks(state0, 7)

	WaitForAllNodes(state0) // if the follower isn't catching up this will timeout
	PrintOneStatus(0, 0)
	ShutDownEverything(t)
}

func TestDBState(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LLLFFFF", map[string]string{"--net": "line"}, 100, 0, 0, t)
	state1 := GetFnodes()[1].State
	state6 := GetFnodes()[6].State // Get node 6
	StatusEveryMinute(state1)

	WaitForMinute(state0, 8)
	RunCmd("Re")
	RunCmd("R4")
	RunCmd("F100")
	RunCmd("6")
	WaitForMinute(state6, 0)
	RunCmd("x")
	RunCmd("F0")
	WaitBlocks(state0, 5)
	RunCmd("x")
	WaitBlocks(state0, 5)
	RunCmd("R0")
	WaitBlocks(state0, 1)

	WaitForAllNodes(state0) // if the follower isn't catching up this will timeout
	PrintOneStatus(0, 0)
	ShutDownEverything(t)
}

func TestCatchupEveryMinute(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true
	//							  01234567890
	state0 := SetupSim("LFFFFFFFFFF", map[string]string{"--debuglog": ".", "--blktime": "6"}, 20, 1, 1, t)

	StatusEveryMinute(state0)

	// knock followers off one per minute
	for i := 0; i < 10; i++ {
		s := GetFnodes()[i+1].State
		RunCmd(fmt.Sprintf("%d", i+1))
		WaitForMinute(s, i)
		RunCmd("x")
	}
	state0.LogPrintf("test", "%s", atomic.WhereAmIString(0))
	WaitBlocks(state0, 2) // wait till they cannot catch up by MMR
	state0.LogPrintf("test", "%s", atomic.WhereAmIString(0))
	WaitMinutes(state0, 1)
	state0.LogPrintf("test", "%s", atomic.WhereAmIString(0))

	RunCmd("T25") // switch to 25 second blocks because dbstate catchup code fails at 6 second blocks
	// bring them all back
	for i := 0; i < 10; i++ {
		state0.LogPrintf("test", "%s %d %s", atomic.WhereAmIString(0), i)
		RunCmd(fmt.Sprintf("%d", i+1))
		WaitMinutes(state0, 1)
		RunCmd("x")
	}

	WaitForAllNodes(state0)
	ShutDownEverything(t)
}

func TestDebugLocation(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	tempdir := os.TempDir() + string(os.PathSeparator) + "logs" + string(os.PathSeparator) // get os agnostic path to the temp directory

	// toss any files that might preexist this run so we don't see old files
	err := os.RemoveAll(tempdir)
	if err != nil {
		panic(err)
	}

	// make sure the directory exists
	err = os.MkdirAll(tempdir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// start a sim with a select set of logs
	state0 := SetupSim("LF", map[string]string{"--debuglog": tempdir + "holding|networkinputs|ackqueue"}, 6, 0, 0, t)
	WaitBlocks(state0, 1)
	ShutDownEverything(t)

	// check the logs exist where we wanted them
	DoesFileExists(tempdir+"fnode0_holding.txt", t)
	DoesFileExists(tempdir+"fnode01_holding.txt", t)
	DoesFileExists(tempdir+"fnode0_networkinputs.txt", t)
	DoesFileExists(tempdir+"fnode01_networkinputs.txt", t)
	DoesFileExists(tempdir+"fnode01_ackqueue.txt", t)

	// toss the files we created since they are no longer needed
	err = os.RemoveAll(tempdir)
	if err != nil {
		panic(err)
	}

}

func TestDebugLocationParse(t *testing.T) {
	tempdir := os.TempDir() + string(os.PathSeparator) + "logs" + string(os.PathSeparator) // get os agnostic path to the temp directory
	stringsToCheck := []string{tempdir + "holding", tempdir + "networkinputs", tempdir + ".", tempdir + "ackqueue"}

	for i := 0; i < len(stringsToCheck); i++ {
		// Checks that the SplitUpDebugLogRegEx function works as expected
		dirlocation, regex := messages.SplitUpDebugLogRegEx(stringsToCheck[i])
		if dirlocation != tempdir {
			t.Fatalf("Error SplitUpDebugLogRegEx() did not return the correct directory location.")
		}
		if strings.Contains(regex, string(os.PathSeparator)) {
			t.Fatalf("Error SplitUpDebugLogRegEx() did not return the correct directory regex.")
		}
	}
}

func DoesFileExists(path string, t *testing.T) {
	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Error checking for File: %s", err)
	} else {
		t.Logf("Found file %s", path)
	}
	if os.IsNotExist(err) {
		t.Fatalf("File %s doesn't exist", path)
	} else {
		t.Logf("Found file %s", path)
	}

}

func LaggingAuditElection(t *testing.T, lag int, recovery int) {
	state0 := SetupSim("LLLAF", map[string]string{"--debuglog": ".", "--blktime": "15"}, 15, 0, 0, t)
	WaitForBlock(state0, 6)
	WaitForAllNodes(state0)

	RunCmd("3")
	RunCmd("x")
	WaitBlocks(state0, lag) // make audit lag behind

	RunCmd("1")
	RunCmd("x") // take out a leader

	RunCmd("3")
	RunCmd("x") // bring back audit

	WaitMinutes(state0, 2) // wait for audit to be elected

	RunCmd("1")
	RunCmd("x") // bring back leader-should become Audit

	// Do we need to wait by timeclock? rather than waitblocks
	WaitBlocks(state0, recovery) // give time to come back

	WaitForAllNodes(state0) // REVIEW: is this desired? rather than using WaitBlocks as above?

	AssertAuthoritySet(t, "LALLF") // leader
	Halt(t)
}

// test electing an audit that is 1 block behind
func TestLaggingAuditElection1(t *testing.T) {
	LaggingAuditElection(t, 1, 2) // KLUDGE: fails when recovery = 2
}

// test electing an audit that is 2 blocks behind
func TestLaggingAuditElection2(t *testing.T) {
	LaggingAuditElection(t, 2, 2)
}
