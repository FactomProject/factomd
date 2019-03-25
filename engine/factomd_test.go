package engine_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/primitives/random"

	"github.com/FactomProject/factomd/activations"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/wsapi"
)

func TestSetupANetwork(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{"--debuglog": "", "--blktime": "15"}, 14, 0, 0, t)

	RunCmd("9")  // Puts the focus on node 9
	RunCmd("x")  // Takes Node 9 Offline
	RunCmd("w")  // Point the WSAPI to send API calls to the current node.
	RunCmd("10") // Puts the focus on node 9
	RunCmd("8")  // Puts the focus on node 8
	RunCmd("w")  // Point the WSAPI to send API calls to the current node.
	RunCmd("7")
	WaitBlocks(state0, 1) // Wait for 1 block

	WaitForMinute(state0, 2) // Waits for minute 2
	RunCmd("F100")           //  Set the Delay on messages from all nodes to 100 milliseconds
	RunCmd("S10")            // Set Drop Rate to 1.0 on everyone
	RunCmd("g10")            // Adds 10 identities to your identity pool.

	fn1 := GetFocus()
	PrintOneStatus(0, 0)
	if fn1.State.FactomNodeName != "FNode07" {
		t.Fatalf("Expected FNode07, but got %s", fn1.State.FactomNodeName)
	}
	RunCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	RunCmd("g1")             // // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 4) // Waits for 4 "Minutes"
	RunCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 5) // Waits for 5 "Minutes"
	RunCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 6) // Waits for 6 "Minutes"
	WaitBlocks(state0, 1)    // Waits for 1 block
	WaitForMinute(state0, 1) // Waits for 1 "Minutes"
	RunCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 2) // Waits for 2 "Minutes"
	RunCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	RunCmd("g20")            // Adds 20 identities to your identity pool.
	WaitBlocks(state0, 1)
	RunCmd("9") // Focuses on Node 9
	RunCmd("x") // Brings Node 9 back Online
	RunCmd("8") // Focuses on Node 8

	time.Sleep(100 * time.Millisecond)

	fn2 := GetFocus()
	PrintOneStatus(0, 0)
	if fn2.State.FactomNodeName != "FNode08" {
		t.Fatalf("Expected FNode08, but got %s", fn1.State.FactomNodeName)
	}

	RunCmd("i") // Shows the identities being monitored for change.
	// Test block recording lengths and error checking for pprof
	RunCmd("b100") // Recording delays due to blocked go routines longer than 100 ns (0 ms)

	RunCmd("b") // specifically how long a block will be recorded (in nanoseconds).  1 records all blocks.

	RunCmd("babc") // Not sure that this does anything besides return a message to use "bnnn"

	RunCmd("b1000000") // Recording delays due to blocked go routines longer than 1000000 ns (1 ms)

	RunCmd("/") // Sort Status by Chain IDs

	RunCmd("/") // Sort Status by Node Name

	RunCmd("a1")             // Shows Admin block for Node 1
	RunCmd("e1")             // Shows Entry credit block for Node 1
	RunCmd("d1")             // Shows Directory block
	RunCmd("f1")             // Shows Factoid block for Node 1
	RunCmd("a100")           // Shows Admin block for Node 100
	RunCmd("e100")           // Shows Entry credit block for Node 100
	RunCmd("d100")           // Shows Directory block
	RunCmd("f100")           // Shows Factoid block for Node 1
	RunCmd("yh")             // Nothing
	RunCmd("yc")             // Nothing
	RunCmd("r")              // Rotate the WSAPI around the nodes
	WaitForMinute(state0, 1) // Waits 1 "Minute"

	RunCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits 3 "Minutes"
	WaitBlocks(fn1.State, 3) // Waits for 3 blocks

	ShutDownEverything(t)

}

func TestLoad(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLF", map[string]string{"--debuglog": ""}, 15, 0, 0, t)

	RunCmd("2")   // select 2
	RunCmd("R30") // Feed load
	WaitBlocks(state0, 10)
	RunCmd("R0") // Stop load
	WaitBlocks(state0, 1)
	ShutDownEverything(t)
} // testLoad(){...}

// Test that we don't put invalid TX into a block.  This is done by creating transactions that are just outside
// the time for the block, and we let the block catch up.  The code should validate against the block time of the
// block to ensure that we don't record an invalid transaction in the block relative to the block time.
func TestTXTimestampsAndBlocks(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	go RunCmd("Re") // Turn on tight allocation of EC as soon as the simulator is up and running
	state0 := SetupSim("LLLAAAFFF", map[string]string{"--blktime": "20"}, 24, 0, 0, t)
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

	go RunCmd("Re") // Turn on tight allocation of EC as soon as the simulator is up and running
	state0 := SetupSim("LLLAAAFFF", map[string]string{"--debuglog": "."}, 24, 0, 0, t)
	StatusEveryMinute(state0)

	RunCmd("7") // select node 1
	RunCmd("x") // take out 7 from the network
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

	ht7 := GetFnodes()[7].State.GetLLeaderHeight()
	ht6 := GetFnodes()[6].State.GetLLeaderHeight()

	if ht7 != ht6 {
		t.Fatalf("Node 7 was at dbheight %d which didn't match Node 6 at dbheight %d", ht7, ht6)
	}
	ShutDownEverything(t)
} // testLoad2(){...}
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
} // testLoad(){...}

func TestMakeALeader(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LF", map[string]string{}, 5, 0, 0, t)

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

func TestActivationHeightElection(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 16, 2, 2, t)

	// Kill the last two leader to cause a double election
	RunCmd("3")
	RunCmd("x")
	RunCmd("4")
	RunCmd("x")

	WaitMinutes(state0, 2) // make sure they get faulted

	// bring them back
	RunCmd("3")
	RunCmd("x")
	RunCmd("4")
	RunCmd("x")
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	CheckAuthoritySet(t)

	if GetFnodes()[3].State.Leader {
		t.Fatalf("Node 3 should not be a leader")
	}
	if GetFnodes()[4].State.Leader {
		t.Fatalf("Node 4 should not be a leader")
	}
	if !GetFnodes()[5].State.Leader {
		t.Fatalf("Node 5 should be a leader")
	}
	if !GetFnodes()[6].State.Leader {
		t.Fatalf("Node 6 should be a leader")
	}

	CheckAuthoritySet(t)

	if state0.IsActive(activations.ELECTION_NO_SORT) {
		t.Fatalf("ELECTION_NO_SORT active too early")
	}

	for !state0.IsActive(activations.ELECTION_NO_SORT) {
		WaitBlocks(state0, 1)
	}

	WaitForMinute(state0, 2) // Don't Fault at the end of a block

	// Cause a new double elections by killing the new leaders
	RunCmd("5")
	RunCmd("x")
	RunCmd("6")
	RunCmd("x")
	WaitMinutes(state0, 2) // make sure they get faulted
	// bring them back
	RunCmd("5")
	RunCmd("x")
	RunCmd("6")
	RunCmd("x")
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	CheckAuthoritySet(t)

	if GetFnodes()[5].State.Leader {
		t.Fatalf("Node 5 should not be a leader")
	}
	if GetFnodes()[6].State.Leader {
		t.Fatalf("Node 6 should not be a leader")
	}
	if !GetFnodes()[3].State.Leader {
		t.Fatalf("Node 3 should be a leader")
	}
	if !GetFnodes()[4].State.Leader {
		t.Fatalf("Node 4 should be a leader")
	}

	ShutDownEverything(t)
}
func TestAnElection(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLAAF", map[string]string{}, 9, 1, 1, t)

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

	state0 := SetupSim("LLLLLAAF", map[string]string{"--debuglog": ""}, 7, 2, 2, t)

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

	state0 := SetupSim("LLLLLLLAAAAF", map[string]string{"--debuglog": ""}, 9, 3, 3, t)

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

	state0 := SetupSim("LLLLLAAF", map[string]string{"--debuglog": "."}, 8, 2, 2, t)

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

func TestMultiple7Election(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLLLLLLLFLLFLFLLLFLAAFAAAAFA", map[string]string{"--blktime": "60", "--debuglog": "."}, 10, 7, 7, t)

	WaitForMinute(state0, 2)

	// Take 7 nodes off line
	for i := 1; i < 8; i++ {
		RunCmd(fmt.Sprintf("%d", i))
		RunCmd("x")
	}
	// force them all to be faulted
	WaitMinutes(state0, 1)

	// bring them back online
	for i := 1; i < 8; i++ {
		RunCmd(fmt.Sprintf("%d", i))
		RunCmd("x")
	}

	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
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
	state0 := SetupSim("LAF", map[string]string{"--blktime": "15"}, 6, 0, 0, t)
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
	state0 := SetupSim("LAF", map[string]string{"--blktime": "15"}, 6, 0, 0, t)
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

func TestDBsigElectionEvery2Block_long(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	iterations := 1
	state := SetupSim("LLLLLLAF", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 35, 6, 6, t)

	RunCmd("S10") // Set Drop Rate to 1.0 on everyone

	for j := 0; j < iterations; j++ {
		// for leader 1 thu 7 kill each in turn
		for i := 1; i < 7; i++ {
			s := GetFnodes()[i].State
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
			WaitForMinute(state, 1) // Wait till FNode0 move ahead a minute (the election is over)
			s.LogPrintf("faulting", "Start %s\n", s.FactomNodeName)
			s.SetNetStateOff(false) // resurrect the victim

			// Wait till the should have updated by DBSTATE
			WaitBlocks(state, 2)
			WaitForMinute(state, 1)
			WaitForAllNodes(state)

			CheckAuthoritySet(t) // check the authority set is as expected
		}
	}
	WaitForAllNodes(state)
	ShutDownEverything(t)

}

func TestDBSigElection(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	state0 := SetupSim("LLLAF", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 8, 1, 1, t)

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

func TestGrants_long(t *testing.T) {
	if RanSimTest {
		return
	}

	makeExpected := func(grants []state.HardGrant) []interfaces.ITransAddress {
		var rval []interfaces.ITransAddress
		for _, g := range grants {
			rval = append(rval, factoid.NewOutAddress(g.Address, g.Amount))
		}
		return rval
	}

	RanSimTest = true

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "", "--faulttimeout": "10", "--blktime": "5"}, 300, 0, 0, t)

	grants := state.GetHardCodedGrants()

	// find all the heights we care about
	heights := map[uint32][]state.HardGrant{}
	min := uint32(9999999)
	max := uint32(0)
	grantBalances := map[string]int64{} // Compute the expected final balances
	// TODO: (does not account for cancels)
	for _, g := range grants {
		heights[g.DBh] = append(heights[g.DBh], g)
		if min > g.DBh {
			min = g.DBh
		}
		if max < g.DBh {
			max = g.DBh
		}
		// keep a list of grant addresses
	}

	// Build a list of grant addresses
	for _, g := range grants {
		userAddr := primitives.ConvertFctAddressToUserStr(g.Address)
		_, ok := grantBalances[userAddr]
		if !ok {
			grantBalances[userAddr] = state0.FactoidState.GetFactoidBalance(g.Address.Fixed()) // Save initial balance
		}
		grantBalances[userAddr] += int64(g.Amount) // Add the grant amount
	}

	fmt.Println("Waiting for grant payout")
	// run the state till we are past the 100 block delay and check the final balances
	WaitBlocks(state0, int(max+1+constants.COINBASE_DECLARATION+constants.COINBASE_PAYOUT_FREQUENCY*2))

	// check the final balances of the accounts
	for addr, balance := range grantBalances {
		factoidBalance := state0.FactoidState.GetFactoidBalance(factoid.NewAddress(primitives.ConvertUserStrToAddress(addr)).Fixed())
		if balance != factoidBalance {
			t.Errorf("FinalBalanceMismatch for %s. Got %d expected %d", addr, balance, factoidBalance)
		}
	}

	// loop thru the dbheights  to get the admin block and check them and make sure the payouts get returned
	for dbheight := uint32(min - constants.COINBASE_PAYOUT_FREQUENCY*2); dbheight <= uint32(max+constants.COINBASE_PAYOUT_FREQUENCY*2); dbheight++ {
		expected := makeExpected(heights[dbheight])
		gotGrants := state.GetGrantPayoutsFor(dbheight)
		if len(expected) != len(gotGrants) {
			t.Errorf("Expected %d grants but found %d", len(expected), len(gotGrants))
		} else if len(expected) > 0 {
			fmt.Printf("Got %d expected grants at %d\n", len(expected), dbheight)
		}

		for i, _ := range expected {
			if !expected[i].GetAddress().IsSameAs(gotGrants[i].GetAddress()) ||
				expected[i].GetAmount() != gotGrants[i].GetAmount() ||
				expected[i].GetUserAddress() != gotGrants[i].GetUserAddress() {
				t.Errorf("Expected: %v ", expected[i])
				t.Errorf("but found %v for grant #%d at %d", gotGrants[i], i, dbheight)
			} else {
				fmt.Printf("Got grants %v\n", expected[i])
			}
			//fmt.Println(p.GetAmount(), p.GetUserAddress())
		}
		//descriptorHeight := dbheight - constants.COINBASE_DECLARATION

		ablock, err := state0.DB.FetchABlockByHeight(dbheight)
		if err != nil {
			panic(fmt.Sprintf("Missing coinbase, admin block at height %d could not be retrieved", dbheight))
		}

		abe := ablock.FetchCoinbaseDescriptor()
		if abe != nil {
			desc := abe.(*adminBlock.CoinbaseDescriptor)
			coinBaseOutputs := map[string]uint64{}
			for _, o := range desc.Outputs {
				coinBaseOutputs[primitives.ConvertFctAddressToUserStr(o.GetAddress())] = o.GetAmount()
			}
			if len(expected) != len(coinBaseOutputs) && !(len(coinBaseOutputs) == 1 && dbheight%constants.COINBASE_PAYOUT_FREQUENCY == 0) {
				t.Errorf("Expected %d grants but found %d at height %d", len(expected), len(coinBaseOutputs), dbheight)
				PrintList("coinbase", coinBaseOutputs)
			}
			for i, _ := range expected {
				address := expected[i].GetUserAddress()
				cbAmount := coinBaseOutputs[address]
				amount := expected[i].GetAmount()
				if amount != cbAmount {
					t.Errorf("Expected: %v ", expected[i])
					t.Errorf("but found %v:%v for grant #%d at %d", address, cbAmount, i, dbheight)
				}
				//fmt.Println(p.GetAmount(), p.GetUserAddress())
			}
		}
	} // for all dbheights {...}

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

func TestTestNetCoinBaseActivation_long(t *testing.T) {
	if RanSimTest {
		return
	}

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "", "--faulttimeout": "10"}, 168, 0, 0, t)
	fmt.Println("Simulation configured")
	nextBlock := uint32(11 + constants.COINBASE_DECLARATION) // first grant is at 11 so it pays at 21
	fmt.Println("Wait till first grant should payout")
	WaitForBlock(state0, int(nextBlock)) // wait for the first coin base payout to be generated
	factoidState0 := state0.FactoidState.(*state.FactoidState)
	CBT := factoidState0.GetCoinbaseTransaction(nextBlock, state0.GetLeaderTimestamp())
	oldCBDelay := constants.COINBASE_DECLARATION
	if oldCBDelay != 10 {
		t.Fatalf("constants.COINBASE_DECLARATION = %d expect 10\n", constants.COINBASE_DECLARATION)
	}
	if len(CBT.GetOutputs()) != 1 {
		t.Fatalf("Expected first payout at block %d\n", nextBlock)
	} else {
		fmt.Println("Got first payout")
	}

	fmt.Println("Wait till activation height")
	blk := activations.ActivationMap[activations.TESTNET_COINBASE_PERIOD].ActivationHeight["LOCAL"]
	WaitForBlock(state0, blk)
	if constants.COINBASE_DECLARATION != 140 {
		t.Fatalf("constants.COINBASE_DECLARATION = %d expect 140\n", constants.COINBASE_DECLARATION)
	}

	nextBlock += constants.COINBASE_DECLARATION - oldCBDelay + 1
	fmt.Println("Wait till second grant should payout with the new activation height")
	WaitForBlock(state0, int(nextBlock+1)) // next payout passed new activation (should be paid)
	CBT = factoidState0.GetCoinbaseTransaction(nextBlock, state0.GetLeaderTimestamp())
	if len(CBT.GetOutputs()) != 0 {
		t.Fatalf("Expected first payout at block %d\n", nextBlock)
	}
	fmt.Println("Wait to shut down")
	StatusEveryMinute(state0)
	WaitForAllNodes(state0)
	ShutDownEverything(t)
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
	WaitMinutes(state0, 1) // Wait till fault completes
	RunCmd("x")

	WaitBlocks(state0, 2)    // wait till the victim is back as the audit server
	WaitForMinute(state0, 1) // Wait till ablock is loaded
	WaitForAllNodes(state0)
	WaitForMinute(state3, 1) // Wait till node 3 is following by minutes

	WaitForAllNodes(state0)
	ShutDownEverything(t)
}
func TestRandom(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	if random.RandUInt8() > 200 {
		t.Fatal("Failed")
	}

}

func TestBadDBStateUnderflow(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true
	state0 := SetupSim("LF", map[string]string{}, 6, 0, 0, t)

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

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "", "--faulttimeout": "10", "--blktime": "5"}, 120, 0, 0, t)
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

	state0 := SetupSim("LLLAAFFFFF", map[string]string{"--debuglog": "", "--blktime": "20"}, 10, 0, 0, t)
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

	state0 := SetupSim("LFF", map[string]string{"--debuglog": "", "--blktime": "10"}, 100, 0, 0, t)
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

	state0 := SetupSim("LLLFFFF", map[string]string{"--net": "line", "--debuglog": ".", "--blktime": "10"}, 100, 0, 0, t)
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

func SystemCall(cmd string) {
	fmt.Println("SystemCall(\"", cmd, "\")")
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		foo := err.Error()
		fmt.Println(foo)
		os.Exit(1)
		panic(err)
	}
	fmt.Print(string(out))
}

func TestDebugLocation(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	os.MkdirAll("../../logs", os.ModePerm)

	state0 := SetupSim("LF", map[string]string{"--debuglog": "../../logs/."}, 5, 0, 0, t)

	RunCmd("1") // select node 1
	RunCmd("l") // make him a leader
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	// Adjust expectations
	Leaders++
	Followers--

	DoesFileExists("../../logs/fnode0_holding.txt", t);
	DoesFileExists("../../logs/fnode01_holding.txt", t);
	DoesFileExists("../../logs/fnode0_networkinputs.txt", t);
	DoesFileExists("../../logs/fnode01_networkinputs.txt", t);
	DoesFileExists("../../logs/fnode0_election.txt", t);
	DoesFileExists("../../logs/fnode01_election.txt", t);
	DoesFileExists("../../logs/fnode0_ackqueue.txt", t);
	DoesFileExists("../../logs/fnode01_ackqueue.txt", t);



	ShutDownEverything(t)
}

func DoesFileExists(path string, t *testing.T) {
	_, err := os.Stat(path)
	if err != nil { t.Fatalf("Error checking for File: ", err) } else { fmt.Println("We good!") }
	if os.IsNotExist(err) { t.Fatalf("File doesn't exist ")} else { fmt.Println("We good!") }

}