// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/logger"
	"github.com/FactomProject/factomd/util"
)

// setup subsystem loggers
var (
	rpcLog    = logger.NewLogFromConfig(util.ReadConfig().Log.LogPath, util.ReadConfig().Log.logLevel, "RPC")
	serverLog = logger.NewLogFromConfig(util.ReadConfig().Log.LogPath, util.ReadConfig().Log.logLevel, "SERV")
	wsLog     = logger.NewLogFromConfig(util.ReadConfig().Log.LogPath, util.ReadConfig().Log.logLevel, "WSAPI")
)
