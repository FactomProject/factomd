package interfaces

// FIXME: make this an actual interface
type Log struct {
	LogPrintf       func(name string, format string, more ...interface{})
	LogMessage      func(name string, note string, msg IMsg)
	StateLogMessage func(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, comment string, msg IMsg)
	StateLogPrintf  func(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, format string, more ...interface{})
}
