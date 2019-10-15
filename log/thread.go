package log

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
)

// returns thread_id, <filename>:<line> where the thread was spawned

type ICaller interface {
	GetID() int
	GetCaller() string
}

type ThreadLogger struct {
	interfaces.Log
	Caller ICaller
}

// allow for thread-aware logging
func New(caller ICaller) *ThreadLogger {
	return &ThreadLogger{
		Caller: caller,
	}
}

// REVIEW: may want to design a different method of adding thread/caller to logs
// add thread id/caller to message or formatter
func extendFormat(caller ICaller, format string) string {
	return fmt.Sprintf("%s %v/%v", format, caller.GetID(), caller.GetCaller())
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
