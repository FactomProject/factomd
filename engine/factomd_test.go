package engine_test

import (
	"os"
	"testing"
	"time"

	"fmt"
	"runtime"

	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	"github.com/pkg/profile"
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

	start := time.Now()
	defer func() { fmt.Printf("Execution took:%v\n", time.Since(start)) }()

	runtime.SetMutexProfileFraction(5)
	path := os.Getenv("HOME") + "/go/src/github.com/FactomProject/factomd"
	fmt.Println("Profile path:" + path)
	defer profile.Start(profile.MutexProfile, profile.ProfilePath(path)).Stop()

	runCmd := func(cmd string) {
		os.Stderr.WriteString("Executing: " + cmd + "\n")
		InputChan <- cmd
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
		"faulttimeout=15")

	params := ParseCmdLine(args)
	state0 := Factomd(params, false).(*state.State)
	state0.MessageTally = true
	time.Sleep(30 * time.Second)

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
	WaitMinutes(state0, 3)

	runCmd("1")
	runCmd("l")
	runCmd("")
	runCmd("")

	// Allocate 3 audit servers
	runCmd("o")
	runCmd("")
	runCmd("")

	WaitBlocks(state0, 1)
	WaitMinutes(state0, 1)

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
	WaitMinutes(state0, 2)
	runCmd("F100")
	runCmd("S10")
	runCmd("g10")

	fn1 := GetFocus()
	if fn1.State.FactomNodeName != "FNode07" {
		t.Fatalf("Expected FNode07, but got %s", fn1.State.FactomNodeName)
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
