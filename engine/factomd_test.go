package engine_test

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/FactomProject/factomd/activations"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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
			// Make all the nodes update thier status
			for _, n := range GetFnodes() {
				n.State.SetString()
			}
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
		TimeNow(s)
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
func WaitMinutesQuite(s *state.State, min int) {
	sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds

	newMinute := (s.CurrentMinute + min) % 10
	newBlock := int(s.LLeaderHeight) + (s.CurrentMinute+min)/10
	for int(s.LLeaderHeight) < newBlock {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
	for s.CurrentMinute != newMinute {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
}

func WaitMinutes(s *state.State, min int) {
	fmt.Printf("WaitMinutes(%d)\n", min)
	TimeNow(s)
	WaitMinutesQuite(s, min)
	TimeNow(s)
}

// We can only run 1 simtest!
var ranSimTest = false

func TestSetupANetwork(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--net=alot+",
		"--enablenet=false",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=10",
		"--logPort=37000",
		"--port=37001",
		"--controlpanelport=37002",
		"--networkport=37003",
		"--startdelay=1",
		//"--debuglog=.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		"--checkheads=false",
	)

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
	WaitBlocks(state0, 1)
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
	PrintOneStatus(0, 0)
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
	PrintOneStatus(0, 0)
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
	PrintOneStatus(0, 0)
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
	WaitForMinute(state0, 3)
	WaitBlocks(fn1.State, 3)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

	time.Sleep(10 * time.Second)
	PrintOneStatus(0, 0)
	dblim := 12
	if state0.LLeaderHeight > uint32(dblim) {
		t.Fatalf("Failed to shut down factomd via ShutdownChan expected DBHeight %d got %d", dblim, state0.LLeaderHeight)
	}

}

func TestLoad(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		os.Stdout.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--enablenet=false",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=2",
		"--startdelay=1",
		//"--debuglog=F.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
	)

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(3 * time.Second)
	StatusEveryMinute(state0)
	t.Log("Allocated 2 nodes")
	if len(GetFnodes()) != 2 {
		t.Fatal("Should have allocated 2 nodes")
		t.Fail()
	}

	WaitForMinute(state0, 3)
	runCmd("g3")
	WaitBlocks(state0, 1)
	// Allocate 1 leaders
	WaitForMinute(state0, 1)

	runCmd("1") // select node 1
	runCmd("l") // make 1 a leader
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

	if leadercnt != 2 {
		t.Fatalf("found %d leaders, expected 2", leadercnt)
	}

	runCmd("2")   // select 2
	runCmd("R30") // Feed load
	WaitBlocks(state0, 30)
	runCmd("R0") // Stop load
	WaitBlocks(state0, 1)

} // testLoad(){...}

func TestMakeALeader(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		os.Stdout.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--enablenet=false",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=2",
		"--startdelay=1",
		//"--debuglog=F.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		"--checkheads=false",
	)

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(3 * time.Second)
	StatusEveryMinute(state0)
	t.Log("Allocated 2 nodes")
	if len(GetFnodes()) != 2 {
		t.Fatal("Should have allocated 2 nodes")
		t.Fail()
	}

	WaitForMinute(state0, 3)
	runCmd("g1")
	WaitBlocks(state0, 1)
	// Allocate 1 leaders
	WaitForMinute(state0, 1)

	runCmd("1") // select node 1
	runCmd("l") // make him a leader
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

	if leadercnt != 2 {
		t.Fatalf("found %d leaders, expected 2", leadercnt)
	}
	WaitMinutes(state0, 2)
}

func TestActivationHeightElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	var (
		leaders   int = 5
		audits    int = 2
		followers int = 1
		nodes     int = leaders + audits + followers
	)

	runCmd := func(cmd string) {
		os.Stdout.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--net=alot+",
		"--enablenet=false",
		"--logPort=37000",
		"--port=37001",
		"--controlpanelport=37002",
		"--networkport=37003",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		fmt.Sprintf("--count=%d", nodes),
		"--startdelay=1",
		"--debuglog=F.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		"--checkheads=false",
	)
	params := ParseCmdLine(args)

	time.Sleep(5 * time.Second) // wait till the control panel is setup
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(5 * time.Second) // wait till the simulation is setup

	t.Log(fmt.Sprintf("Allocated %d nodes", nodes))
	fnodes := GetFnodes()
	if len(fnodes) != nodes {
		t.Fatalf("Should have allocated %d nodes", nodes)
		t.Fail()
	}

	StatusEveryMinute(state0)
	WaitMinutes(state0, 2)

	runCmd(fmt.Sprintf("g%d", nodes))
	WaitMinutes(state0, 5)
	for {
		pendingCommits := 0
		for _, s := range fnodes {
			pendingCommits += s.State.Commits.Len()
		}
		if pendingCommits == 0 {
			break
		}
		fmt.Printf("Waiting for G command to complete\n")
		WaitMinutes(state0, 1)

	}
	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)
	// Allocate leaders
	runCmd("1")
	for i := 0; i < leaders-1; i++ {
		runCmd("l")
	}

	// Allocate audit servers
	for i := 0; i < audits; i++ {
		runCmd("o")
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 2)
	PrintOneStatus(0, 0)

	CheckAuthoritySet(leaders, audits, t)

	// Kill the last two leader to cause a double election
	runCmd(fmt.Sprintf("%d", leaders-2))
	runCmd("x")
	runCmd(fmt.Sprintf("%d", leaders-1))
	runCmd("x")

	WaitMinutes(state0, 2) // make sure they get faulted

	// bring them back
	runCmd(fmt.Sprintf("%d", leaders-2))
	runCmd("x")
	runCmd(fmt.Sprintf("%d", leaders-1))
	runCmd("x")
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)

	// PrintOneStatus(0, 0)
	if GetFnodes()[leaders-2].State.Leader {
		t.Fatalf("Node %d should not be a leader", leaders-2)
	}
	if GetFnodes()[leaders-1].State.Leader {
		t.Fatalf("Node %d should not be a leader", leaders-1)
	}
	if !GetFnodes()[leaders].State.Leader {
		t.Fatalf("Node %d should be a leader", leaders)
	}
	if !GetFnodes()[leaders+1].State.Leader {
		t.Fatalf("Node %d should be a leader", leaders+1)
	}

	CheckAuthoritySet(leaders, audits, t)

	if state0.IsActive(activations.ELECTION_NO_SORT) {
		t.Fatalf("ELECTION_NO_SORT active too early")
	}

	for !state0.IsActive(activations.ELECTION_NO_SORT) {
		WaitBlocks(state0, 1)
	}

	WaitForMinute(state0, 2) // Don't Fault at the end of a block

	// Cause a new double elections by killing the new leaders
	runCmd(fmt.Sprintf("%d", leaders))
	runCmd("x")
	runCmd(fmt.Sprintf("%d", leaders+1))
	runCmd("x")
	WaitMinutes(state0, 2) // make sure they get faulted
	// bring them back
	runCmd(fmt.Sprintf("%d", leaders))
	runCmd("x")
	runCmd(fmt.Sprintf("%d", leaders+1))
	runCmd("x")
	WaitBlocks(state0, 3)
	WaitMinutes(state0, 1)

	if GetFnodes()[leaders].State.Leader {
		t.Fatalf("Node %d should not be a leader", leaders)
	}
	if GetFnodes()[leaders+1].State.Leader {
		t.Fatalf("Node %d should not be a leader", leaders+1)
	}
	if !GetFnodes()[leaders-1].State.Leader {
		t.Fatalf("Node %d should be a leader", leaders-1)
	}
	if !GetFnodes()[leaders-2].State.Leader {
		t.Fatalf("Node %d should be a leader", leaders-2)
	}

	CheckAuthoritySet(leaders, audits, t)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

	// Sleep one block
	time.Sleep(time.Duration(state0.DirectoryBlockInSeconds) * time.Second)
	if state0.LLeaderHeight > 14 {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}

}
func TestAnElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	var (
		leaders   int = 3
		audits    int = 2
		followers int = 1
		nodes     int = leaders + audits + followers
	)

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--net=alot+",
		"--enablenet=false",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		fmt.Sprintf("--count=%d", nodes),
		"--startdelay=1",
		//"--debuglog=F.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		"--checkheads=false",
	)
	params := ParseCmdLine(args)

	time.Sleep(5 * time.Second) // wait till the control panel is setup
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(5 * time.Second) // wait till the simulation is setup

	t.Log(fmt.Sprintf("Allocated %d nodes", nodes))
	fnodes := GetFnodes()
	if len(fnodes) != nodes {
		t.Fatalf("Should have allocated %d nodes", nodes)
		t.Fail()
	}

	StatusEveryMinute(state0)
	WaitMinutes(state0, 2)

	runCmd("g6")
	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

	for {
		pendingCommits := 0
		for _, s := range fnodes {
			pendingCommits += s.State.Commits.Len()
		}
		if pendingCommits == 0 {
			break
		}
		fmt.Printf("Waiting for G5 to complete\n")
		WaitMinutes(state0, 1)

	}
	// Allocate leaders
	runCmd("1")
	for i := 0; i < leaders-1; i++ {
		runCmd("l")
	}

	// Allocate audit servers
	for i := 0; i < audits; i++ {
		runCmd("o")
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 2)
	PrintOneStatus(0, 0)
	runCmd("2")
	runCmd("w") // point the control panel at 2

	CheckAuthoritySet(leaders, audits, t)

	runCmd(fmt.Sprintf("%d", leaders-1))
	runCmd("x")
	WaitBlocks(state0, 3)
	runCmd("x")

	WaitBlocks(state0, 2)
	WaitMinutes(state0, 2)

	// PrintOneStatus(0, 0)
	if GetFnodes()[leaders-1].State.Leader {
		t.Fatalf("Node %d should not be a leader", leaders-1)
	}
	if !GetFnodes()[leaders].State.Leader && !GetFnodes()[leaders+1].State.Leader {
		t.Fatalf("Node %d or %d should be a leader", leaders, leaders+1)
	}

	CheckAuthoritySet(leaders, audits, t)

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

func Test5up(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	var (
		leaders   int = 3
		audits    int = 0
		followers int = 2
		nodes     int = leaders + audits + followers
	)

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		time.Sleep(100 * time.Millisecond)
		return
	}

	args := append([]string{},

		"--network=LOCAL",
		"--net=alot+",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--enablenet=false",
		//"--debugconsole=localhost",
		"--startdelay=5",
		fmt.Sprintf("-count=%d", nodes),
		//"--debuglog=.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
	)
	params := ParseCmdLine(args)

	time.Sleep(5 * time.Second) // wait till the control panel is setup
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(5 * time.Second) // wait till the simulation is setup

	t.Log(fmt.Sprintf("Allocated %d nodes", nodes))
	fnodes := GetFnodes()
	if len(fnodes) != nodes {
		t.Fatalf("Should have allocated %d nodes", nodes)
		t.Fail()
	}

	StatusEveryMinute(state0)
	WaitMinutes(state0, 2)

	runCmd("g6")
	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

	for {
		pendingCommits := 0
		for _, s := range fnodes {
			pendingCommits += s.State.Commits.Len()
		}
		if pendingCommits == 0 {
			break
		}
		fmt.Printf("Waiting for G5 to complete\n")
		WaitMinutes(state0, 1)

	}
	// Allocate leaders
	runCmd("1")
	for i := 0; i < leaders-1; i++ {
		runCmd("l")
	}

	// Allocate audit servers
	for i := 0; i < audits; i++ {
		runCmd("o")
	}

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 2)
	PrintOneStatus(0, 0)
	runCmd("2")
	runCmd("w") // point the control panel at 2

	CheckAuthoritySet(leaders, audits, t)

	runCmd("R10")
	WaitBlocks(state0, 10)
	runCmd("R0")
	WaitMinutes(state0, 2)

	CheckAuthoritySet(leaders, audits, t)

	WaitBlocks(state0, 1)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

	// Sleep one block
	time.Sleep(time.Duration(state0.DirectoryBlockInSeconds) * time.Second)
	if state0.LLeaderHeight > 13 {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}
	j := state0.SyncingStateCurrent
	for range state0.SyncingState {
		fmt.Println(state0.SyncingState[j])
		j = (j - 1 + len(state0.SyncingState)) % len(state0.SyncingState)
	}

}

func TestDBsigEOMElection(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		os.Stdout.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--enablenet=false",
		"--logPort=37000",
		"--port=37001",
		"--controlpanelport=37002",
		"--networkport=37003",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=7",
		"--startdelay=1",
		"--net=alot+",
		"--debuglog=.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		"--checkheads=false",
		//		"-debugconsole=localhost:8093",
	)

	params := ParseCmdLine(args)
	_ = Factomd(params, false).(*state.State)
	time.Sleep(1 * time.Second)

	state := GetFnodes()[2].State
	state.MessageTally = true
	StatusEveryMinute(state)
	t.Log("Allocated 7 nodes")
	if len(GetFnodes()) != 7 {
		t.Fatal("Should have allocated 7 nodes")
		t.Fail()
	}

	WaitForMinute(state, 1)
	runCmd("g7")
	WaitBlocks(state, 1)
	// Allocate 1 leaders
	WaitForMinute(state, 1)

	runCmd("0")
	runCmd("l") // leaders
	runCmd("l") // leaders
	runCmd("l") // leaders
	runCmd("l") // leaders
	runCmd("l") // leaders
	runCmd("o") // Audit
	runCmd("o") // Audit
	runCmd("l") // leaders
	runCmd("l") // leaders

	WaitBlocks(state, 1)
	WaitForMinute(state, 2)

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

	if leadercnt != 5 {
		t.Fatalf("found %d leaders, expected 5", leadercnt)
	}

	PrintOneStatus(2, 2)

	//// Wait for the activation of the
	//for !state.IsActive(activations.ELECTION_NO_SORT) {
	//	WaitBlocks(state, 1)
	//}

	var wait sync.WaitGroup
	wait.Add(2)

	// wait till after EOM 9 but before DBSIG
	stop0 := func() {
		s := GetFnodes()[0].State
		WaitForMinute(state, 9)
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

	PrintOneStatus(2, 2)

	fmt.Println("Caused Elections")

	//runCmd("E")
	//runCmd("F")
	//runCmd("0")
	//runCmd("p")
	WaitBlocks(state, 3)
	// bring them back
	runCmd("0")
	runCmd("x")
	runCmd("1")
	runCmd("x")
	WaitBlocks(state, 2)

	leadercnt = 0
	auditcnt = 0
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
	if leadercnt != 5 {
		t.Fatalf("found %d leaders, expected 5", leadercnt)
	}
	if auditcnt != 2 {
		t.Fatalf("found %d leaders, expected 2", auditcnt)
	}

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

}
func TestMultiple2Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		os.Stdout.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--enablenet=false",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=10",
		"--startdelay=1",
		"--net=alot+",
		//"--debuglog=F.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		//"--debugconsole=localhost:8093",
		"--checkheads=false",
	)

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(3 * time.Second)
	StatusEveryMinute(state0)
	t.Log("Allocated 10 nodes")
	if len(GetFnodes()) != 10 {
		t.Fatal("Should have allocated 10 nodes")
		t.Fail()
	}

	WaitForMinute(state0, 3)
	runCmd("g15")
	WaitBlocks(state0, 1)
	// Allocate 1 leaders
	WaitForMinute(state0, 1)

	runCmd("1")              // select node 1
	for i := 0; i < 6; i++ { // 1, 2, 3, 4, 5, 6
		runCmd("l") // leaders
	}

	for i := 0; i < 2; i++ { // 8, 9
		runCmd("o") // leaders
	}

	WaitBlocks(state0, 1)
	WaitForMinute(state0, 2)

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

	if leadercnt != 7 {
		t.Fatalf("found %d leaders, expected 7", leadercnt)
	}

	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")

	runCmd("s")
	runCmd("E")
	runCmd("F")
	runCmd("0")
	runCmd("p")
	WaitBlocks(state0, 3)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

}

func TestMultiple3Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		os.Stdout.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--enablenet=false",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=12",
		"--startdelay=1",
		"--net=alot+",
		//"--debuglog=F.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		//"--debugconsole=localhost:8093",
		"--checkheads=false",
	)

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(3 * time.Second)
	StatusEveryMinute(state0)
	t.Log("Allocated 10 nodes")
	if len(GetFnodes()) != 12 {
		t.Fatal("Should have allocated 10 nodes")
		t.Fail()
	}

	WaitForMinute(state0, 3)
	runCmd("g15")
	WaitBlocks(state0, 1)
	// Allocate 1 leaders
	WaitForMinute(state0, 1)

	runCmd("1")              // select node 1
	for i := 0; i < 6; i++ { // 1, 2, 3, 4, 5, 6
		runCmd("l") // leaders
	}

	for i := 0; i < 4; i++ { // 7, 8, 9, 10
		runCmd("o") // audits
	}

	WaitBlocks(state0, 1)
	WaitForMinute(state0, 2)

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

	if leadercnt != 7 {
		t.Fatalf("found %d leaders, expected 7", leadercnt)
	}

	//runCmd("s")
	//runCmd("E")
	//runCmd("F")
	runCmd("0")

	runCmd("1")
	runCmd("x")
	runCmd("2")
	runCmd("x")
	runCmd("3")
	runCmd("x")
	runCmd("0")
	WaitMinutes(state0, 1)
	//runCmd("3")
	//runCmd("x")
	WaitBlocks(state0, 3)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

}

func TestMultiple7Election(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		os.Stdout.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--enablenet=false",
		"--blktime=8",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=25",
		"--startdelay=1",
		"--net=alot+",
		//"--debuglog=F.*",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		//"--debugconsole=localhost:8093",
		"--checkheads=false",
	)

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(3 * time.Second)
	StatusEveryMinute(state0)
	t.Log("Allocated 25 nodes")
	if len(GetFnodes()) != 25 {
		t.Fatal("Should have allocated 25 nodes")
		t.Fail()
	}

	WaitForMinute(state0, 3)
	runCmd("g30")
	WaitBlocks(state0, 1)
	// Allocate 1 leaders
	WaitForMinute(state0, 1)
	runCmd("1")               // select node 1
	for i := 0; i < 14; i++ { // 1, 2, 3, 4, 5, 6
		time.Sleep(100 * time.Millisecond)
		runCmd("l") // leaders
	}

	for i := 0; i < 10; i++ { // 8, 9
		time.Sleep(100 * time.Millisecond)
		runCmd("o") // leaders
	}

	WaitBlocks(state0, 1)
	WaitForMinute(state0, 2)

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

	if leadercnt != 15 {
		t.Fatalf("found %d leaders, expected 15", leadercnt)
	}

	if auditcnt != 10 {
		t.Fatalf("found %d audits, expected 10", auditcnt)
	}

	for i := 1; i < 3; i++ {
		runCmd(fmt.Sprintf("%d", i))
		runCmd("x")
	}

	WaitBlocks(state0, 3)

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
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

func makeExpected(grants []state.HardGrant) []interfaces.ITransAddress {
	var rval []interfaces.ITransAddress
	for _, g := range grants {
		rval = append(rval, factoid.NewOutAddress(g.Address, g.Amount))
	}
	return rval
}

func TestGrants(t *testing.T) {
	if ranSimTest {
		return
	}

	ranSimTest = true

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
		return
	}

	args := append([]string{},
		"--db=Map",
		"--network=LOCAL",
		"--enablenet=false",
		"--blktime=1",
		"--faulttimeout=2",
		"--roundtimeout=2",
		"--count=3",
		"--startdelay=1",
		"--net=alot+",
		"--debuglog=database.txt",
		"--stdoutlog=out.txt",
		"--stderrlog=err.txt",
		//"--debugconsole=localhost:8093",
		"--checkheads=false",
	)

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	WaitForMinute(state0, 3)
	runCmd("g2")
	WaitBlocks(state0, 1)
	// Allocate 1 audit "LAF"
	WaitForMinute(state0, 1)
	runCmd("1") // select node 1
	runCmd("o") // audit

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

		userAddr := primitives.ConvertFctAddressToUserStr(g.Address)
		_, ok := grantBalances[userAddr]
		if !ok {
			grantBalances[userAddr] = state0.FactoidState.GetFactoidBalance(g.Address.Fixed()) // Save initial balance
		}
		grantBalances[userAddr] += int64(g.Amount) // Add the grant amount
	}

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
		}
		for i, _ := range expected {
			if !expected[i].GetAddress().IsSameAs(gotGrants[i].GetAddress()) ||
				expected[i].GetAmount() != gotGrants[i].GetAmount() ||
				expected[i].GetUserAddress() != gotGrants[i].GetUserAddress() {
				t.Errorf("Expected: %v ", expected[i])
				t.Errorf("but found %v for grant #%d at %d", gotGrants[i], i, dbheight)
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
}

func printList(title string, list map[string]uint64) {
	for addr, amt := range list {
		fmt.Printf("%v - %v:%v\n", title, addr, amt)
	}
}
