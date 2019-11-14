package log

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

//
//import (
//	"github.com/FactomProject/factomd/common/interfaces"
//)
//
//type log struct {
//	interfaces.Log
//}
//
//var (
//
//	// KLUDGE: expose package logging for backward compatibility
//	PackageLogger   = &log{}
//	StateLogMessage = PackageLogger.StateLogMessage
//	StateLogPrintf  = PackageLogger.StateLogPrintf
//)

// Log a message with a state timestamp
func StateLogMessage(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, comment string, msg interfaces.IMsg) {
	logFileName := FactomNodeName + "_" + logName + ".txt"
	t := fmt.Sprintf("%7d-:-%d ", DBHeight, CurrentMinute)
	LogMessage(logFileName, t+comment, msg)
}

// Log a printf with a state timestamp
func StateLogPrintf(FactomNodeName string, DBHeight int, CurrentMinute int, logName string, format string, more ...interface{}) {
	logFileName := FactomNodeName + "_" + logName + ".txt"
	t := fmt.Sprintf("%7d-:-%-2d ", DBHeight, CurrentMinute)
	LogPrintf(logFileName, t+format, more...)
}
