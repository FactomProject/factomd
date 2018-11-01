package engine_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives/random"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/activations"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/wsapi"
)

var _ = Factomd
var par = globals.FactomParams{}

var quit = make(chan struct{})

// SetupSim takes care of your options, and setting up nodes
// pass in a string for nodes: 4 Leaders, 3 Audit, 4 Followers: "LLLLAAAFFFF" as the first argument
// Pass in the Network type ex. "LOCAL" as the second argument
// It has default but if you want just add it like "map[string]string{"--Other" : "Option"}" as the third argument
// Pass in t for the testing as the 4th argument

var expectedHeight, leaders, audits, followers int
var startTime, endTime time.Time

//EX. state0 := SetupSim("LLLLLLLLLLLLLLLAAAAAAAAAA",  map[string]string {"--controlpanelsetting" : "readwrite"}, t)
func SetupSim(GivenNodes string, UserAddedOptions map[string]string, height int, electionsCnt int, RoundsCnt int, t *testing.T) *state.State {
	expectedHeight = height
	l := len(GivenNodes)
	CmdLineOptions := map[string]string{
		"--db":                  "Map",
		"--network":             "LOCAL",
		"--net":                 "alot+",
		"--enablenet":           "false",
		"--blktime":             "10",
		"--count":               fmt.Sprintf("%v", l),
		"--startdelay":          "1",
		"--stdoutlog":           "out.txt",
		"--stderrlog":           "out.txt",
		"--checkheads":          "false",
		"--controlpanelsetting": "readwrite",
		"--debuglog":            "faulting|bad",
		"--logPort":          "37000",
		"--port":             "37001",
		"--controlpanelport": "37002",
		"--networkport":      "37003",
	}

	// loop thru the test specific options and overwrite or append to the DefaultOptions
	if UserAddedOptions != nil && len(UserAddedOptions) != 0 {
		for key, value := range UserAddedOptions {
			if key != "--debuglog" {
				CmdLineOptions[key] = value
			} else {
				CmdLineOptions[key] = CmdLineOptions[key] + "|" + value // add debug log flags to the default
			}
			// remove options not supported by the current flags set so we can merge this update into older code bases
		}
	}
	// Finds all of the valid commands and stores them
	optionsArr := make(map[string]bool, 0)
	flag.VisitAll(func(key *flag.Flag) {
		optionsArr["--"+key.Name] = true
	})

	// Loops through CmdLineOptions to removed commands that are not valid
	for i, _ := range CmdLineOptions {
		_, ok := optionsArr[i]
		if !ok {
			fmt.Println("Not Included: " + i + ", Removing from Options")
			delete(CmdLineOptions, i)
		}
	}

	// default the fault time and round time based on the blk time out
	blktime, err := strconv.Atoi(CmdLineOptions["--blktime"])
	if err != nil {
		panic(err)
	}

	if CmdLineOptions["--faulttimeout"] == "" {
		CmdLineOptions["--faulttimeout"] = fmt.Sprintf("%d", blktime/5) // use 2 minutes ...
	}

	if CmdLineOptions["--roundtimeout"] == "" {
		CmdLineOptions["--roundtimeout"] = fmt.Sprintf("%d", blktime/5)
	}

	// built the fake command line
	returningSlice := []string{}
	for key, value := range CmdLineOptions {
		returningSlice = append(returningSlice, key+"="+value)
	}

	fmt.Println("Command Line Arguments:")
	for _, v := range returningSlice {
		fmt.Printf("\t%s\n", v)
	}
	params := ParseCmdLine(returningSlice)
	fmt.Println()

	fmt.Println("Parameter:")
	s := reflect.ValueOf(params).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %25s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())
	}
	fmt.Println()

	blkt := globals.Params.BlkTime
	roundt := elections.RoundTimeout
	et := elections.FaultTimeout
	startTime = time.Now()
	state0 := Factomd(params, false).(*state.State)
	//	statusState = state0
	calctime := time.Duration(float64((height*blkt)+(electionsCnt*et)+(RoundsCnt*roundt))*1.1) * time.Second
	endTime = time.Now().Add(calctime)
	fmt.Println("endTime: ", endTime.String(), "duration:", calctime.String())

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				if int(state0.GetLLeaderHeight()) > height {
					fmt.Printf("Test Timeout: Expected %d blocks\n", height)
					panic("Exceeded expected height")
				}
				if time.Now().After(endTime) {
					fmt.Printf("Test Timeout: Expected it to take %s\n", calctime.String())
					panic("TOOK TOO LONG")
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	state0.MessageTally = true
	fmt.Printf("Starting timeout timer:  Expected test to take %s or %d blocks\n", calctime.String(), height)
	//	StatusEveryMinute(state0)
	WaitMinutes(state0, 1) // wait till initial DBState message for the genesis block is processed
	creatingNodes(GivenNodes, state0)

	t.Logf("Allocated %d nodes", l)
	if len(GetFnodes()) != l {
		t.Fatalf("Should have allocated %d nodes", l)
		t.Fail()
	}
	CheckAuthoritySet(t)
	return state0
}

func creatingNodes(creatingNodes string, state0 *state.State) {
	runCmd(fmt.Sprintf("g%d", len(creatingNodes)))
	WaitMinutes(state0, 1)
	// Wait till all the entries from the g command are processed
	simFnodes := GetFnodes()
	nodes := len(simFnodes)
	for {
		iq := 0
		for _, s := range simFnodes {
			iq += s.State.InMsgQueue().Length()
		}
		iq2 := 0
		for _, s := range simFnodes {
			iq2 += s.State.InMsgQueue2().Length()
		}

		holding := 0
		for _, s := range simFnodes {
			holding += len(s.State.Holding)
		}

		pendingCommits := 0
		for _, s := range simFnodes {
			pendingCommits += s.State.Commits.Len()
		}
		if iq == 0 && iq2 == 0 && pendingCommits == 0 && holding == 0 {
			break
		}
		fmt.Printf("Waiting for g to complete iq == %d && iq2 == %d && pendingCommits == %d && holding == %d\n", iq, iq2, pendingCommits, holding)
		WaitMinutes(state0, 1)

	}
	WaitBlocks(state0, 1) // Wait for 1 block
	WaitForMinute(state0, 1)
	runCmd("0")
	for i, c := range []byte(creatingNodes) {
		switch c {
		case 'L', 'l':
			runCmd("l")
			leaders++
		case 'A', 'a':
			runCmd("o")
			audits++
		case 'F', 'f':
			runCmd(fmt.Sprintf("%d", (i+1)%nodes))
			followers++
			break
		default:
			panic("NOT L, A or F")
		}
	}
	WaitBlocks(state0, 1) // Wait for 1 block
	WaitForMinute(state0, 1)
}

func WaitForAllNodes(state *state.State) {
	height := ""
	simFnodes := GetFnodes()
	PrintOneStatus(0, 0) // Print a status
	fmt.Printf("Wait for all nodes done\n%s", height)
	prevblk := state.LLeaderHeight
	for i := 0; i < len(simFnodes); i++ {
		blk := state.LLeaderHeight
		if prevblk != blk {
			PrintOneStatus(0, 0)
			prevblk = blk
		}
		s := simFnodes[i].State
		height = ""
		if s.LLeaderHeight != blk { // if not caught up, start over
			time.Sleep(100 * time.Millisecond)
			i = 0 // start over
			continue
		}
		height = fmt.Sprintf("%s%s:%d-%d\n", height, s.FactomNodeName, s.LLeaderHeight, s.CurrentMinute)
	}
	fmt.Printf("Wait for all nodes done\n%s", height)
}

func TimeNow(s *state.State) {
	now := time.Now()
	fmt.Printf("%s:%d-:-%d Now %s of %s (remaining %s)\n", s.FactomNodeName, int(s.LLeaderHeight), s.CurrentMinute, now.Sub(startTime).String(), endTime.Sub(startTime).String(), endTime.Sub(now).String())
}

var statusState *state.State

// print the status for every minute for a state
func StatusEveryMinute(s *state.State) {
	if statusState == nil {
		fmt.Fprintf(os.Stdout, "Printing status from %s\n", s.FactomNodeName)
		statusState = s
		go func() {
			for {
				s := statusState
				newMinute := (s.CurrentMinute + 1) % 10
				timeout := 8 // timeout if a minutes takes twice as long as expected
				for s.CurrentMinute != newMinute && timeout > 0 {
					sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds
					time.Sleep(sleepTime * time.Millisecond)                       // wake up and about 4 times per minute
					timeout--
				}
				if timeout <= 0 {
					fmt.Println("Stalled !!!")
				}
				// Make all the nodes update their status
				for _, n := range GetFnodes() {
					n.State.SetString()
				}
				PrintOneStatus(0, 0)
			}
		}()
	} else {
		fmt.Fprintf(os.Stdout, "Printing status from %s", s.FactomNodeName)
		statusState = s

	}
}

// Wait till block = newBlock and minute = newMinute
func WaitForQuiet(s *state.State, newBlock int, newMinute int) {
	//	fmt.Printf("%s: %d-:-%d WaitFor(%d-:-%d)\n", s.FactomNodeName, s.LLeaderHeight, s.CurrentMinute, newBlock, newMinute)
	sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds
	if newBlock*10+newMinute < int(s.LLeaderHeight)*10+s.CurrentMinute {
		panic("Wait for the past")
	}
	for int(s.LLeaderHeight) < newBlock {
		x := int(s.LLeaderHeight)
		// wait for the next block
		for int(s.LLeaderHeight) == x {
			time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
		}
		if int(s.LLeaderHeight) < newBlock {
			TimeNow(s)
		}
	}

	// wait for the right minute
	for s.CurrentMinute != newMinute {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
}

func WaitMinutes(s *state.State, min int) {
	fmt.Printf("%s: %d-:-%d WaitMinutes(%d)\n", s.FactomNodeName, s.LLeaderHeight, s.CurrentMinute, min)
	newTime := int(s.LLeaderHeight)*10 + s.CurrentMinute + min
	newBlock := newTime / 10
	newMinute := newTime % 10
	WaitForQuiet(s, newBlock, newMinute)
	TimeNow(s)
}

// Wait so many blocks
func WaitBlocks(s *state.State, blks int) {
	fmt.Printf("%s: %d-:-%d WaitBlocks(%d)\n", s.FactomNodeName, s.LLeaderHeight, s.CurrentMinute, blks)
	newBlock := int(s.LLeaderHeight) + blks
	WaitForQuiet(s, newBlock, 0)
	TimeNow(s)
}

// Wait for a specific blocks
func WaitForBlock(s *state.State, newBlock int) {
	fmt.Printf("%s: %d-:-%d WaitForBlock(%d)\n", s.FactomNodeName, s.LLeaderHeight, s.CurrentMinute, newBlock)
	WaitForQuiet(s, newBlock, 0)
	TimeNow(s)
}

// Wait to a given minute.
func WaitForMinute(s *state.State, newMinute int) {
	fmt.Printf("%s: %d-:-%d WaitForMinute(%d)\n", s.FactomNodeName, s.LLeaderHeight, s.CurrentMinute, newMinute)
	if newMinute > 10 {
		panic("invalid minute")
	}
	newBlock := int(s.LLeaderHeight)
	if s.CurrentMinute > newMinute {
		newBlock++
	}
	WaitForQuiet(s, newBlock, newMinute)
	TimeNow(s)
}

func CheckAuthoritySet(t *testing.T) {
	leadercnt := 0
	auditcnt := 0
	followercnt := 0

	for _, fn := range GetFnodes() {
		s := fn.State
		if s.Leader {
			leadercnt++
		} else {
			list := s.ProcessLists.Get(s.LLeaderHeight)
			foundAudit, _ := list.GetAuditServerIndexHash(s.GetIdentityChainID())
			if foundAudit {
				auditcnt++
			} else {
				followercnt++
			}
		}
	}

	if leadercnt != leaders {
		t.Fatalf("found %d leaders, expected %d", leadercnt, leaders)
	}
	if auditcnt != audits {
		t.Fatalf("found %d audit servers, expected %d", auditcnt, audits)
		t.Fail()
	}
	if followercnt != followers {
		t.Fatalf("found %d followers, expected %d", followercnt, followers)
		t.Fail()
	}
}

// We can only run 1 simtest!
var ranSimTest = false

func runCmd(cmd string) {
	os.Stdout.WriteString("Executing: " + cmd + "\n")
	os.Stderr.WriteString("Executing: " + cmd + "\n")
	InputChan <- cmd
	return
}

func shutDownEverything(t *testing.T) {
	CheckAuthoritySet(t)
	quit <- struct{}{}
	close(quit)
	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}
	fnodes := GetFnodes()
	currentHeight := fnodes[0].State.LLeaderHeight
	// Sleep one block
	time.Sleep(time.Duration(globals.Params.BlkTime) * time.Second)

	if currentHeight < fnodes[0].State.LLeaderHeight {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}

	PrintOneStatus(0, 0) // Print a final status
	fmt.Printf("Test took %d blocks and %s time\n", GetFnodes()[0].State.LLeaderHeight, time.Now().Sub(startTime))

}
func v2Request(req *primitives.JSON2Request, port int) (*primitives.JSON2Response, error) {
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	portStr := fmt.Sprintf("%d", port)
	resp, err := http.Post(
		"http://localhost:"+portStr+"/v2",
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := primitives.NewJSON2Response()
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}
	return nil, nil
}

func TestSetupANetwork(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{}, 14, 0, 0, t)

	runCmd("9")  // Puts the focus on node 9
	runCmd("x")  // Takes Node 9 Offline
	runCmd("w")  // Point the WSAPI to send API calls to the current node.
	runCmd("10") // Puts the focus on node 9
	runCmd("8")  // Puts the focus on node 8
	runCmd("w")  // Point the WSAPI to send API calls to the current node.
	runCmd("7")
	WaitBlocks(state0, 1) // Wait for 1 block

	WaitForMinute(state0, 2) // Waits for minute 2
	runCmd("F100")           //  Set the Delay on messages from all nodes to 100 milliseconds
	runCmd("S10")            // Set Drop Rate to 1.0 on everyone
	runCmd("g10")            // Adds 10 identities to your identity pool.

	fn1 := GetFocus()
	PrintOneStatus(0, 0)
	if fn1.State.FactomNodeName != "FNode07" {
		t.Fatalf("Expected FNode07, but got %s", fn1.State.FactomNodeName)
	}
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	runCmd("g1")             // // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 4) // Waits for 4 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 5) // Waits for 5 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 6) // Waits for 6 "Minutes"
	WaitBlocks(state0, 1)    // Waits for 1 block
	WaitForMinute(state0, 1) // Waits for 1 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 2) // Waits for 2 "Minutes"
	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	runCmd("g20")            // Adds 20 identities to your identity pool.
	WaitBlocks(state0, 1)
	runCmd("9") // Focuses on Node 9
	runCmd("x") // Brings Node 9 back Online
	runCmd("8") // Focuses on Node 8

	time.Sleep(100 * time.Millisecond)

	fn2 := GetFocus()
	PrintOneStatus(0, 0)
	if fn2.State.FactomNodeName != "FNode08" {
		t.Fatalf("Expected FNode08, but got %s", fn1.State.FactomNodeName)
	}

	runCmd("i") // Shows the identities being monitored for change.
	// Test block recording lengths and error checking for pprof
	runCmd("b100") // Recording delays due to blocked go routines longer than 100 ns (0 ms)

	runCmd("b") // specifically how long a block will be recorded (in nanoseconds).  1 records all blocks.

	runCmd("babc") // Not sure that this does anything besides return a message to use "bnnn"

	runCmd("b1000000") // Recording delays due to blocked go routines longer than 1000000 ns (1 ms)

	runCmd("/") // Sort Status by Chain IDs

	runCmd("/") // Sort Status by Node Name

	runCmd("a1")             // Shows Admin block for Node 1
	runCmd("e1")             // Shows Entry credit block for Node 1
	runCmd("d1")             // Shows Directory block
	runCmd("f1")             // Shows Factoid block for Node 1
	runCmd("a100")           // Shows Admin block for Node 100
	runCmd("e100")           // Shows Entry credit block for Node 100
	runCmd("d100")           // Shows Directory block
	runCmd("f100")           // Shows Factoid block for Node 1
	runCmd("yh")             // Nothing
	runCmd("yc")             // Nothing
	runCmd("r")              // Rotate the WSAPI around the nodes
	WaitForMinute(state0, 1) // Waits 1 "Minute"

	runCmd("g1")             // Adds 1 identities to your identity pool.
	WaitForMinute(state0, 3) // Waits 3 "Minutes"
	WaitBlocks(fn1.State, 3) // Waits for 3 blocks

	shutDownEverything(t)
}

func TestLoad(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLF", map[string]string{"--debuglog": "."}, 15, 0, 0, t)

	runCmd("2")   // select 2
	runCmd("R30") // Feed load
	WaitBlocks(state0, 10)
	runCmd("R0") // Stop load
	WaitBlocks(state0, 1)
	shutDownEverything(t)
} // testLoad(){...}

func TestLoad2(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	go runCmd("Re") // Turn on tight allocation of EC as soon as the simulator is up and running
	state0 := SetupSim("LLLAAAFFF", map[string]string{"--debuglog": "."}, 24, 0, 0, t)
	StatusEveryMinute(state0)

	runCmd("7") // select node 1
	runCmd("x") // take out 7 from the network
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)

	runCmd("R30") // Feed load
	WaitBlocks(state0, 3)
	runCmd("Rt60")
	runCmd("T20")
	runCmd("R.5")
	WaitBlocks(state0, 2)
	runCmd("x")
	runCmd("R0")
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 3)

	ht7 := GetFnodes()[7].State.GetLLeaderHeight()
	ht6 := GetFnodes()[6].State.GetLLeaderHeight()

	if ht7 != ht6 {
		t.Fatalf("Node 7 was at dbheight %d which didn't match Node 6 at dbheight %d", ht7, ht6)
	}
	shutDownEverything(t)
} // testLoad2(){...}

// The intention of this test is to detect the EC overspend/duplicate commits (FD-566) bug.
// the bug happened when the FCT transaction and the commits arrived in different orders on followers vs the leader.
// Using a message delay, drop and tree network makes this likely
//
func TestLoadScrambled(t *testing.T) {
	if ranSimTest {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestLoadScrambled: %v", r)
		}
	}()

	ranSimTest = true

	// use a tree so the messages get reordered
	state0 := SetupSim("LLFFFFFF", map[string]string{"--net": "tree"}, 32, 0, 0, t)
	//TODO: Why does this run longer than expected?

	runCmd("2")     // select 2
	runCmd("F1000") // set the message delay
	runCmd("S10")   // delete 1% of the messages
	runCmd("r")     // rotate the load around the network
	runCmd("R3")    // Feed load
	WaitBlocks(state0, 10)
	runCmd("R0") // Stop load
	WaitBlocks(state0, 1)

	shutDownEverything(t)
} // testLoad(){...}

func TestMakeALeader(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LF", map[string]string{}, 5, 0, 0, t)

	runCmd("1") // select node 1
	runCmd("l") // make him a leader
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	// Adjust expectations
	leaders++
	followers--
	shutDownEverything(t)
}

func TestActivationHeightElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLLAAF", map[string]string{}, 16, 2, 2, t)

	// Kill the last two leader to cause a double election
	runCmd("3")
	runCmd("x")
	runCmd("4")
	runCmd("x")

	WaitMinutes(state0, 2) // make sure they get faulted

	// bring them back
	runCmd("3")
	runCmd("x")
	runCmd("4")
	runCmd("x")
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
	runCmd("5")
	runCmd("x")
	runCmd("6")
	runCmd("x")
	WaitMinutes(state0, 2) // make sure they get faulted
	// bring them back
	runCmd("5")
	runCmd("x")
	runCmd("6")
	runCmd("x")
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

	shutDownEverything(t)
}

func TestAnElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLAAF", map[string]string{}, 9, 1, 1, t)
	StatusEveryMinute(state0)
	WaitMinutes(state0, 2)

	runCmd("2")
	runCmd("w") // point the control panel at 2

	// remove the last leader
	runCmd("2")
	runCmd("x")
	// wait for the election
	WaitMinutes(state0, 2)
	//bring him back
	runCmd("x")
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
	shutDownEverything(t)

}

func TestDBsigEOMElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

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
	runCmd("0")
	runCmd("x")
	runCmd("1")
	runCmd("x")
	// wait for him to update via dbstate and become an audit
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	shutDownEverything(t)

}

func TestMultiple2Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLLAAF", map[string]string{"--debuglog": ".*"}, 7, 2, 2, t)

	WaitForMinute(state0, 2)

	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")
	WaitForMinute(state0, 1)
	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")

	runCmd("E")
	runCmd("F")
	runCmd("0")
	runCmd("p")

	WaitBlocks(state0, 2)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	shutDownEverything(t)

}

func TestMultiple3Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLLLLAAAAF", map[string]string{"--debuglog": ".*"}, 9, 3, 3, t)

	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")
	runCmd("3")
	runCmd("x")
	runCmd("0")
	WaitMinutes(state0, 1)
	runCmd("3")
	runCmd("x")
	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")
	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 3)
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0)
	shutDownEverything(t)

}

func TestMultiple7Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LLLLLLLLLLLLLLLAAAAAAAF", map[string]string{"--blktime": "25"}, 7, 7, 7, t)

	WaitForMinute(state0, 2)

	// Take 7 nodes off line
	for i := 1; i < 8; i++ {
		runCmd(fmt.Sprintf("%d", i))
		runCmd("x")
	}
	// force them all to be faulted
	WaitMinutes(state0, 1)

	// bring them back online
	for i := 1; i < 8; i++ {
		runCmd(fmt.Sprintf("%d", i))
		runCmd("x")
	}

	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	shutDownEverything(t)
}

func TestMultipleFTAccountsAPI(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{"--blktime": "15"}, 6, 0, 0, t)
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

	apiCall := func(arrayOfFactoidAccounts []string) *walletcall {
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
	resp2 := apiCall(arrayOfFactoidAccounts)

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
	resp3 := apiCall(ToTestPermAndTempBetweenBlocks)
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
	resp_5 := apiCall(ToTestPermAndTempBetweenBlocks)
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

	resp_6 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok = resp_6.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	if x["ack"] != x["saved"] {
		t.Fatalf("Expected acknowledged and saved balances to be the same")
	}
	shutDownEverything(t)
}

func TestMultipleECAccountsAPI(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	state0 := SetupSim("LLLLAAAFFF", map[string]string{"--blktime": "15"}, 6, 0, 0, t)
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

	apiCall := func(arrayOfECAccounts []string) *walletcall {
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
	resp2 := apiCall(arrayOfECAccounts)

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
	resp3 := apiCall(ToTestPermAndTempBetweenBlocks)
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
	resp_5 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok = resp_5.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}

	if int64(x["ack"].(float64)) == int64(x["saved"].(float64)) {
		t.Fatalf("Expected  temp[%d] to not match perm[%d]", int64(x["ack"].(float64)), int64(x["saved"].(float64)))
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

	resp_6 := apiCall(ToTestPermAndTempBetweenBlocks)
	x, ok = resp_6.Result.Balances[1].(map[string]interface{})
	if ok != true {
		fmt.Println(x)
	}
	if x["ack"] != x["saved"] {
		t.Fatalf("Expected " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]) + " but got " + fmt.Sprint(x["ack"]) + ", " + fmt.Sprint(x["saved"]))
	}
	WaitForAllNodes(state0)
	shutDownEverything(t)
}

func TestDBsigElectionEvery2Block_long(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	iterations := 1
	state := SetupSim("LLLLLLAF", map[string]string{"--debuglog": "fault|badmsg|network|process|dbsig", "--faulttimeout": "10"}, 32, 6, 6, t)

	runCmd("S10") // Set Drop Rate to 1.0 on everyone

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
	shutDownEverything(t)

}

func TestDBSigElection(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	state0 := SetupSim("LLLAF", map[string]string{"--debuglog": "fault|badmsg|network|process|dbsig", "--faulttimeout": "10"}, 8, 1, 1, t)

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

	shutDownEverything(t)
}

func makeExpected(grants []state.HardGrant) []interfaces.ITransAddress {
	var rval []interfaces.ITransAddress
	for _, g := range grants {
		rval = append(rval, factoid.NewOutAddress(g.Address, g.Amount))
	}
	return rval
}
func TestGrants_long(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "fault|badmsg|network|process|dbsig", "--faulttimeout": "10", "--blktime": "5"}, 300, 0, 0, t)

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
				printList("coinbase", coinBaseOutputs)
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

	shutDownEverything(t)
}

func printList(title string, list map[string]uint64) {
	for addr, amt := range list {
		fmt.Printf("%v - %v:%v\n", title, addr, amt)
	}
}

func TestCoinbaseCancel(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	state0 := SetupSim("LFFFFF", map[string]string{"-blktime": "5"}, 30, 0, 0, t)
	// Make it quicker
	constants.COINBASE_PAYOUT_FREQUENCY = 2
	constants.COINBASE_DECLARATION = constants.COINBASE_PAYOUT_FREQUENCY * 2

	WaitMinutes(state0, 2)
	runCmd("g10") // Adds 10 identities to your identity pool.
	WaitBlocks(state0, 1)
	// Assign identities
	runCmd("1")
	runCmd("t")
	runCmd("2")
	runCmd("t")
	runCmd("3")
	runCmd("t")
	runCmd("4")
	runCmd("t")
	runCmd("5")
	runCmd("t")

	WaitBlocks(state0, 2)
	// Promotions, create 3 feds and 3 audits
	runCmd("1")
	runCmd("l")
	runCmd("2")
	runCmd("l")
	runCmd("3")
	runCmd("o")
	runCmd("4")
	runCmd("o")
	runCmd("5")
	runCmd("o")

	WaitBlocks(state0, 3)
	WaitForBlock(state0, 15)
	WaitMinutes(state0, 1)
	// Cancel coinbase of 18 (14+ delay of 4) with a majority of the authority set, should succeed
	runCmd("1")
	runCmd("L14.1")
	runCmd("2")
	runCmd("L14.1")
	runCmd("3")
	runCmd("L14.1")
	runCmd("4")
	runCmd("L14.1")
	WaitForBlock(state0, 17)
	WaitMinutes(state0, 1)

	// attempt cancel coinbase of  20 (16+ delay of 4) without a majority of the authority set.  Should fail
	// This tests 3 of 6 canceling, which is not a majority (but almost is)
	// all feds
	runCmd("0")
	runCmd("L16.1")
	runCmd("1")
	runCmd("L16.1")
	runCmd("2")
	runCmd("L16.1")
	WaitForBlock(state0, 21)
	WaitForMinute(state0, 9)

	// attempt cancel coinbase of  22 (18+ delay of 4) without a majority of the authority set.  Should fail
	// This tests 3 of 6 canceling, which is not a majority (but almost is)
	// all audits
	runCmd("3")
	runCmd("L18.1")
	runCmd("4")
	runCmd("L18.1")
	runCmd("5")
	runCmd("L18.1")
	WaitForBlock(state0, 23)
	WaitForMinute(state0, 2)

	// attempt cancel coinbase of  24 (20+ delay of 4) without a majority of the authority set.  Should fail
	// This tests 3 of 6 canceling, which is not a majority (but almost is)
	// 2 audit 1 fed
	runCmd("2")
	runCmd("L20.1")
	runCmd("4")
	runCmd("L20.1")
	runCmd("5")
	runCmd("L20.1")
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

	//shutDownEverythingWithoutAuthCheck(t)  see 9cf214e9140d767ea172b06a6e4b748475a9c494 for shutDownEverythingWithoutAuthCheck()

}

func TestTestNetCoinBaseActivation_long(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	// reach into the activation an hack the TESTNET_COINBASE_PERIOD to be early so I can check it worked.
	activations.ActivationMap[activations.TESTNET_COINBASE_PERIOD].ActivationHeight["LOCAL"] = 22

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "fault|badmsg|network|process|dbsig", "--faulttimeout": "10", "--blktime": "10"}, 168, 0, 0, t)
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
	WaitForBlock(state0, 25)
	if constants.COINBASE_DECLARATION != 140 {
		t.Fatalf("constants.COINBASE_DECLARATION = %d expect 140\n", constants.COINBASE_DECLARATION)
	}

	nextBlock += oldCBDelay + 1
	fmt.Println("Wait till second grant should payout if the activation fails")
	WaitForBlock(state0, int(nextBlock+1)) // next old payout passed activation (should not be paid)
	CBT = factoidState0.GetCoinbaseTransaction(nextBlock, state0.GetLeaderTimestamp())
	if len(CBT.GetOutputs()) != 0 {
		t.Fatalf("because the payout delay changed there is no payout at block %d\n", nextBlock)
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
	shutDownEverything(t)
}

func TestElection9(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	state0 := SetupSim("LLAL", map[string]string{"--debuglog": ".|fault|badmsg|network|process|dbsig", "--faulttimeout": "10"}, 8, 1, 1, t)
	StatusEveryMinute(state0)
	CheckAuthoritySet(t)

	state3 := GetFnodes()[3].State
	if !state3.IsLeader() {
		panic("Can't kill a audit and cause an election")
	}
	runCmd("3")
	WaitForMinute(state3, 9) // wait till the victim is at minute 9
	runCmd("x")
	WaitMinutes(state0, 1) // Wait till fault completes
	runCmd("x")

	WaitBlocks(state0, 2)    // wait till the victim is back as the audit server
	WaitForMinute(state0, 1) // Wait till ablock is loaded
	WaitForAllNodes(state0)

	WaitForMinute(state3, 1) // Wait till node 3 is following by minutes

	WaitForAllNodes(state0)
	shutDownEverything(t)
}

// Cheap tests for developing binary search commits algorithm

func TestPass(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

}

func TestFail(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true
	t.Fatal("Failed")

}

func TestRandom(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	if random.RandUInt8() > 200 {
		t.Fatal("Failed")
	}
}

func TestFactoidDBState(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	state0 := SetupSim("LAF", map[string]string{"--debuglog": "fault|badmsg|network|process|dbsig", "--faulttimeout": "10", "--blktime": "5"}, 120, 0, 0, t)
	WaitForBlock(state0, 5)

	go func() {
		for i := 0; i <= 1000; i++ {
			FundWallet(state0, 10000)
			time.Sleep(time.Duration(random.RandIntBetween(250, 1250)) * time.Millisecond)
		}
	}()

	runCmd("2")
	for i := 0; i < 20; i++ {
		WaitMinutes(state0, i)
		runCmd("x")
		WaitMinutes(state0, 1+i)
		runCmd("x")
		WaitBlocks(state0, 2)
	}
	WaitForAllNodes(state0)
	shutDownEverything(t)
}

func TestNoMMR(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	state0 := SetupSim("LLLAAFFFFF", map[string]string{"--debuglog": "fault|badmsg|network|process|exec|missing", "--blktime": "20"}, 10, 0, 0, t)
	state.MMR_enable = false // turn off MMR processing
	StatusEveryMinute(state0)
	runCmd("R10") // turn on some load
	WaitBlocks(state0, 5)
	runCmd("R0") // turn off load
	WaitForAllNodes(state0)
	shutDownEverything(t)
}

// construct a new factoid transaction
func newTransaction(amt uint64, userSecretIn string, userPublicOut string, ecPrice uint64) (*factoid.Transaction, error) {

	inSec := factoid.NewAddress(primitives.ConvertUserStrToAddress(userSecretIn))
	outPub := factoid.NewAddress(primitives.ConvertUserStrToAddress(userPublicOut))

	var sec [64]byte
	copy(sec[:32], inSec.Bytes()) // pass 32 byte key in a 64 byte field for the crypto library

	pub := ed.GetPublicKey(&sec) // get the public key for our FCT source address

	rcd := factoid.NewRCD_1(pub[:]) // build the an RCD "redeem condition data structure"

	inAdd, err := rcd.GetAddress()
	if err != nil {
		panic(err)
	}

	trans := new(factoid.Transaction)
	trans.AddInput(inAdd, amt)
	trans.AddOutput(outPub, amt)

	/*
		userIn := primitives.ConvertFctAddressToUserStr(inAdd)
		userOut := primitives.ConvertFctAddressToUserStr(outPub)
		fmt.Printf("Txn %v %v -> %v\n", amt, userIn, userOut)
	*/

	// REVIEW: why is this different from engine.FundWallet() ?
	//trans.AddRCD(rcd)
	trans.AddAuthorization(rcd)
	trans.SetTimestamp(primitives.NewTimestampNow())

	fee, err := trans.CalculateFee(ecPrice)
	if err != nil {
		return trans, err
	}

	input, err := trans.GetInput(0)
	if err != nil {
		return trans, err
	}
	input.SetAmount(amt + fee)

	dataSig, err := trans.MarshalBinarySig()
	if err != nil {
		return trans, err
	}
	sig := factoid.NewSingleSignatureBlock(inSec.Bytes(), dataSig)
	trans.SetSignatureBlock(0, sig)

	return trans, nil

}

func AssertEquals(t *testing.T, a interface{}, b interface{}) {
	AssertEqualsMsg(t, a, b, "")
}

func AssertEqualsMsg(t *testing.T, a interface{}, b interface{}, msg string) {
	if a != b {
		t.Fatalf("%v != %v  %s", a, b, msg)
	}
}

func AssertNil(t *testing.T, a interface{}) {
	AssertEquals(t, a, nil)
}

func TestFeeTxnCreate(t *testing.T) {
	var oneFct uint64 = 100000000 // Factoshis
	var ecPrice uint64 = 10000

	balance := oneFct
	inUser := "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK" // FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q
	outAddress := "FA2s2SJ5Cxmv4MzpbGxVS9zbNCjpNRJoTX4Vy7EZaTwLq3YTur4u"

	for i := 0; i < 10; i++ {
		txn, _ := newTransaction(balance, inUser, outAddress, ecPrice)
		fee, _ := txn.CalculateFee(ecPrice)
		balance = balance - fee
		AssertEquals(t, 12*ecPrice, fee)
	}
}

func TestTxnCreate(t *testing.T) {
	var amt uint64 = 100000000
	var ecPrice uint64 = 10000

	inUser := "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK" // FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q
	//outUser := "Fs2GCfAa2HBKaGEUWCtw8eGDkN1CfyS6HhdgLv8783shkrCgvcpJ" // FA2s2SJ5Cxmv4MzpbGxVS9zbNCjpNRJoTX4Vy7EZaTwLq3YTur4u
	outAddress := "FA2s2SJ5Cxmv4MzpbGxVS9zbNCjpNRJoTX4Vy7EZaTwLq3YTur4u"

	txn, err := newTransaction(amt, inUser, outAddress, ecPrice)
	AssertNil(t, err)

	err = txn.ValidateSignatures()
	AssertNil(t, err)

	err = txn.Validate(1)
	AssertNil(t, err)

	if err := txn.Validate(0); err == nil {
		t.Fatalf("expected coinbase txn to error")
	}

	// test that we are sending to the address we thought
	AssertEquals(t, outAddress, txn.Outputs[0].GetUserAddress())

}

func sendTxn(s *state.State, amt uint64, userSecretIn string, userPubOut string, ecPrice uint64) (*factoid.Transaction, error) {
	txn, _ := newTransaction(amt, userSecretIn, userPubOut, ecPrice)
	msg := new(messages.FactoidTransaction)
	msg.SetTransaction(txn)
	s.APIQueue().Enqueue(msg)
	return txn, nil
}

func getBalance(s *state.State, userStr string) int64 {
	return s.FactoidState.GetFactoidBalance(factoid.NewAddress(primitives.ConvertUserStrToAddress(userStr)).Fixed())
}

// generate a pair of user-strings Fs.., FA..
func randomFctAddressPair() (string, string) {
	pkey := primitives.RandomPrivateKey()
	privUserStr, _ := primitives.PrivateKeyStringToHumanReadableFactoidPrivateKey(pkey.PrivateKeyString())
	_, _, pubUserStr,_ := factoid.PrivateKeyStringToEverythingString(pkey.PrivateKeyString())

	return privUserStr, pubUserStr
}

func TestProcessedBlockFailure(t *testing.T) {
	if ranSimTest {
		return
	}
	ranSimTest = true

	// a genesis block address w/ funding
	bankSecret := "Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"
	bankAddress := "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q"
	_ = bankAddress
	_ = bankSecret

	var depositSecrets []string
	var depositAddresses []string

	for i:=0; i<120; i++  {
		priv, addr := randomFctAddressPair()
		depositSecrets = append(depositSecrets, priv)
		depositAddresses = append(depositAddresses, addr)
	}

	var maxBlocks = 500
	state0 := SetupSim("LAF", map[string]string{"--debuglog": ".",}, maxBlocks+1, 0, 0, t)
	var ecPrice uint64 = state0.GetFactoshisPerEC() //10000
	var oneFct uint64 = factom.FactoidToFactoshi("1")

	waitForDeposit := func(i int, amt uint64) uint64 {
		balance := getBalance(state0, depositAddresses[i])
		TimeNow(state0)
		fmt.Printf("%v waitForDeposit %v %v - %v = diff: %v \n", i, depositAddresses[i], balance, amt, balance-int64(amt))
		var waited bool
		for balance != int64(amt) {
			waited = true
			balance = getBalance(state0, depositAddresses[i])
			time.Sleep(time.Millisecond*100)
		}
		if waited {
			fmt.Printf("%v waitForDeposit %v %v - %v = diff: %v \n", i, depositAddresses[i], balance, amt, balance-int64(amt))
			TimeNow(state0)
		}
		return uint64(balance)
	}
	_ = waitForDeposit

	initialBalance := 10*oneFct
	fee := 12*ecPrice

	prepareTransactions := func(bal uint64) ([]func(), uint64, int) {

		var transactions []func()
		var i int

		for i = 0; i < len(depositAddresses)-1; i += 1 {
			bal -= fee

			in := i
			out := i+1
			send := bal

			txn := func() {
				//fmt.Printf("TXN %v %v => %v \n", send, depositAddresses[in], depositAddresses[out])
				sendTxn(state0, send, depositSecrets[in], depositAddresses[out], ecPrice)
			}
			transactions = append(transactions, txn)
		}
		return transactions, bal, i
	}

	// offset to send initial blocking transaction
	offset := 1

	mkTransactions := func() { // txnGenerator
		// fund the start address
		sendTxn(state0, initialBalance, bankSecret, depositAddresses[0], ecPrice)
		WaitMinutes(state0, 1)
		waitForDeposit(0, initialBalance)
        transactions, finalBalance, finalAddress := prepareTransactions(initialBalance)

		var sent []int
        var unblocked bool = false

		for i:=1; i<len(transactions); i++ {
		    sent = append(sent, i)
		    //fmt.Printf("offset: %v <=> i:%v", offset, i)
		    if i == offset {
		    	fmt.Printf("\n==>TXN offset%v\n", offset)
				transactions[0]() // unblock the transactions
				unblocked = true
			}
			transactions[i]()
		}
		if ! unblocked{
			transactions[0]() // unblock the transactions
		}
		offset++ // next time start further in the future
		fmt.Printf("send chained transations")
		waitForDeposit(finalAddress, finalBalance)

		// empty final address returning remaining funds to bank
		sendTxn(state0, finalBalance-fee, depositSecrets[finalAddress], bankAddress, ecPrice)
		waitForDeposit(finalAddress, 0)
	}
	_ = mkTransactions

	for x:= 1; x<= 120; x++ {
		mkTransactions()
		WaitBlocks(state0, 1)
	}

	WaitForAllNodes(state0)
	shutDownEverything(t)
}
