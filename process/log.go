// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package process

import (
	"github.com/FactomProject/factomd/logger"
	"github.com/FactomProject/factomd/util"
)

// setup subsystem loggers
var (
	procLog = logger.NewLogFromConfig(util.ReadConfig().Log.LogPath, util.ReadConfig().Log.LogLevel, "PROC")
)
