package log

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
)

var (
	traceMutex sync.Mutex
	files      map[string]*os.File
	enabled    map[string]bool
	TestRegex  *regexp.Regexp
	sequence   int
	history    *([16384][32]byte) // Last 16k messages logged
	h          int                // head of history
	msgmap     map[[32]byte]interfaces.IMsg
)

// Check a filename and see if logging is on for that filename
// If it never ben see then check with the regex. If it has been seen then just look it up in the map
// assumes traceMutex is locked already
func CheckFileName(name string) bool {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	return checkFileName(name)
}

func checkFileName(name string) bool {
	if globals.Params.DebugLogRegEx == "" {
		return false
	}
	checkForChangesInDebugRegex()
	flag, old := enabled[name]
	if !old {
		flag = TestRegex.Match([]byte(name))
		enabled[name] = flag
	}
	return flag
}

func checkForChangesInDebugRegex() {
	// if  the regex string has changed ...
	if globals.Params.DebugLogRegEx != globals.LastDebugLogRegEx {
		globals.Params.DebugLogLocation, globals.Params.DebugLogRegEx = SplitUpDebugLogRegEx(globals.Params.DebugLogRegEx)

		TestRegex = nil // throw away the old regex
		globals.LastDebugLogRegEx = globals.Params.DebugLogRegEx
	}
	//strip quotes if they are included in the string
	if globals.Params.DebugLogRegEx != "" && (globals.Params.DebugLogRegEx[0] == '"' || globals.Params.DebugLogRegEx[0] == '\'') {
		globals.Params.DebugLogRegEx = globals.Params.DebugLogRegEx[1 : len(globals.Params.DebugLogRegEx)-1] // Trim the "'s
	}
	// if we haven't compiled the regex ...
	if TestRegex == nil && globals.Params.DebugLogRegEx != "" {
		theRegex, err := regexp.Compile("(?i)" + globals.Params.DebugLogRegEx) // force case insensitive
		if err != nil {
			panic(err)
		}
		enabled = make(map[string]bool) // create a clean cache of enabled files
		TestRegex = theRegex
	}
	globals.LastDebugLogRegEx = globals.Params.DebugLogRegEx
}

func SplitUpDebugLogRegEx(DebugLogRegEx string) (string, string) {
	lastSlashIndex := strings.LastIndex(DebugLogRegEx, string(os.PathSeparator))
	regex := DebugLogRegEx[lastSlashIndex+1:]
	dirlocation := DebugLogRegEx[0 : lastSlashIndex+1]
	return dirlocation, regex
}

// assumes traceMutex is locked already
func getTraceFile(name string) (f *os.File) {
	checkForChangesInDebugRegex()
	//traceMutex.Lock()	defer traceMutex.Unlock()
	name = globals.Params.DebugLogLocation + strings.ToLower(name)
	if !checkFileName(name) {
		return nil
	}
	if files == nil {
		files = make(map[string]*os.File)
	}
	f, _ = files[name]
	if f != nil {
		_, err := os.Stat(name)
		if os.IsNotExist(err) {
			// The file was deleted out from under us
			f.Close() // close the old log
			f = nil   // make the code reopen the file
		}
	}
	if f == nil {
		fmt.Println("Creating " + (name))
		var err error
		f, err = os.Create(name)
		if err != nil {
			panic(err)
		}
		files[name] = f
		f.WriteString(time.Now().String() + "\n")
	}
	return f
}

func addmsg(msg interfaces.IMsg) {
	mh := msg.GetMsgHash()
	if mh == nil || reflect.ValueOf(mh).IsNil() {
		return
	}

	hash := mh.Fixed()
	if history == nil {
		history = new([16384][32]byte)
	}
	if msgmap == nil {
		msgmap = make(map[[32]byte]interfaces.IMsg)
	}
	remove := history[h] // get the oldest message
	delete(msgmap, remove)
	history[h] = hash
	msgmap[hash] = msg
	h = (h + 1) % cap(history) // move the head
}

func getmsg(hash [32]byte) interfaces.IMsg {
	if msgmap == nil {
		msgmap = make(map[[32]byte]interfaces.IMsg)
	}
	rval, _ := msgmap[hash]
	return rval
}

var logWhere bool = false // log GoID() of the caller.
var findHex *regexp.Regexp

// Look up the hex string in the map of names...
func LookupName(s string) string {
	n, ok := globals.FnodeNames[s]
	if ok {
		return "<" + n + ">"
	}
	return ""
}

// unused -- of.File is written by direct calls to write and not buffered and the os closes the files on exit.
func Cleanup() {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	for name, f := range files {
		delete(files, name)
		f.Close()
	}
}

// Look thru a string and annotate the string with names based on any hex strings
func addNodeNames(s string) (rval string) {
	var err error
	if findHex == nil {
		findHex, err = regexp.Compile("[A-Fa-f0-9]{6,}")
		if err != nil {
			panic(err)
		}
	}
	hexStr := findHex.FindAllStringIndex(s, -1)
	if hexStr == nil {
		return s
	}
	// loop thru the matches last to first and add the name. Start with the last so the positions don't change
	// as we add text
	for i := len(hexStr); i > 0; {
		i--
		// if it's a short hex string
		if hexStr[i][1]-hexStr[i][0] != 64 {
			l := s[hexStr[i][0]:hexStr[i][1]]
			s = s[:hexStr[i][1]] + LookupName(l) + s[hexStr[i][1]:]
		} else {
			// Shorten 32 byte IDs to [3:6]
			l := s[hexStr[i][0]:hexStr[i][1]]
			name := LookupName(l)
			if name != "" {
				s = s[:hexStr[i][0]] + s[hexStr[i][0]+6:hexStr[i][0]+12] + name + s[hexStr[i][1]:]
			}
		}
	}
	return s
}
