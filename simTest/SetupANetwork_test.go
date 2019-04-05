package simtest

import "time"

import (
	"testing"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestSetupANetwork(t *testing.T) {

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
