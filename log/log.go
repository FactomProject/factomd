package log

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"

	"github.com/FactomProject/factomd/common/constants"

	. "github.com/FactomProject/factomd/modules/logging"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
)

var (
	// Create a global logger that adds sequence numbers and timestamps
	GlobalLogger Ilogger

	// Keep a history of logged messages
	msghistlock sync.Mutex
	history     *([16384][32]byte) // Last 16k messages logged
	h           int                // head of history
	msgmap      map[[32]byte]interfaces.IMsg
)

func init() {
	// Create a global FileLogger that assigned the filenames based on thread and logname

	// Create a global logger that adds sequence numbers and timestamps
	GlobalLogger = NewSequenceLogger(NewFileLogger("./."))
	GlobalLogger.AddNameField("fnode", Formatter("%s"), "")
	GlobalLogger.AddNameField("logname", Formatter("%s.txt"), "unknown-log")
	//Add the default print fields comment then message
	GlobalLogger.AddPrintField("dbht", Formatter("%12s"), "")
	GlobalLogger.AddPrintField("comment", Formatter("[%-45v]"), "")
	GlobalLogger.AddPrintField("hash", Formatter("%-38v"), "")
	GlobalLogger.AddPrintField("message", MsgFormatter, "")
}

func LogPrintf(name string, format string, more ...interface{}) {

	/* todo: add multiline support
	lines := strings.Split(fmt.Sprintf(format, more...), "\n")
	now := time.Now().Local()
	for i, text := range lines {
		var s string
		switch i {
		case 0:
			s = fmt.Sprintf("%9d %02d:%02d:%02d.%03d %s %s\n", sequence, now.Hour()%24, now.Minute()%60, now.Second()%60, (now.Nanosecond()/1e6)%1000, text, where)
		default:
			s = fmt.Sprintf("%9d %02d:%02d:%02d.%03d %s\n", sequence, now.Hour()%24, now.Minute()%60, now.Second()%60, (now.Nanosecond()/1e6)%1000, text)
		}
		s = addNodeNames(s)
		myfile.WriteString(s)
	}
	*/

	GlobalLogger.Log(LogData{"logname": name, "comment": Delay_formater(format, more...)})
}

func LogMessage(name string, note string, msg interfaces.IMsg) {

	if msg == nil {
		// include nonempty hash to get spacing right
		GlobalLogger.Log(LogData{"logname": name, "comment": note, "message": "-nil-"})
		return
	}

	GlobalLogger.Log(LogData{"logname": name, "comment": note, "message": msg})
}

// Check a filename and see if logging is on for that filename
// If it never ben see then check with the regex. If it has been seen then just look it up in the map
// assumes traceMutex is locked already

func CheckFileName(name string) bool { return true } // hack for now

func addmsg(msg interfaces.IMsg) {
	msghistlock.Lock()
	defer msghistlock.Unlock()
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
	msghistlock.Lock()
	defer msghistlock.Unlock()
	if msgmap == nil {
		msgmap = make(map[[32]byte]interfaces.IMsg)
	}
	rval, _ := msgmap[hash]
	return rval
}

// Look up the hex string in the map of names...
func lookupName(s string) string {
	n, ok := globals.FnodeNames[s]
	if ok {
		return "<" + n + ">"
	}
	return ""
}

var findHex *regexp.Regexp

func init() {
	var err error
	// regex to find 6 hex digit strings, used t add names to hex ID in logs
	// or Fed or Aud so the columns in authority set listing looks right in logs
	findHex, err = regexp.Compile("(Fed)|(Aud)|([A-Fa-f0-9]{6,})")
	if err != nil {
		panic(err)
	}
}

// Look thru a string and annotate the string with names based on any hex strings
func addNodeNames(s string) (rval string) {
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
			s = s[:hexStr[i][1]] + lookupName(l) + s[hexStr[i][1]:]
		} else {
			// Shorten 32 byte IDs to [3:6]
			l := s[hexStr[i][0]:hexStr[i][1]]
			name := lookupName(l)
			if name != "" {
				s = s[:hexStr[i][0]] + s[hexStr[i][0]+6:hexStr[i][0]+12] + name + s[hexStr[i][1]:]
			}
		}
	}
	return s
}

// this is where eventually we will handle the addname and message history stuff we used to do
func MsgFormatter(v interface{}) string {
	msg, _ := v.(interfaces.IMsg)

	if msg == nil {
		// include nonempty hash to get spacing right
		return "-nil-"
	}

	hash := "??????"
	mhash := "??????"
	rhash := "??????"
	t := byte(0)
	var embeddedMsg interfaces.IMsg

	t = msg.Type()
	// work around message that don't have hashes yet ...
	mh := msg.GetMsgHash()
	if mh != nil && !reflect.ValueOf(mh).IsNil() {
		mhash = mh.String()[:6]
	}
	h := msg.GetHash()
	if h != nil && !reflect.ValueOf(h).IsNil() {
		hash = h.String()[:6]
	}
	rh := msg.GetRepeatHash()
	if rh != nil && !reflect.ValueOf(rh).IsNil() {
		rhash = rh.String()[:6]
	}

	if msg.Type() != constants.ACK_MSG && msg.Type() != constants.MISSING_DATA {
		addmsg(msg) // Keep message we have seen for a while
	}

	// to := ""
	//if msg.IsPeer2Peer() {
	//	if 0 == msg.GetOrigin() {
	//		to = "RandomPeer"
	//	} else {
	//		// right for sim... what about network ?
	//		to = fmt.Sprintf("FNode%02d", msg.GetOrigin()-1)
	//	}
	//} else {
	//	//to = "broadcast"
	//}

	embeddedHash := ""
	switch t {
	case constants.VOLUNTEERAUDIT, constants.ACK_MSG:
		embeddedHash = fmt.Sprintf(" EmbeddedMsg: m-%x", msg.GetHash().Bytes()[:3])
		fixed := msg.GetHash().Fixed()
		embeddedMsg = getmsg(fixed)
		if embeddedMsg == nil {
			embeddedHash += "(unknown)"
		}
	}
	if embeddedMsg == nil {
		return addNodeNames(fmt.Sprintf("M-%v|R-%v|H-%v|%p %s EmbeddedMsg:%s", mhash, rhash, hash, msg, msg.String(), embeddedHash))
	}
	return addNodeNames(fmt.Sprintf("M-%v|R-%v|H-%v|%p %s %s\nEmbeddedMsg:%s", mhash, rhash, hash, msg, msg.String(), embeddedHash, MsgFormatter(embeddedMsg)))

}
