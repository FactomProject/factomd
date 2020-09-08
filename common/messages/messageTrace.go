package messages

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
)

/*
KLUDGE: refactor to expose logging methods
under original location inside messages package
*/
var LogPrintf = log.LogPrintf
var CheckFileName = log.CheckFileName
var LogMessage = log.LogMessage

type foo func(data []byte) (interfaces.IMsg, error)

var FP func(data []byte) (interfaces.IMsg, error)

// Hack to get around import loop
func Unmarshal_Message(data []byte) (interfaces.IMsg, error) {
	msg, err := FP(data)
	return msg, err
}
