package engine_test

import (
	"os"
	"testing"
	"time"

	"fmt"

	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
)

var _ = Factomd

func TimeNow(s *state.State) {
	fmt.Printf("%s:%d/%d\n", s.FactomNodeName, int(s.LLeaderHeight), s.CurrentMinute)
}

// print the status for every minute for a state
func StatusEveryMinute(s *state.State) {
	go func() {
		for {
			WaitMinutes(s, 1)
			PrintOneStatus(0, 0)
		}
	}()
}

// Wait so many blocks
func WaitBlocks(s *state.State, blks int) {
	fmt.Printf("WaitBlocks(%d)\n", blks)
	TimeNow(s)
	newBlock := int(s.LLeaderHeight) + blks
	for int(s.LLeaderHeight) < newBlock {
		time.Sleep(time.Second)
	}
	TimeNow(s)
}

// Wait to a given minute.  If we are == to the minute or greater, then
// we first wait to the start of the next block.
func WaitForMinute(s *state.State, min int) {
	fmt.Printf("WaitForMinute(%d)\n", min)
	TimeNow(s)
	if s.CurrentMinute >= min {
		for s.CurrentMinute > 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	for min > s.CurrentMinute {
		time.Sleep(100 * time.Millisecond)
	}
	TimeNow(s)
}

// Wait some number of minutes
func WaitMinutes(s *state.State, min int) {
	fmt.Printf("WaitMinutes(%d)\n", min)
	TimeNow(s)
	newMinute := (s.CurrentMinute + min) % 10
	for s.CurrentMinute != newMinute {
		time.Sleep(100 * time.Millisecond)
	}
	TimeNow(s)
}

func TestSetupANetwork(t *testing.T) {

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		time.Sleep(100 * time.Millisecond)
		<-ProcessChan
		return
	}

	args := append([]string{},
		"-db=Map",
		"-network=LOCAL",
		"-net=alot+",
		"-enablenet=true",
		"-blktime=15",
		"-count=10",
		"-logPort=37000",
		"-port=37001",
		"-ControlPanelPort=37002",
		"-networkPort=37003",
		"-startdelay=1",
		"-faulttimeout=15")

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(3 * time.Second)

	t.Log("Allocated 10 nodes")
	if len(GetFnodes()) != 10 {
		t.Fatal("Should have allocated 10 nodes")
		t.Fail()
	}

	runCmd("s")
	runCmd("9")
	runCmd("x")
	runCmd("w")
	runCmd("10")
	runCmd("8")
	runCmd("w")
	WaitBlocks(state0, 1)
	runCmd("g10")
	WaitBlocks(state0, 3)
	// Allocate 4 leaders
	WaitForMinute(state0, 3)

	runCmd("1")
	runCmd("l")
	runCmd("")
	runCmd("")

	// Allocate 3 audit servers
	runCmd("o")
	runCmd("")
	runCmd("")

	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)

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

	if leadercnt != 4 {
		t.Fatalf("found %d leaders, expected 4", leadercnt)
	}

	if auditcnt != 3 {
		t.Fatalf("found %d audit servers, expected 3", auditcnt)
		t.Fail()
	}
	WaitForMinute(state0, 2)
	runCmd("F100")
	runCmd("S10")
	runCmd("g10")

	fn1 := GetFocus()
	if fn1.State.FactomNodeName != "FNode07" {
		t.Fatalf("Expected FNode07, but got %s", fn1.State.FactomNodeName)
	}
	runCmd("g1")
	WaitForMinute(state0, 3)
	runCmd("g1")
	WaitForMinute(state0, 4)
	runCmd("g1")
	WaitForMinute(state0, 5)
	runCmd("g1")
	WaitForMinute(state0, 6)
	WaitBlocks(state0, 1)
	WaitForMinute(state0, 1)
	runCmd("g1")
	WaitForMinute(state0, 2)
	runCmd("g1")
	WaitForMinute(state0, 3)
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
	WaitForMinute(state0, 1)
	runCmd("g1")
	runCmd("2")
	runCmd("x")
	WaitForMinute(state0, 3)
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

func TestAnElection(t *testing.T) {

	var (
		leaders   int = 3
		audits    int = 2
		followers int = 1
		nodes     int = leaders + audits + followers
	)

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		//		time.Sleep(100 * time.Millisecond)
		<-ProcessChan
		return
	}

	args := append([]string{},
		"-db=Map",
		"-network=LOCAL",
		"-net=alot+",
		"-enablenet=true",
		"-blktime=10",
		fmt.Sprintf("-count=%d", nodes),
		"-startdelay=1",
		"-faulttimeout=999999",
	)
	HandleLogfiles("out.txt", "out.txt")
	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(5 * time.Second) // wait till the simulation is setup
	LaunchDebugServer("localhost")

	t.Log(fmt.Sprintf("Allocated %d nodes", nodes))
	fnodes := GetFnodes()
	if len(fnodes) != nodes {
		t.Fatalf("Should have allocated %d nodes", nodes)
		t.Fail()
	}

	time.Sleep(1000 * time.Second)
	runCmd("g5")
	WaitBlocks(state0, 1)
	WaitMinutes(state0, 2)

	// Allocate leaders
	runCmd("1")
	for i := 0; i < leaders-1; i++ {
		runCmd("l")
	}

	// Allocate audit servers
	for i := 0; i < audits; i++ {
		runCmd("o")
	}

	WaitBlocks(state0, 2)
	WaitMinutes(state0, 2)
	PrintOneStatus(0, 0)

	CheckAuthoritySet(leaders, audits, t)

	runCmd(fmt.Sprintf("%d", leaders-1))
	runCmd("x")
	WaitBlocks(state0, 1)
	WaitMinutes(state0, 2)
	runCmd("x")

	WaitBlocks(state0, 300)
	WaitMinutes(state0, 2)

	PrintOneStatus(0, 0)
	if GetFnodes()[leaders-1].State.Leader {
		t.Fatalf("Node %d should not be a leader", leaders-1)
	}
	if !GetFnodes()[leaders].State.Leader && !GetFnodes()[leaders+1].State.Leader {
		t.Fatalf("Node %d or %d should be a leader", leaders, leaders+1)
	}

	CheckAuthoritySet(leaders, audits, t)

	// Now swap back...
	WaitBlocks(state0, 1)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

	// Sleep one block
	time.Sleep(time.Duration(state0.DirectoryBlockInSeconds) * time.Second)
	if state0.LLeaderHeight > 9 {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}
}

func CheckAuthoritySet(leaders int, audits int, t *testing.T) {
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
	if leadercnt != leaders {
		t.Fatalf("found %d leaders, expected %d", leadercnt, leaders)
	}
	if auditcnt != audits {
		t.Fatalf("found %d audit servers, expected %d", auditcnt, audits)
		t.Fail()
	}
}
