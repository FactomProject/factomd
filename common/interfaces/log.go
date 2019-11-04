package interfaces

type Log interface {
	LogPrintf(name string, format string, more ...interface{})
	LogMessage(name string, note string, msg IMsg)
	StateLogMessage(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, comment string, msg IMsg)
	StateLogPrintf(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, format string, more ...interface{})
}
