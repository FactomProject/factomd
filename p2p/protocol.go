// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Global variables for the p2p protocol
var (
	CurrentLoggingLevel                         = Verbose // Start at verbose because it takes a few seconds for the controller to adjust to what you set.
	CurrentNetwork                              = TestNet
	NetworkListenPort             string        = "8108"
	NodeID                        uint64        = 0           // Random number used for loopback protection
	MinumumQualityScore           int32         = -200        // if a peer's score is less than this we ignore them.
	BannedQualityScore            int32         = -2147000000 // Used to ban a peer
	OnlySpecialPeers              bool          = false
	NumberPeersToConnect          int           = 12
	MaxNumberIncommingConnections int           = 150
	MaxNumberOfRedialAttempts     int           = 15
	NetworkStatusInterval         time.Duration = time.Second * 22
	PingInterval                  time.Duration = time.Second * 15
	TimeBetweenRedials            time.Duration = time.Second * 20
	PeerSaveInterval              time.Duration = time.Second * 30
	PeerRequestInterval           time.Duration = time.Second * 180

	// Testing metrics
	TotalMessagesRecieved       uint64 = 0
	TotalMessagesSent           uint64 = 0
	ApplicationMessagesRecieved uint64 = 0

	CRCKoopmanTable *crc32.Table = crc32.MakeTable(crc32.Koopman)
	RandomGenerator *rand.Rand   // seeded pseudo-random number generator

)

const (
	// ProtocolVersion is the latest version this package supports
	ProtocolVersion uint16 = 01
	// ProtocolVersionMinimum is the earliest version this package supports
	ProtocolVersionMinimum uint16 = 01
	// Don't think we need this.
	// ProtocolCookie         uint32 = uint32([]bytes("Fact"))
	// Used in generating message CRC values
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

	// LocalNet represents any arbitrary/private network
	LocalNet NetworkID = 0xbeaded
)

// Map of network ids to strings for easy printing of network ID
var NetworkIDStrings = map[NetworkID]string{
	MainNet:  "MainNet",
	TestNet:  "TestNet",
	LocalNet: "LocalNet",
}

func (n *NetworkID) String() string {
	if net, ok := NetworkIDStrings[*n]; ok {
		return net
	}
	return fmt.Sprintf("Unknown NetworkID: %x", *n)
}

const ( // iota is reset to 0
	Silence     uint8 = iota // 0 Say nothing. A log output with level "Silence" is ALWAYS printed.
	Significant              // 1 Significant messages that should be logged in normal ops
	Fatal                    // 2 Log only fatal errors (fatal errors are always logged even on "Silence")
	Errors                   // 3 Log all errors (many errors may be expected)
	Notes                    // 4 Log notifications, usually significant events
	Debugging                // 5 Log diagnostic info, pretty low level
	Verbose                  // 6 Log everything
)

// Map of network ids to strings for easy printing of network ID
var LoggingLevels = map[uint8]string{
	Silence:     "Silence",     // Say nothing. A log output with level "Silence" is ALWAYS printed.
	Significant: "Significant", // Significant things that should be printed, but aren't necessary errors.
	Fatal:       "Fatal",       // Log only fatal errors (fatal errors are always logged even on "Silence")
	Errors:      "Errors",      // Log all errors (many errors may be expected)
	Notes:       "Notes",       // Log notifications, usually significant events
	Debugging:   "Debugging",   // Log diagnostic info, pretty low level
	Verbose:     "Verbose",     // Log everything
}

func silence(component string, format string, v ...interface{}) {
	log(Silence, component, format, v...)
}
func significant(component string, format string, v ...interface{}) {
	log(Significant, component, format, v...)
}
func logfatal(component string, format string, v ...interface{}) {
	log(Fatal, component, format, v...)
}
func logerror(component string, format string, v ...interface{}) {
	log(Errors, component, format, v...)
}
func note(component string, format string, v ...interface{}) {
	log(Notes, component, format, v...)
}
func debug(component string, format string, v ...interface{}) {
	log(Debugging, component, format, v...)
}
func verbose(component string, format string, v ...interface{}) {
	log(Verbose, component, format, v...)
}

// log is the base log function to produce parsable log output for mass metrics consumption
func log(level uint8, component string, format string, v ...interface{}) {
	message := strings.Replace(fmt.Sprintf(format, v...), ",", "-", -1) // Make CSV parsable.
	// levelStr := LoggingLevels[level]
	// host, _ := os.Hostname()
	// fmt.Fprintf(os.Stdout, "%s, %s, %d, %s, (%s), %d/%d, %s \n", now.String(), host, os.Getpid(), component, levelStr, level, CurrentLoggingLevel, message)

	now := time.Now().Format("01/02/2006 15:04:05.000")
	if level <= CurrentLoggingLevel { // lower level means more severe. "Silence" level always printed, overriding silence.
		fmt.Fprintf(os.Stdout, "%s, %s, %d/%d, %s \n", now, component, level, CurrentLoggingLevel, message)
		// fmt.Fprintf(os.Stdout, "%s, %d, %s, (%s), %s\n", host, os.Getpid(), component, levelStr, message)
	}
	if level == Fatal {
		fmt.Fprintf(os.Stderr, "%s, %s, %d/%d, %s \n", now, component, level, CurrentLoggingLevel, message)
		// BUGBUG - take out this exit before shipping JAYJAY TODO, or check that all fatals are fatal.
		os.Exit(1)
	}
}
