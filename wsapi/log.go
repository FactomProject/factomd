// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// setup subsystem loggers
var (
	wsDebugLog *log.Entry
	wsLog      *log.Entry
)

// NewLogFromConfig outputs logs to a file given by logpath
func NewLogFromConfig(logPath, logLevel, prefix string) *log.Entry {
	var logFile io.Writer
	if logPath == "stdout" {
		logFile = os.Stdout
	} else {
		logFile, _ = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	}

	logger := log.New()
	logger.SetOutput(logFile)
	if logLevel != "none" {
		lvl, err := log.ParseLevel(logLevel)
		if err != nil {
			panic(err)
		}
		logger.SetLevel(lvl)
	}
	return logger.WithField("prefix", prefix)
}

func InitLogs(logPath, logLevel string) {
	wsDebugLog = NewLogFromConfig(logPath, logLevel, "APIDEBUGLOG")
	wsLog = NewLogFromConfig(logPath, logLevel, "WSAPI")
}
