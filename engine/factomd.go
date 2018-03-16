// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"runtime"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"

	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

// winServiceMain is only invoked on Windows.  It detects when btcd is running
// as a service and reacts accordingly.
//var winServiceMain func() (bool, error)

// packageLogger is the general logger for all engine related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{"package": "engine"})

// Build sets the factomd build id using git's SHA
// Version sets the semantic version number of the build
// $ go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.=`cat VERSION`"
// It also seems to need to have the previous binary deleted if recompiling to have this message show up if no code has changed.
// Since we are tracking code changes, then there is no need to delete the binary to use the latest message
var Build string
var FactomdVersion string = "BuiltWithoutVersion"

func Factomd(params *FactomParams, listenToStdin bool) interfaces.IState {
	log.Printf("Go compiler version: %s\n", runtime.Version())
	log.Printf("Using build: %s\n", Build)
	log.Printf("Version: %s\n", FactomdVersion)


	state0 := new(state.State)
	state0.IsRunning = true
	state0.SetLeaderTimestamp(primitives.NewTimestampFromMilliseconds(0))

	go NetStart(state0, params, listenToStdin)
	return state0
}
