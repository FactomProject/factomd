package engine_test

import (
	"testing"

	"flag"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	"time"
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
			time.Sleep(100*time.Millisecond)
		}
	}

	for min > s.CurrentMinute {
		time.Sleep(100*time.Millisecond)
	}
}


func TestFactomdMain(t *testing.T) {
	{
		var svar string
		flag.StringVar(&svar, "svar", "bo", "a string var")
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

	go Factomd(args, false)
	time.Sleep(3 * time.Second)

	t.Log("Allocated 10 nodes")
	if len(GetFnodes()) != 10 {
		t.Log("Should have allocated 10 nodes")
		t.Fail()
	}
	n0 := GetFnodes()[0]

	InputChan <- "s"

	waitBlocks(n0.State, 1)
	InputChan <- "g10"
	waitBlocks(n0.State, 1)
	// Allocate 4 leaders

	waitMinutes(n0.State, 1)

	InputChan <- "l"
	time.Sleep(100 * time.Millisecond)
	InputChan <- "l"
	time.Sleep(100 * time.Millisecond)
	InputChan <- "l"
	time.Sleep(100 * time.Millisecond)
	InputChan <- "l"
	time.Sleep(100 * time.Millisecond)

	// Allocate 3 audit servers
	InputChan <- "o"
	time.Sleep(100 * time.Millisecond)
	InputChan <- "o"
	time.Sleep(100 * time.Millisecond)
	InputChan <- "o"
	time.Sleep(100 * time.Millisecond)

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
		t.Logf("found %d ", leadercnt)
		t.Fail()
	}

	if auditcnt != 3 {
		t.Logf("found %d ", auditcnt)
		t.Fail()
	}

	t.Log("Run to a dbht of 10")
	n0.State.DirectoryBlockInSeconds = 4
	for n0.State.LLeaderHeight < 10 {
		time.Sleep(time.Second)
	}
	for n0.State.CurrentMinute < 1 {
		time.Sleep(time.Second)
	}
	t.Log("Shutting down Node 0")
	n0.State.ShutdownChan <- 1
	time.Sleep(15 * time.Second)
	if n0.State.LLeaderHeight > 10 {
		t.Log("Failed to shut down factomd via ShutdownChan")
		t.Fail()
	}
}
