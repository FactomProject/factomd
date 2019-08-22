package simtest

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/util/atomic"

	"github.com/FactomProject/factomd/activations"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/wsapi"
)

func TestAnElection(t *testing.T) {
	if RanSimTest {
		return
	}

	RanSimTest = true

	state0 := SetupSim("LLLAAF", map[string]string{"--blktime": "15"}, 9, 1, 1, t)

	StatusEveryMinute(state0)
	WaitMinutes(state0, 2)

	RunCmd("2")
	RunCmd("w") // point the control panel at 2

	// remove the last leader
	RunCmd("2")
	RunCmd("x")
	// wait for the election
	WaitMinutes(state0, 2)
	//bring him back
	RunCmd("x")

	// wait for him to update via dbstate and become an audit
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)

	// PrintOneStatus(0, 0)
	if GetFnodes()[2].State.Leader {
		t.Fatalf("Node 2 should not be a leader")
	}
	if !GetFnodes()[3].State.Leader && !GetFnodes()[4].State.Leader {
		t.Fatalf("Node 3 or 4  should be a leader")
	}

	WaitForAllNodes(state0)
	ShutDownEverything(t)

}
