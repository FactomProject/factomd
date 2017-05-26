// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/log"
)

// setup subsystem loggers
var (
	rpcLog    *log.FLogger
	serverLog *log.FLogger
	wsLog     *log.FLogger
)

func InitLogs(logPath, logLevel string) {
	rpcLog = log.NewLogFromConfig(logPath, logLevel, "RPC")
	serverLog = log.NewLogFromConfig(logPath, logLevel, "SERV")
	wsLog = log.NewLogFromConfig(logPath, logLevel, "WSAPI")
}
