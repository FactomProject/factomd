// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"os"
)

// This is a global... where *should* it be?
var (
	CurrentLoggingLevel = Verbose // Start at verbose because it takes a few seconds for the controller to adjust to what you set.
	CurrentNetwork      = TestNet
)

const (
	// ProtocolVersion is the latest version this package supports
	ProtocolVersion uint16 = 01
	// ProtocolVersionMinimum is the earliest version this package supports
	ProtocolVersionMinimum uint16 = 01
	// Don't think we need this.
	// ProtocolCookie         uint32 = uint32([]bytes("Fact"))
)

// NOTE JAYJAY -- define node service levels (if we need them?)
// to allow us to filter what messages go to what nodes (eg: full nodes, etc.)
// But this feels a bit too much like the netowrking is getting itno the applications business.

// NetworkIdentifier represents the P2P network we are participating in (eg: test, nmain, etc.)
type NetworkID uint32

// Network indicators.
// TODO JAYJAY - this should go to a higher level, like the application levle?
const (
	// MainNet represents the production network
	MainNet NetworkID = 0xfeedbeef

	// TestNet represents a testing network
	TestNet NetworkID = 0xdeadbeef
)

// Map of network ids to strings for easy printing of network ID
var NetworkIDStrings = map[NetworkID]string{
	MainNet: "MainNet",
	TestNet: "TestNet",
}

func (n *NetworkID) String() string {
	if net, ok := NetworkIDStrings[*n]; ok {
		return net
	}
	return fmt.Sprintf("Unknown NetworkID: %x", *n)
}

const ( // iota is reset to 0
	Silence   uint8 = iota // 0 Say nothing. A log output with level "Silence" is ALWAYS printed.
	Fatal                  // 1 Log only fatal errors (fatal errors are always logged even on "Silence")
	Errors                 // 2 Log all errors (many errors may be expected)
	Notes                  // 3 Log notifications, usually significant events
	Debugging              // 4 Log diagnostic info, pretty low level
	Verbose                // 5 Log everything
)

// Map of network ids to strings for easy printing of network ID
var LoggingLevels = map[uint8]string{
	Silence:   "Silence",   // Say nothing. A log output with level "Silence" is ALWAYS printed.
	Fatal:     "Fatal",     // Log only fatal errors (fatal errors are always logged even on "Silence")
	Errors:    "Errors",    // Log all errors (many errors may be expected)
	Notes:     "Notes",     // Log notifications, usually significant events
	Debugging: "Debugging", // Log diagnostic info, pretty low level
	Verbose:   "Verbose",   // Log everything
}

func silence(linebreak bool, format string, v ...interface{}) {
	log(Silence, linebreak, format, v...)
}
func logfatal(linebreak bool, format string, v ...interface{}) {
	log(Fatal, linebreak, format, v...)
}
func logerror(linebreak bool, format string, v ...interface{}) {
	log(Errors, linebreak, format, v...)
}
func note(linebreak bool, format string, v ...interface{}) {
	log(Notes, linebreak, format, v...)
}
func debug(linebreak bool, format string, v ...interface{}) {
	log(Debugging, linebreak, format, v...)
}
func verbose(linebreak bool, format string, v ...interface{}) {
	log(Verbose, linebreak, format, v...)
}
func log(level uint8, linebreak bool, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	levelStr := LoggingLevels[level]
	breakStr := ""
	if linebreak {
		breakStr = "\n"
	}
	if level <= CurrentLoggingLevel { // lower level means more severe. "Silence" level always printed, overriding silence.
		fmt.Fprintf(os.Stdout, "%d (%s) %d/%d \t- %s  %s", os.Getpid(), levelStr, level, CurrentLoggingLevel, message, breakStr)
	}
	if level == Fatal {
		fmt.Fprintf(os.Stderr, "%d (%s) %d/%d ERROR:\t- %s  %s", os.Getpid(), levelStr, level, CurrentLoggingLevel, message, breakStr)

		// BUGBUG - take out this exit before shipping JAYJAY TODO
		os.Exit(1)
	}
}
