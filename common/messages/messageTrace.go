package messages

import "github.com/FactomProject/factomd/log"

/*
KLUDGE: refactor to expose logging methods
under original location inside messages package
 */
var LogPrintf = log.LogPrintf
var CheckFileName = log.CheckFileName
var StateLogMessage = log.StateLogMessage
var StateLogPrintf = log.StateLogPrintf
var LogMessage = log.LogMessage

