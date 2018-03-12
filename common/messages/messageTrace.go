package messages

import (
	"fmt"
	"os"
	"regexp"
	"sync"

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
func checkFileName(name string) bool {
	if TestRegex == nil {

		theRegex, err := regexp.Compile(globals.DebugLogRegEx)
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
	if msg == nil {
		t = 0
		hash = "????????"
		rhash = "????????"
		msgString = "-nil-"
	} else {
		t = msg.Type()
		msgString = msg.String()
		// work around message that don't have hashes yet ...
		h := msg.GetMsgHash()
		if h == nil {
			hash = "????????"
		} else {
			hash = h.String()[:8]
		}
		h = msg.GetRepeatHash()
		if h == nil {
			rhash = "????????"
		} else {
			rhash = h.String()[:8]
		}
	}
	embeddedHash := ""

	switch t {
	case constants.ACK_MSG:
		m := msg.(*Ack)
		embeddedHash = fmt.Sprintf(" EmbeddedMsg %26v[%2v]:%v", constants.MessageName(m.Type()), m.Type(), m.MessageHash.String()[:8])
	case constants.MISSING_MSG_RESPONSE:
		m := msg.(*MissingMsgResponse).MsgResponse
		embeddedHash = fmt.Sprintf(" EmbeddedMsg %26v[%2v]:%v", constants.MessageName(m.Type()), m.Type(), m.GetHash().String()[:8])
	}

	s := fmt.Sprintf("%5v %20s M-%v|R-%v %26s[%2v]:%v {%v}\n", seq, note, hash, rhash, constants.MessageName(byte(t)), t,
		embeddedHash, msgString)
	s = addNodeNames(s)

	myfile.WriteString(s)
}

var findHex *regexp.Regexp

// Look up the hex string in the map of names...
func lookupName(s string) string {
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
	for i := len(hexStr); i > 0; {
		i--
		l := s[hexStr[i][0]:hexStr[i][1]]
		s = s[:hexStr[i][1]] + lookupName(l) + s[hexStr[i][1]:]
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
	s := fmt.Sprintf("%5v %s\n", seq, fmt.Sprintf(format, more...))
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

// unused -- of.File is written by direct calls to write and not buffered and the os closes the files on exit.
func Cleanup() {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	for name, f := range files {
		delete(files, name)
		f.Close()
	}
}
