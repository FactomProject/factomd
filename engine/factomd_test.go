package engine_test

import (
	"os"
	"testing"
	"time"

	"fmt"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	"io/ioutil"
	"os/user"
)

var _ = Factomd

// Wait so many blocks
func WaitBlocks(s *state.State, blks int) {
	currentHeight := int(s.LLeaderHeight)
	for int(s.LLeaderHeight) < currentHeight+blks {
		time.Sleep(time.Second)
	}
}

// Wait to a given minute.  If we are == to the minute or greater, then
// we first wait to the start of the next block.
func WaitMinutes(s *state.State, min int) {
	if s.CurrentMinute >= min {
		for s.CurrentMinute > 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	for min > s.CurrentMinute {
		time.Sleep(100 * time.Millisecond)
	}
}

func TestSetupANetwork(t *testing.T) {

	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()

	startCap := func() {
		rescueStdout = os.Stdout
		r, w, _ = os.Pipe()
		os.Stdout = w
	}
	endCap := func() string {
		<-ProcessChan
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stdout = rescueStdout
		return string(out)
	}

	runCmd := func(cmd string) string {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		startCap()
		InputChan <- cmd
		//		time.Sleep(1000*time.Millisecond) // Uncommenting this makes us ill at this point.
		v := endCap()
		return v
	}

	usr, err := user.Current()

	if err != nil {
		panic(err)
	}

	if usr.Username == "clay" {

		done := make(chan struct{})

		timeout := func(seconds int, updatePeriod int) {
			for {
				for seconds > 0 {
					select {
					case <-done:
						return
					default:
						fmt.Printf("\nTimeout in %02d:%02d:%02d timeout\n", int(seconds/3600), int(seconds/60)%60, seconds%60)
						s := seconds % updatePeriod
						if s != 0 { // get delay aligned to period
							time.Sleep(time.Duration(s) * time.Second)
							seconds -= s
						} else {
							time.Sleep(time.Duration(updatePeriod) * time.Second)
							seconds -= updatePeriod
						}
					}
				}
				fmt.Println("Test Timeout")
				os.Exit(1)
			}
		}
		endTimeout := func() { done <- struct{}{} }
		go timeout(630, 10)
		defer endTimeout()
	}
	nodeCount := 10
	expectedLeaderCount := 3
	expectedAuditCount := 4
	expectedFollowerCount := nodeCount - expectedLeaderCount - expectedAuditCount // what's left

	args := append([]string{},
		"-db=Map",
		"-network=LOCAL",
		"-net=alot+",
		"-enablenet=true",
		"-blktime=20",
		"-count="+fmt.Sprintf("%d", nodeCount),
		"-logPort=37000",
		"-port=37001",
		"-ControlPanelPort=37002",
		"-networkPort=37003",
		"-startdelay=1",
		"faulttimeout=15",
		" -netdebug 4")

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(3 * time.Second)

	t.Logf("Allocated %d nodes", nodeCount)
	if len(GetFnodes()) != nodeCount {
		t.Fatalf("Should have allocated %d nodes", nodeCount)
		t.Fail()
	}

	runCmd("s") // start display of status
	WaitMinutes(state0, 3)
	runCmd("g10") // Create 10 identity (one FCT transaction and a pile of chain and entry creation)
	runCmd("9")   // select 9
	runCmd("x")   // take it offline
	runCmd("w")   // make the API point to current (for code coverage, there is no traffic)
	runCmd("10")
	runCmd("8")
	runCmd("w") // make the API point to 8 it will

	WaitBlocks(state0, 2) // wait till the dust settles
	// Allocate leaders
	WaitMinutes(state0, 1) // don't start at the beginning of the block (third minute absolute)
	runCmd("1")            // select node 1
	for i := 0; i < expectedLeaderCount-1; i++ {
		if i == 0 {
			runCmd("l") // make current node a leader, advance to next node
		} else {
			runCmd("") // Repeat make current node a leader, advance to next node
		}
	}

	// Allocate audit servers
	for i := 0; i < expectedAuditCount; i++ {
		runCmd("o") // make current node an audit, advance to next node
	}

	WaitBlocks(state0, 2)  // wait till the dust settles (relative one block)
	WaitMinutes(state0, 2) // don't start at the beginning of the block (third minute absolute)
	WaitMinutes(state0, 1) // don't start at the beginning of the block (third minute absolute)

	leadercnt := 0
	auditcnt := 0
	for _, fn := range GetFnodes() {
		s := fn.State
		if s.Leader {
			leadercnt++
		}
		list := s.ProcessLists.Get(s.LLeaderHeight)
		if foundAudit, _ := list.GetAuditServerIndexHash(s.GetIdentityChainID()); foundAudit {
			auditcnt++
		}
	}
	followercnt := nodeCount - leadercnt - auditcnt

	if leadercnt != expectedLeaderCount {
		t.Fatalf("found %d leaders, expected %d", leadercnt, expectedLeaderCount)
	}

	if auditcnt != expectedAuditCount {
		t.Fatalf("found %d audit servers, expected %d", auditcnt, expectedAuditCount)
		t.Fail()
	}

	if followercnt != expectedFollowerCount {
		t.Fatalf("found %d audit servers, expected %d", auditcnt, expectedAuditCount)
		t.Fail()
	}
	WaitMinutes(state0, 2)
	runCmd("F100")
	runCmd("S10")
	runCmd("g10")

	fn1 := GetFocus()
	expectedName := fmt.Sprintf("FNode%02d", expectedLeaderCount+expectedAuditCount)
	if fn1.State.FactomNodeName != expectedName {
		t.Fatalf("Expected %s, but got %s", expectedName, fn1.State.FactomNodeName)
	}
	runCmd("g1")
	WaitMinutes(state0, 3)
	runCmd("g1")
	WaitMinutes(state0, 4)
	runCmd("g1")
	WaitMinutes(state0, 5)
	runCmd("g1")
	WaitMinutes(state0, 6)
	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)
	runCmd("g1")
	WaitMinutes(state0, 2)
	runCmd("g1")
	WaitMinutes(state0, 3)
	runCmd("g20")
	WaitBlocks(state0, 1)
	runCmd("9")
	runCmd("x")
	runCmd("8")

	time.Sleep(100 * time.Millisecond)

	fn2 := GetFocus()
	if fn2.State.FactomNodeName != "FNode08" {
		t.Fatalf("Expected FNode08, but got %s", fn1.State.FactomNodeName)
	}

	runCmd("i")
	// Test block recording lengths and error checking for pprof
	runCmd("b100")

	runCmd("b")

	runCmd("babc")

	runCmd("b1000000")

	runCmd("/")

	runCmd("/")

	runCmd("a1")
	runCmd("e1")
	runCmd("d1")
	runCmd("f1")
	runCmd("a100")
	runCmd("e100")
	runCmd("d100")
	runCmd("f100")
	runCmd("yh")
	runCmd("yc")
	runCmd("r")
	WaitMinutes(state0, 1)
	runCmd("g1")
	runCmd("2")
	runCmd("x")
	WaitMinutes(state0, 1)
	runCmd("x")
	runCmd("g3")
	WaitBlocks(fn1.State, 1)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

	time.Sleep(10 * time.Second)
	if state0.LLeaderHeight > 13 {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}
}

