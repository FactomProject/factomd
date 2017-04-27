package engine_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
)

var _ = Factomd

// Wait so many blocks
func waitBlocks(s *state.State, blks int) {
	currentHeight := int(s.LLeaderHeight)
	for int(s.LLeaderHeight) < currentHeight+blks {
		time.Sleep(time.Second)
	}
}

// Wait to a given minute.  If we are == to the minute or greater, then
// we first wait to the start of the next block.
func waitMinutes(s *state.State, min int) {
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
		v := endCap()
		return v
	}

	args := append([]string{},
		"-db=Map",
		"-network=LOCAL",
		"-enablenet=true",
		"-blktime=15",
		"-count=10",
		"-logPort=37000",
		"-port=37001",
		"-ControlPanelPort=37002",
		"-networkPort=37003")

	params := ParseCmdLine(args)
	go Factomd(params, false)
	time.Sleep(3 * time.Second)

	t.Log("Allocated 10 nodes")
	if len(GetFnodes()) != 10 {
		t.Fatal("Should have allocated 10 nodes")
		t.Fail()
	}
	n0 := GetFnodes()[0]

	runCmd("s")
	runCmd("9")
	runCmd("x")
	runCmd("8")
	runCmd("")
	waitBlocks(n0.State, 1)
	runCmd("g10")
	waitBlocks(n0.State, 1)
	// Allocate 4 leaders

	waitMinutes(n0.State, 1)

	runCmd("l")
	runCmd("")
	runCmd("")
	runCmd("")

	// Allocate 3 audit servers
	runCmd("o")
	runCmd("")
	runCmd("")

	waitBlocks(n0.State, 1)
	waitMinutes(n0.State, 1)

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

	fn1 := GetFocus()
	if fn1.State.FactomNodeName != "FNode07" {
		t.Fatalf("Expected FNode07, but got %s", fn1.State.FactomNodeName)
	}

	runCmd("8")

	time.Sleep(100 * time.Millisecond)

	fn2 := GetFocus()
	if fn2.State.FactomNodeName != "FNode08" {
		t.Fatalf("Expected FNode08, but got %s", fn1.State.FactomNodeName)
	}

	// Test block recording lengths and error checking for pprof
	runCmd("b100")

	runCmd("b")

	runCmd("babc")

	runCmd("b1000000")

	runCmd("/")

	runCmd("w")

	runCmd("g101")

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
	time.Sleep(5 * time.Second)
	runCmd("r")
	runCmd("9")
	runCmd("x")
	waitBlocks(fn1.State, 3)

	runCmd("T10")
	t.Log("Run to a dbht of 10")
	n0.State.DirectoryBlockInSeconds = 4
	for n0.State.LLeaderHeight < 8 {
		time.Sleep(time.Second)
	}
	for n0.State.CurrentMinute < 1 {
		time.Sleep(time.Second)
	}

	t.Log("Shutting down the network")
	for _, fn := range GetFnodes() {
		fn.State.ShutdownChan <- 1
	}

	time.Sleep(15 * time.Second)
	if n0.State.LLeaderHeight > 10 {
		t.Fatal("Failed to shut down factomd via ShutdownChan")
	}
}
