package longtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestLeaderModule(t *testing.T) {
	// watch logs for leader and networkouput for filtered messages
	params := map[string]string{"--debuglog": "."}
	state0 := SetupSim("LF", params, 7, 0, 0, t)

	RunCmd("R1")
	WaitForBlock(state0, 4)
}
