package testHelper

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)
// SetupSim takes care of your options, and setting up nodes
// pass in a string for nodes: 4 Leaders, 3 Audit, 4 Followers: "LLLLAAAFFFF" as the first argument
// Pass in the Network type ex. "LOCAL" as the second argument
// It has default but if you want just add it like "map[string]string{"--Other" : "Option"}" as the third argument
// Pass in t for the testing as the 4th argument

var par = globals.FactomParams{}

var quit = make(chan struct{})

var ExpectedHeight, Leaders, Audits, Followers int
var startTime, endTime time.Time

//EX. state0 := SetupSim("LLLLLLLLLLLLLLLAAAAAAAAAA",  map[string]string {"--controlpanelsetting" : "readwrite"}, t)
func SetupSim(GivenNodes string, UserAddedOptions map[string]string, height int, electionsCnt int, RoundsCnt int, t *testing.T) *state.State {
	ExpectedHeight = height
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
	if len(engine.GetFnodes()) != l {
		t.Fatalf("Should have allocated %d nodes", l)
		t.Fail()
	}
	CheckAuthoritySet(t)
	return state0
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

	for _, fn := range engine.GetFnodes() {
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

	if leadercnt != Leaders {
		t.Fatalf("found %d leaders, expected %d", leadercnt, Leaders)
	}
	if auditcnt != Audits {
		t.Fatalf("found %d audit servers, expected %d", auditcnt, Audits)
		t.Fail()
	}
	if followercnt != Followers {
		t.Fatalf("found %d followers, expected %d", followercnt, Followers)
		t.Fail()
	}
}

// We can only run 1 simtest!
var RanSimTest = false

func RunCmd(cmd string) {
	os.Stdout.WriteString("Executing: " + cmd + "\n")
	os.Stderr.WriteString("Executing: " + cmd + "\n")
	engine.InputChan <- cmd
	return
}

func ShutDownEverything(t *testing.T) {
	CheckAuthoritySet(t)
	quit <- struct{}{}
	close(quit)
	t.Log("Shutting down the network")
	for _, fn := range engine.GetFnodes() {
		fn.State.ShutdownChan <- 1
	}
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


func creatingNodes(creatingNodes string, state0 *state.State) {
	RunCmd(fmt.Sprintf("g%d", len(creatingNodes)))
	WaitMinutes(state0, 1)
	// Wait till all the entries from the g command are processed
	simFnodes := engine.GetFnodes()
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
	RunCmd("0")
	for i, c := range []byte(creatingNodes) {
		switch c {
		case 'L', 'l':
			RunCmd("l")
			Leaders++
		case 'A', 'a':
			RunCmd("o")
			Audits++
		case 'F', 'f':
			RunCmd(fmt.Sprintf("%d", (i+1)%nodes))
			Followers++
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
	simFnodes := engine.GetFnodes()
	engine.PrintOneStatus(0, 0) // Print a status
	fmt.Printf("Wait for all nodes done\n%s", height)
	prevblk := state.LLeaderHeight
	for i := 0; i < len(simFnodes); i++ {
		blk := state.LLeaderHeight
		if prevblk != blk {
			engine.PrintOneStatus(0, 0)
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

