// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/logger"
)

// setup subsystem loggers
var (
	rpcLog    *logger.FLogger
	serverLog *logger.FLogger
	wsLog     *logger.FLogger
)

func InitLogs(logPath, logLevel string) {
	rpcLog = logger.NewLogFromConfig(logPath, logLevel, "RPC")
	serverLog = logger.NewLogFromConfig(logPath, logLevel, "SERV")
	wsLog = logger.NewLogFromConfig(logPath, logLevel, "WSAPI")
}
