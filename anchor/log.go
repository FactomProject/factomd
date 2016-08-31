// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package anchor

import (
	"os"
	"strings"

	"github.com/FactomProject/factomd/logger"
	"github.com/FactomProject/factomd/util"
)

var (
	cfg        = util.ReadConfig("")
	homedir    = cfg.App.HomeDir
	network    = strings.ToLower(cfg.App.Network) + "-"
	logPath    = cfg.Log.LogPath
	logLevel   = cfg.Log.LogLevel
	logfile, _ = os.OpenFile(homedir+network+logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
)

// setup subsystem loggers
var (
	anchorLog = logger.New(logfile, logLevel, "ANCH")
)
