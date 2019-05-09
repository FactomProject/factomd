package testHelper

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
)

var par = globals.FactomParams{}

var quit = make(chan struct{})

// SetupSim takes care of your options, and setting up nodes
// pass in a string for nodes: 4 Leaders, 3 Audit, 4 Followers: "LLLLAAAFFFF" as the first argument
// Pass in the Network type ex. "LOCAL" as the second argument
// It has default but if you want just add it like "map[string]string{"--Other" : "Option"}" as the third argument
// Pass in t for the testing as the 4th argument

var ExpectedHeight, Leaders, Audits, Followers int
var startTime, endTime time.Time
var RanSimTest = true // KLUDGE disables all sim tests during a group unit test run
// NOTE: going forward breaking out a test into a file under ./simTest allows it to run on CI.

//EX. state0 := SetupSim("LLLLLLLLLLLLLLLAAAAAAAAAA",  map[string]string {"--controlpanelsetting" : "readwrite"}, t)
func SetupSim(GivenNodes string, UserAddedOptions map[string]string, height int, electionsCnt int, RoundsCnt int, t *testing.T) *state.State {
	fmt.Println("SetupSim(", GivenNodes, ",", UserAddedOptions, ",", height, ",", electionsCnt, ",", RoundsCnt, ")")
	ExpectedHeight = height
	l := len(GivenNodes)

	dirBase, _ := os.Getwd()
	dirBase = dirBase + "/.sim/"
	os.Mkdir(dirBase, 0600)
	factomHome := dirBase+GetTestName()
	os.Setenv("FACTOM_HOME", factomHome)

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
		"--debuglog":            ".|faulting|bad",
		"--logPort":             "37000",
		"--port":                "37001",
		"--controlpanelport":    "37002",
		"--networkport":         "37003",
		"--factomhome":          factomHome,
	}

	// loop thru the test specific options and overwrite or append to the DefaultOptions
	if UserAddedOptions != nil && len(UserAddedOptions) != 0 {
		for key, value := range UserAddedOptions {
			if key != "--debuglog" && value != "" {
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
	params := engine.ParseCmdLine(returningSlice)
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
	state0 := engine.Factomd(params, false).(*state.State)
	//	statusState = state0
	calctime := time.Duration(float64(((height+3)*blkt)+(electionsCnt*et)+(RoundsCnt*roundt))*1.1) * time.Second
	endTime = time.Now().Add(calctime)
	fmt.Println("endTime: ", endTime.String(), "duration:", calctime.String())

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				if int(state0.GetLLeaderHeight())-3 > height { // always give us 3 extra block to finish
					fmt.Printf("Test Timeout: Expected %d blocks (%s)\n", height, calctime.String())
					panic(fmt.Sprintf("Test Timeout: Expected %d blocks (%s)\n", height, calctime.String()))
				}
				if time.Now().After(endTime) {
					fmt.Printf("Test Timeout: Expected it to take %s (%d blocks)\n", calctime.String(), height)
					panic(fmt.Sprintf("Test Timeout: Expected it to take %s (%d blocks)\n", calctime.String(), height))
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	state0.MessageTally = true

	fmt.Printf("Starting timeout timer:  Expected test to take %s or %d blocks\n", calctime.String(), height)
	StatusEveryMinute(state0)
	WaitMinutes(state0, 1) // wait till initial DBState message for the genesis block is processed
	creatingNodes(GivenNodes, state0, t)

	t.Logf("Allocated %d nodes", l)
	if len(engine.GetFnodes()) != l {
		t.Fatalf("Should have allocated %d nodes", l)
		t.Fail()
	}
	if Audits == 0 && Leaders == 0 {
		// if no requested promotions then assume we loaded from a database and the test will check
	} else {
		CheckAuthoritySet(t)
	}
	return state0
}

func creatingNodes(creatingNodes string, state0 *state.State, t *testing.T) {
	RunCmd(fmt.Sprintf("g%d", len(creatingNodes)))
	WaitBlocks(state0, 3) // Wait for 2 blocks because ID scans is for block N-1
	WaitMinutes(state0, 1)
	// Wait till all the entries from the g command are processed
	simFnodes := engine.GetFnodes()
	nodes := len(simFnodes)
	if len(creatingNodes) > nodes {
		t.Fatalf("Should have allocated %d nodes", len(creatingNodes))
	}
	WaitForMinute(state0, 1)
	for i, c := range []byte(creatingNodes) {
		fmt.Println("it:", i, c)
		switch c {
		case 'L', 'l':
			if i != 0 {
				RunCmd(fmt.Sprintf("%d", i))
				RunCmd("l")
			}
			Leaders++
		case 'A', 'a':
			RunCmd(fmt.Sprintf("%d", i))
			RunCmd("o")
			Audits++
		case 'F', 'f':
			Followers++
			break
		default:
			panic("NOT L, A or F")
		}
	}
	WaitBlocks(state0, 1) // Wait for 1 block
	WaitForMinute(state0, 1)
	WaitForAllNodes(state0) // make sure everyone is caught up
}

func WaitForAllNodes(state *state.State) {
	height := ""
	simFnodes := engine.GetFnodes()
	engine.PrintOneStatus(0, 0) // Print a status
	fmt.Printf("Wait for all nodes done\n%s", height)
	block := state.LLeaderHeight
	minute := state.CurrentMinute
	target := int(block*10) + minute

	for i := 0; i < len(simFnodes); i++ {
		s := simFnodes[i].State
		h := int(s.LLeaderHeight*10) + s.CurrentMinute

		if !s.GetNetStateOff() && h < target { // if not caught up, start over
			fmt.Printf("WaitForAllNodes: Waiting on FNode%2d\n", i)
			time.Sleep(100 * time.Millisecond)
			i--
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
				for _, n := range engine.GetFnodes() {
					n.State.SetString()
				}

				engine.PrintOneStatus(0, 0)
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

func CountAuthoritySet() (int, int, int) {
	leadercnt := 0
	auditcnt := 0
	followercnt := 0

	for i, fn := range engine.GetFnodes() {
		s := fn.State
		if s.Leader {
			fmt.Printf("Found Leader   %d %x\n", i, s.GetIdentityChainID().Bytes()[3:6])
			leadercnt++
		} else {
			list := s.ProcessLists.Get(s.LLeaderHeight)
			foundAudit, _ := list.GetAuditServerIndexHash(s.GetIdentityChainID())
			if foundAudit {
				fmt.Printf("Found Audit     %d %x\n", i, s.GetIdentityChainID().Bytes()[3:6])
				auditcnt++
			} else {
				fmt.Printf("Found Follower %d %x\n", i, s.GetIdentityChainID().Bytes()[3:6])

				followercnt++
			}
		}
	}

	return leadercnt, auditcnt, followercnt
}

func AdjustAuthoritySet(adjustingNodes string) {
	lead := Leaders
	audit := Audits
	follow := Followers

	for _, c := range []byte(adjustingNodes) {
		switch c {
		case 'L':
			lead--
		case 'A':
			audit--
		case 'F':
			follow--
			break
		default:
			panic("NOT L, A or F")
		}
	}

	fmt.Printf("AdjustAuthoritySet DIFF: L: %v, F: %v, A: %v\n", lead, audit, follow)
	Leaders = Leaders - lead
	Audits = Audits - audit
	Followers = Followers - follow
}

func isAuditor(fnode int) bool {
	nodes := engine.GetFnodes()
	list := nodes[0].State.ProcessLists.Get(nodes[0].State.LLeaderHeight)
	foundAudit, _ := list.GetAuditServerIndexHash(nodes[fnode].State.GetIdentityChainID())
	return foundAudit
}

func isFollower(fnode int) bool {
	return !(isAuditor(fnode) || engine.GetFnodes()[fnode].State.Leader)
}

func AssertAuthoritySet(t *testing.T, givenNodes string) {
	nodes := engine.GetFnodes()
	for i, c := range []byte(givenNodes) {
		switch c {
		case 'L':
			assert.True(t, nodes[i].State.Leader, "Expected node %v to be a leader", i)
		case 'A':
			assert.True(t, isAuditor(i), "Expected node %v to be an auditor", i)
		default:
			assert.True(t, isFollower(i), "Expected node %v to be a follower", i)
		}
	}
}

func CheckAuthoritySet(t *testing.T) {

	leadercnt, auditcnt, followercnt := CountAuthoritySet()

	if leadercnt != Leaders {
		engine.PrintOneStatus(0, 0)
		t.Fatalf("found %d leaders, expected %d", leadercnt, Leaders)
	}
	if auditcnt != Audits {
		engine.PrintOneStatus(0, 0)
		t.Fatalf("found %d audit servers, expected %d", auditcnt, Audits)
		t.Fail()
	}
	if followercnt != Followers {
		engine.PrintOneStatus(0, 0)
		t.Fatalf("found %d followers, expected %d", followercnt, Followers)
		t.Fail()
	}
}

func RunCmd(cmd string) {
	os.Stdout.WriteString("Executing: " + cmd + "\n")
	globals.InputChan <- cmd
	return
}

func Halt(t *testing.T) {
	quit <- struct{}{}
	close(quit)
	t.Log("Shutting down the network")
	for _, fn := range engine.GetFnodes() {
		fn.State.ShutdownChan <- 1
	}
	// sleep long enough for everyone to see the shutdown.
	time.Sleep(time.Duration(globals.Params.BlkTime) * time.Second)
}

func ShutDownEverything(t *testing.T) {
	CheckAuthoritySet(t)
	Halt(t)
	fnodes := engine.GetFnodes()
	currentHeight := fnodes[0].State.LLeaderHeight
	// Sleep one block
	time.Sleep(time.Duration(globals.Params.BlkTime) * time.Second)

	if currentHeight < fnodes[0].State.LLeaderHeight {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}

	engine.PrintOneStatus(0, 0) // Print a final status
	fmt.Printf("Test took %d blocks and %s time\n", engine.GetFnodes()[0].State.LLeaderHeight, time.Now().Sub(startTime))
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

func ResetFactomHome(t *testing.T) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	homeDir := dir + "/.sim/" + GetTestName()

	t.Logf("Removing old test run in %s", homeDir)
	os.MkdirAll(homeDir, 0755)
	os.RemoveAll(homeDir)

	os.Setenv("FACTOM_HOME", homeDir)
	os.MkdirAll(homeDir+"/.factom/m2", 0755)
	return string(homeDir), nil
}

func AddFNode() {
	engine.AddNode()
	Followers++
}
