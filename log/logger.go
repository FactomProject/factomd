package log

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/util/atomic"
)

type log struct {
	interfaces.Log
}

var (

	// KLUDGE: expose package logging for backward compatibility
	PackageLogger   = &log{}
	LogPrintf       = PackageLogger.LogPrintf
	LogMessage      = PackageLogger.LogMessage
	StateLogMessage = PackageLogger.StateLogMessage
	StateLogPrintf  = PackageLogger.StateLogPrintf
)

func (*log) LogPrintf(name string, format string, more ...interface{}) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	myfile := getTraceFile(name)
	if myfile == nil {
		return
	}

	var where string

	if logWhere { // global for debugging
		where = fmt.Sprintf("<%s>", atomic.Goid())
	}

	sequence++
	// handle multi-line printf's
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
}

// Log a message with a state timestamp
func (l *log) StateLogMessage(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, comment string, msg interfaces.IMsg) {
	logFileName := FactomNodeName + "_" + logName + ".txt"
	t := fmt.Sprintf("%7d-:-%d ", DBHeight, CurrentMinute)
	LogMessage(logFileName, t+comment, msg)
}

// Log a printf with a state timestamp
func (l *log) StateLogPrintf(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, format string, more ...interface{}) {
	logFileName := FactomNodeName + "_" + logName + ".txt"
	t := fmt.Sprintf("%7d-:-%d ", DBHeight, CurrentMinute)
	l.LogPrintf(logFileName, t+format, more...)
}

func (l *log) LogMessage(name string, note string, msg interfaces.IMsg) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	l.logMessage(name, note, msg)
}

// Assumes called managed the locks so we can recurse for multi part messages
func (l *log) logMessage(name string, note string, msg interfaces.IMsg) {
	myfile := getTraceFile(name)
	if myfile == nil {
		return
	}

	var where string

	if logWhere {
		where = fmt.Sprintf("<%s>", atomic.Goid())
	}

	sequence++
	embeddedHash := ""
	to := ""
	hash := "??????"
	mhash := "??????"
	rhash := "??????"
	messageType := ""
	t := byte(0)
	msgString := "-nil-"
	var embeddedMsg interfaces.IMsg

	if msg != nil {
		t = msg.Type()
		msgString = msg.String()

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

		if msg.IsPeer2Peer() {
			if 0 == msg.GetOrigin() {
				to = "RandomPeer"
			} else {
				// right for sim... what about network ?
				to = fmt.Sprintf("FNode%02d", msg.GetOrigin()-1)
			}
		} else {
			//to = "broadcast"
		}
		switch t {
		case constants.VOLUNTEERAUDIT, constants.ACK_MSG:
			embeddedHash = fmt.Sprintf(" EmbeddedMsg: %x", msg.GetHash().Bytes()[:3])
			fixed := msg.GetHash().Fixed()
			embeddedMsg = getmsg(fixed)
			if embeddedMsg == nil {
				embeddedHash += "(unknown)"
			}
		}
		messageType = constants.MessageName(byte(t))
	}

	// handle multi-line printf's
	lines := strings.Split(msgString, "\n")

	lines[0] = lines[0] + embeddedHash + " " + to

	now := time.Now().Local()

	for i, text := range lines {
		var s string
		switch i {
		case 0:
			s = fmt.Sprintf("%9d %02d:%02d:%02d.%03d %-50s M-%v|R-%v|H-%v|%p %26s[%2v]:%v %s\n", sequence, now.Hour()%24, now.Minute()%60, now.Second()%60, (now.Nanosecond()/1e6)%1000,
				note, mhash, rhash, hash, msg, messageType, t, text, where)
		default:
			s = fmt.Sprintf("%9d %02d:%02d:%02d.%03d %-50s M-%v|R-%v|H-%v|%p %30s:%v\n", sequence, now.Hour()%24, now.Minute()%60, now.Second()%60, (now.Nanosecond()/1e6)%1000,
				note, mhash, rhash, hash, msg, "continue:", text)
		}
		s = addNodeNames(s)
		myfile.WriteString(s)
	}

	if embeddedMsg != nil {
		l.logMessage(name, note+" EmbeddedMsg:", embeddedMsg)
	}
}
