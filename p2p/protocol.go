// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"os"
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

// This is a global... where *should* it be?
var (
	P2PCurrentLoggingLevel = Silence
)

const ( // iota is reset to 0
	Silence   uint8 = iota // Say nothing. A log output with level "Silence" is ALWAYS printed.
	Fatal                  // Log only fatal errors (fatal errors are always logged even on "Silence")
	Errors                 // Log all errors (many errors may be expected)
	Notes                  // Log notifications, usually significant events
	Debugging              // Log diagnostic info, pretty low level
	Verbose                // Log everything
)

func silence(linebreak bool, format string, v ...interface{}) {
	log(Silence, linebreak, format, v...)
}
func fatal(linebreak bool, format string, v ...interface{}) {
	log(Fatal, linebreak, format, v...)
}
func error(linebreak bool, format string, v ...interface{}) {
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
	if level >= P2PCurrentLoggingLevel {
		breakStr := ""
		if linebreak {
			breakStr = "\n"
		}
		if level < Notes { // eg this is an error or higher
			fmt.Fprintf(os.Stderr, fmt.Sprintf("%d: %s MARK %s", os.Getpid()), fmt.Sprintf(format, v...), breakStr)
		} else {
			fmt.Fprintf(os.Stdout, fmt.Sprintf("%d: %s MARK %s", os.Getpid()), fmt.Sprintf(format, v...), breakStr)
		}
	}
	if level == Fatal {
		os.Exit(1)
	}
}
