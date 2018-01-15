// Copyright 2017 Factom Foundation
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

	"github.com/FactomProject/factomd/common/primitives"
	atomic2 "github.com/FactomProject/factomd/util/atomic"
)

// This file contains the global variables and utility functions for the p2p network operation.  The global variables and constants can be tweaked here.

// BlockFreeChannelSend will remove things from the queue to make room for new messages if the queue is full.
// This prevents channel blocking on full.
//		Returns: The number of elements cleared from the channel to make room
func BlockFreeChannelSend(channel chan interface{}, message interface{}) int {
	removed := 0
	highWaterMark := int(float64(cap(channel)) * 0.95)
	clen := len(channel)
	switch {
	case highWaterMark < clen:
		str, _ := primitives.EncodeJSONString(message)
		significant("protocol", "nonBlockingChanSend() - DROPPING MESSAGES. Channel is over 90 percent full! \n channel len: \n %d \n 90 percent: \n %d \n last message type: %v", len(channel), highWaterMark, str)
		for highWaterMark <= len(channel) { // Clear out some messages
			removed++
			<-channel
		}
		fallthrough
	default:
		select { // hits default if sending message would block.
		case channel <- message:
		default:
		}
	}
	return removed
}

// Global variables for the p2p protocol
var (
	CurrentLoggingLevelVar       atomic2.AtomicUint8 = atomic2.AtomicUint8(Errors) // Start at verbose because it takes a few seconds for the controller to adjust to what you set.
	CurrentNetwork                                   = TestNet
	NetworkListenPort                                = "8108"
	BroadcastFlag                                    = "<BROADCAST>"
	RandomPeerFlag                                   = "<RANDOMPEER>"
	NodeID                       uint64              = 0           // Random number used for loopback protection
	MinimumQualityScore          int32               = -200        // if a peer's score is less than this we ignore them.
	BannedQualityScore           int32               = -2147000000 // Used to ban a peer
	MinimumSharingQualityScore   int32               = 20          // if a peer's score is less than this we don't share them.
	OnlySpecialPeers                                 = false
	NetworkDeadline                                  = time.Duration(30) * time.Second
	NumberPeersToConnect                             = 32
	NumberPeersToBroadcast                           = 100
	MaxNumberIncomingConnections                     = 150
	MaxNumberOfRedialAttempts                        = 5 // How many missing pings (and other) before we give up and close.
	StandardChannelSize                              = 5000
	NetworkStatusInterval                            = time.Second * 9
	ConnectionStatusInterval                         = time.Second * 122
	PingInterval                                     = time.Second * 15
	TimeBetweenRedials                               = time.Second * 20
	PeerSaveInterval                                 = time.Second * 30
	PeerRequestInterval                              = time.Second * 180
	PeerDiscoveryInterval                            = time.Hour * 4

	// Testing metrics
	TotalMessagesReceived       uint64
	TotalMessagesSent           uint64
	ApplicationMessagesReceived uint64

	CRCKoopmanTable = crc32.MakeTable(crc32.Koopman)
	RandomGenerator *rand.Rand // seeded pseudo-random number generator

)

func CurrentLoggingLevel() uint8 {
	return CurrentLoggingLevelVar.Load()
}

const (
	// ProtocolVersion is the latest version this package supports
	ProtocolVersion uint16 = 8
	// ProtocolVersionMinimum is the earliest version this package supports
	ProtocolVersionMinimum uint16 = 8
)

// NetworkIdentifier represents the P2P network we are participating in (eg: test, nmain, etc.)
type NetworkID uint32

// Network indicators.
const (
	// MainNet represents the production network
	MainNet NetworkID = 0xfeedbeef

	// TestNet represents a testing network
	TestNet NetworkID = 0xdeadbeef

	// LocalNet represents any arbitrary/private network
	LocalNet NetworkID = 0xbeaded
)

func (n *NetworkID) String() string {
	switch *n {
	case MainNet:
		return "MainNet"
	case TestNet:
		return "TestNet"
	case LocalNet:
		return "LocalNet"
	default:
		return fmt.Sprintf("CustomNet ID: %x\n", *n)
	}
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

func dot(dot string) {
	if Notes < CurrentLoggingLevel() {
		switch dot {
		case "":
			fmt.Printf(".")
		default:
			fmt.Printf(dot)
		}
	}
}

func silence(component string, format string, v ...interface{}) {
	logP(Silence, component, format, v...)
}
func significant(component string, format string, v ...interface{}) {
	logP(Significant, component, format, v...)
}
func logfatal(component string, format string, v ...interface{}) {
	logP(Fatal, component, format, v...)
}
func logerror(component string, format string, v ...interface{}) {
	logP(Errors, component, format, v...)
}
func note(component string, format string, v ...interface{}) {
	logP(Notes, component, format, v...)
}
func debug(component string, format string, v ...interface{}) {
	logP(Debugging, component, format, v...)
}
func verbose(component string, format string, v ...interface{}) {
	logP(Verbose, component, format, v...)
}

// logP is the base log function to produce parsable log output for mass metrics consumption
func logP(level uint8, component string, format string, v ...interface{}) {
	message := strings.Replace(fmt.Sprintf(format, v...), ",", "-", -1) // Make CSV parsable.
	// levelStr := LoggingLevels[level]
	// host, _ := os.Hostname()
	// fmt.Fprintf(os.Stdout, "%s, %s, %d, %s, (%s), %d/%d, %s \n", now.String(), host, os.Getpid(), component, levelStr, level, CurrentLoggingLevel, message)

	now := time.Now().Format("2006-01-02 15:04:05")
	if level <= CurrentLoggingLevel() { // lower level means more severe. "Silence" level always printed, overriding silence.
		fmt.Printf("%s, %s, %s \n", now, component, message)
		// fmt.Fprintf(os.Stdout, "%s, %d, %s, (%s), %s\n", host, os.Getpid(), component, levelStr, message)
	}
	if level == Fatal {
		fmt.Println("===== SIGNIFICANT ERROR ====== \n Something is very wrong, and should be looked into!")
		fmt.Fprintf(os.Stderr, "%s, %s, %s \n", now, component, message)
		panic(message)
	}
}
