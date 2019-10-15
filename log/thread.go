package log

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
)

// returns thread_id, <filename>:<line> where the thread was spawned
type CallerHandle func() (threadID int, threadCaller string)

type ThreadLogger struct {
	interfaces.Log
	Caller CallerHandle
}

// TODO: make this an arg to New instead of CallerHandle
type ThreadHandle interface{
	GetID() int
	GetCaller() string
}

// allow for thread-aware logging
func New(caller CallerHandle) *ThreadLogger {
	// FIXME:: can we mak
	return &ThreadLogger{
		Caller: caller,
	}
}

// REVIEW: may want to design a different method of adding thread/caller to logs
// add thread id/caller to message or formatter
func extendFormat(caller CallerHandle, format string) string {
	t, c := caller()
	return fmt.Sprintf("%s %v/%v", format, t, c)
}

func (l *ThreadLogger) LogPrintf(name string, format string, more ...interface{}) {
	PackageLogger.LogPrintf(name, extendFormat(l.Caller, format), more...)
}

func (l *ThreadLogger) LogMessage(name string, note string, msg interfaces.IMsg) {
	PackageLogger.LogMessage(name, extendFormat(l.Caller, note), msg)
}

func (l *ThreadLogger) StateLogMessage(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, comment string, msg interfaces.IMsg) {
	PackageLogger.StateLogMessage(
		FactomNodeName,
		DBHeight,
		CurrentMinute,
		logName,
		extendFormat(l.Caller, comment),
		msg)
}

func (l *ThreadLogger) StateLogPrintf(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, format string, more ...interface{}) {
	PackageLogger.StateLogPrintf(
		FactomNodeName,
		DBHeight,
		CurrentMinute,
		logName,
		extendFormat(l.Caller, format),
		more...)
}
