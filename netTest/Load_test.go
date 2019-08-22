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

func TestLoad(t *testing.T) {
	// use a tree so the messages get reordered
	//state0 := SetupSim("LLF", map[string]string{}, 15, 0, 0, t)

	RunCmd("2")   // select 2
	RunCmd("R30") // Feed load
	WaitBlocks(state0, 10)
	RunCmd("R0") // Stop load
	WaitBlocks(state0, 1)
	ShutDownEverything(t)
}
