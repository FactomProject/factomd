package globals

import (
	"io"
	"sync"
	"time"
)

var FnodeNames map[string]string = make(map[string]string) /// use by MessageTrace debug code
var Params *FactomParams
var StartTime time.Time
var LastDebugLogRegEx string      // used to detect if FactomParams.DebugLogRegEx was changed by the control panel
var InputChan = make(chan string) // Get commands here

/****************************************************************
	DEBUG logging to keep full hash. Turned on from command line
 ****************************************************************/
var HashMutex sync.Mutex
var Hashlog io.Writer
var Hashes map[[32]byte]bool
var HashesInOrder [10000]*[32]byte
var HashNext int

func init() {
	Params = new(FactomParams)
}
