package simtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"time"
)

import (
	"testing"
)

func TestSetupANetwork(t *testing.T) {

	state0 := simulation.SetupSim("LLLLAAAFFF", map[string]string{"--debuglog": ""}, 20, 0, 0, t)

	simulation.RunCmd("9")  // Puts the focus on node 9
	simulation.RunCmd("x")  // Takes Node 9 Offline
	simulation.RunCmd("w")  // Point the WSAPI to send API calls to the current node.
	simulation.RunCmd("10") // Puts the focus on node 9
	simulation.RunCmd("8")  // Puts the focus on node 8
	simulation.RunCmd("w")  // Point the WSAPI to send API calls to the current node.
	simulation.RunCmd("7")
	simulation.WaitBlocks(state0, 1) // Wait for 1 block

	simulation.WaitForMinute(state0, 2) // Waits for minute 2
	simulation.RunCmd("F100")           //  Set the Delay on messages from all nodes to 100 milliseconds
	// .15 second minutes is too fast for dropping messages until the dropping is fixed (FD-971) is fixed
	// could change to 4 second minutes and turn this back on -- Clay
	//	RunCmd("S10")            // Set Drop Rate to 1.0 on everyone
	simulation.RunCmd("g10") // Adds 10 identities to your identity pool.

	fn1 := simulation.GetFocus()
	simulation.PrintOneStatus(0, 0)
	if fn1.State.FactomNodeName != "FNode07" {
		t.Fatalf("Expected FNode07, but got %s", fn1.State.FactomNodeName)
	}
	simulation.RunCmd("g1")             // Adds 1 identities to your identity pool.
	simulation.WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	simulation.RunCmd("g1")             // // Adds 1 identities to your identity pool.
	simulation.WaitForMinute(state0, 4) // Waits for 4 "Minutes"
	simulation.RunCmd("g1")             // Adds 1 identities to your identity pool.
	simulation.WaitForMinute(state0, 5) // Waits for 5 "Minutes"
	simulation.RunCmd("g1")             // Adds 1 identities to your identity pool.
	simulation.WaitForMinute(state0, 6) // Waits for 6 "Minutes"
	simulation.WaitBlocks(state0, 1)    // Waits for 1 block
	simulation.WaitForMinute(state0, 1) // Waits for 1 "Minutes"
	simulation.RunCmd("g1")             // Adds 1 identities to your identity pool.
	simulation.WaitForMinute(state0, 2) // Waits for 2 "Minutes"
	simulation.RunCmd("g1")             // Adds 1 identities to your identity pool.
	simulation.WaitForMinute(state0, 3) // Waits for 3 "Minutes"
	simulation.RunCmd("g20")            // Adds 20 identities to your identity pool.
	simulation.WaitBlocks(state0, 1)
	simulation.RunCmd("9") // Focuses on Node 9
	simulation.RunCmd("x") // Brings Node 9 back Online
	simulation.RunCmd("8") // Focuses on Node 8

	time.Sleep(100 * time.Millisecond)

	fn2 := simulation.GetFocus()
	simulation.PrintOneStatus(0, 0)
	if fn2.State.FactomNodeName != "FNode08" {
		t.Fatalf("Expected FNode08, but got %s", fn1.State.FactomNodeName)
	}

	simulation.RunCmd("i") // Shows the identities being monitored for change.
	// Test block recording lengths and error checking for pprof
	simulation.RunCmd("b100") // Recording delays due to blocked go routines longer than 100 ns (0 ms)

	simulation.RunCmd("b") // specifically how long a block will be recorded (in nanoseconds).  1 records all blocks.

	simulation.RunCmd("babc") // Not sure that this does anything besides return a message to use "bnnn"

	simulation.RunCmd("b1000000") // Recording delays due to blocked go routines longer than 1000000 ns (1 ms)

	simulation.RunCmd("/") // Sort Status by Chain IDs

	simulation.RunCmd("/") // Sort Status by Node Name

	simulation.RunCmd("a1")             // Shows Admin block for Node 1
	simulation.RunCmd("e1")             // Shows Entry credit block for Node 1
	simulation.RunCmd("d1")             // Shows Directory block
	simulation.RunCmd("f1")             // Shows Factoid block for Node 1
	simulation.RunCmd("a100")           // Shows Admin block for Node 100
	simulation.RunCmd("e100")           // Shows Entry credit block for Node 100
	simulation.RunCmd("d100")           // Shows Directory block
	simulation.RunCmd("f100")           // Shows Factoid block for Node 1
	simulation.RunCmd("yh")             // Nothing
	simulation.RunCmd("yc")             // Nothing
	simulation.RunCmd("r")              // Rotate the WSAPI around the nodes
	simulation.WaitForMinute(state0, 1) // Waits 1 "Minute"

	simulation.RunCmd("g1")             // Adds 1 identities to your identity pool.
	simulation.WaitForMinute(state0, 3) // Waits 3 "Minutes"
	simulation.WaitBlocks(fn1.State, 3) // Waits for 3 blocks

	simulation.ShutDownEverything(t)

}
