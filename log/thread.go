package log

import "github.com/FactomProject/factomd/common/interfaces"

type ThreadLogger struct {
	interfaces.Log
	ThreadID int
	Caller *string
	logger *log
}

type CallerHandler func() string

// allow for thread-aware logging
func New(threadID int, caller *string) *ThreadLogger {
	return &ThreadLogger{
		ThreadID: threadID,
		Caller: caller,
		logger: &log{},
	}
}

func (l *ThreadLogger) LogPrintf(name string, format string, more ...interface{}) {
	// TODO inject thread info into log
	_ = l.Caller
	_ = l.ThreadID
}

func (l *ThreadLogger) LogMessage(name string, note string, msg interfaces.IMsg) {
	// TODO inject thread info into log
}

func (l *ThreadLogger) StateLogMessage(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, comment string, msg interfaces.IMsg) {
	// TODO inject thread info into log
}

func (l *ThreadLogger) StateLogPrintf(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, format string, more ...interface{}) {
	// TODO inject thread info into log
}
