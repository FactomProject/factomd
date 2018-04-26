package messages

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
)

//TODO: Cache message hash to message string with age out...
var (
	traceMutex sync.Mutex
	files      map[string]*os.File
	enabled    map[string]bool
	TestRegex  *regexp.Regexp
	sequence   int
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
	if TestRegex == nil {
		if globals.Params.DebugLogRegEx[0] == '"' || globals.Params.DebugLogRegEx[0] == '\'' {
			globals.Params.DebugLogRegEx = globals.Params.DebugLogRegEx[1 : len(globals.Params.DebugLogRegEx)-1] // Trim the leading "
		}
		theRegex, err := regexp.Compile(globals.Params.DebugLogRegEx)
		if err != nil {
			panic(err)
		}
		files = make(map[string]*os.File)
		enabled = make(map[string]bool)
		TestRegex = theRegex
	}

	flag, old := enabled[name]
	if !old {
		flag = TestRegex.Match([]byte(name))
		enabled[name] = flag
	}
	return flag
}

// assumes traceMutex is locked already
func getTraceFile(name string) (f *os.File) {
	//traceMutex.Lock()	defer traceMutex.Unlock()
	if !checkFileName(name) {
		return nil
	}
	if files == nil {
		files = make(map[string]*os.File)
	}
	f, _ = files[name]
	if f == nil {
		fmt.Println("Creating " + name)
		var err error
		f, err = os.Create(name)
		if err != nil {
			panic(err)
		}
		files[name] = f
	}
	return f
}

var history [16384][32]byte // Last 16k messages logged
var h int                   // head of history
var msgmap map[[32]byte]string = make(map[[32]byte]string)

func addmsg(hash [32]byte, msg string) {
	remove := history[h] // get the oldest message
	delete(msgmap, remove)
	history[h] = hash
	msgmap[hash] = msg
	h = (h + 1) % cap(history) // move the head
}

func getmsg(hash [32]byte) string {
	rval, ok := msgmap[hash]
	if !ok {
		rval = fmt.Sprintf("UnknownMsg: %x", hash[:3])
	}
	return rval
}
func LogMessage(name string, note string, msg interfaces.IMsg) {

	traceMutex.Lock()
	defer traceMutex.Unlock()
	myfile := getTraceFile(name)
	if myfile == nil {
		return
	}
	sequence++
	seq := sequence
	var t byte
	var rhash, hash, msgString string
	embeddedHash := ""
	hash = "??????"
	rhash = "??????"
	if msg == nil {
		t = 0
		msgString = "-nil-"
	} else {
		t = msg.Type()
		msgString = msg.String()
		// work around message that don't have hashes yet ...
		h := msg.GetMsgHash()
		if h != nil {
			hash = h.String()[:6]
		}
		h = msg.GetRepeatHash()
		if h != nil {
			rhash = h.String()[:6]
		}

		switch t {
		case constants.ACK_MSG:
			ack := msg.(*Ack)
			byte := ack.GetHash().Fixed
			embeddedHash = fmt.Sprintf(" EmbeddedMsg: %s", getmsg(byte()))
		case constants.MISSING_MSG_RESPONSE:
			mm := msg.(*MissingMsgResponse)
			embeddedHash = fmt.Sprintf(" EmbeddedMsg: %s | %s", mm.MsgResponse.String(), mm.AckResponse.String())
		case constants.MISSING_DATA:
			md := msg.(*MissingData)
			embeddedHash = fmt.Sprintf(" EmbeddedMsg: %s", getmsg(md.RequestHash.Fixed()))

		default:
			if msg.GetMsgHash() != nil {
				bytes := msg.GetMsgHash().Fixed()
				addmsg(bytes, msgString) // Keep message we have seen for a while
			}
		}
	}

	now := time.Now().Local()

	s := fmt.Sprintf("%7v %02d:%02d:%02d %-25s M-%v|R-%v %26s[%2v]:%v%v\n", seq, now.Hour()%24, now.Minute()%60, now.Second()%60,
		note, hash, rhash, constants.MessageName(byte(t)), t,
		msgString, embeddedHash)
	s = addNodeNames(s)

	myfile.WriteString(s)
}

var findHex *regexp.Regexp

// Look up the hex string in the map of names...
func LookupName(s string) string {
	n, ok := globals.FnodeNames[s]
	if ok {
		return "<" + n + ">"
	}
	return ""
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

func LogPrintf(name string, format string, more ...interface{}) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	myfile := getTraceFile(name)
	if myfile == nil {
		return
	}
	seq := sequence
	now := time.Now().Local()
	s := fmt.Sprintf("%7v %02d:%02d:%02d %s\n", seq, now.Hour()%24, now.Minute()%60, now.Second()%60, fmt.Sprintf(format, more...))
	s = addNodeNames(s)
	myfile.WriteString(s)
}

// stringify it in the caller to avoid having to deal with the import loop
func LogParcel(name string, note string, msg string) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	myfile := getTraceFile(name)
	if myfile == nil {
		return
	}
	sequence++
	seq := sequence

	myfile.WriteString(fmt.Sprintf("%5v %26s %s\n", seq, note, msg))
}

// Log a message with a state timestamp
func StateLogMessage(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, comment string, msg interfaces.IMsg) {
	logFileName := FactomNodeName + "_" + logName + ".txt"
	t := fmt.Sprintf("%d-:-%d ", DBHeight, CurrentMinute)
	LogMessage(logFileName, t+comment, msg)
}

// Log a printf with a state timestamp
func StateLogPrintf(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, format string, more ...interface{}) {
	logFileName := FactomNodeName + "_" + logName + ".txt"
	t := fmt.Sprintf("%d-:-%d ", DBHeight, CurrentMinute)
	LogPrintf(logFileName, t+format, more...)
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
