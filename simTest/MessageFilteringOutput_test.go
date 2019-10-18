package simtest

import (
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestFilterAPIOutput(t *testing.T) {

	state0 := SetupSim("LLLLLAAF", map[string]string{"--debuglog": "networkoutputs"}, 25, 1, 1, t)

	RunCmd("1")
	RunCmd("w")
	RunCmd("s")

	apiRegex := "EOM.*5.*minute +1" // It has two spaces.
	SetOutputFilter(apiRegex)

	WaitBlocks(state0, 5)

	// The message-filter call we did above should have caused an election and SO, Node01 should not be a leader anymore.
	if GetFnodes()[1].State.Leader {
		t.Fatalf("Node01 should not be leader!")
	}
	CheckAuthoritySet(t)

	// Check Node01 Network Output logs to make sure there are no Dropped messaged besides the ones for our Regex
	out := SystemCall(`grep "Drop, matched filter Regex" fnode01_networkoutputs.txt | grep -Ev "` + apiRegex + `" | wc -l`)

	if strings.TrimSuffix(strings.Trim(string(out), " "), "\n") != string("0") {
		t.Fatalf("Filter missed let a message pass 1.")
	}

	// Checks Node01 Network Outputs to make sure there are no Sent broadcast including our Regex
	out2 := SystemCall(`grep "Send broadcast" fnode01_networkoutputs.txt | grep "` + apiRegex + `" | grep -v "EmbeddedMsg" | wc -l`)

	if strings.TrimSuffix(strings.Trim(string(out2), " "), "\n") != string("0") {
		t.Fatalf("Filter missed let a message pass 2.")
	}

	ShutDownEverything(t)
}
