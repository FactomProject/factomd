package p2p

import (
    
)

const (
    // ProtocolVersion is the latest version this package supports
    ProtocolVersion uint16 = 01
    // ProtocolVersionMinimum is the earliest version this package supports
    ProtocolVersionMinimum uint16 = 01
    ProtocolCookie uint32 = binary("Fact")
)

// NOTE JAYJAY -- define node service levels (if we need them?) 
// to allow us to filter what messages go to what nodes (eg: full nodes, etc.)
// But this feels a bit too much like the netowrking is getting itno the applications business.

// NetworkIdentifier represents the P2P network we are participating in (eg: test, nmain, etc.)
type NetworkID uint32

// Network indicators.
// TODO JAYJAY - this should go to a higher level, like the application levle?
const NetworkID (
    // MainNet represents the production network
    MainNet = 0xfeedbeef
    
    // TestNet represents a testing network
    TestNet = 0xdeadbeef
)

// Map of network ids to strings for easy printing of network ID
var networkIDStrings = map[NetworkID]string{
    MainNet: "MainNet",
    TestNet: "TestNet",
}

func (n *NetworkID) String() string {
    if net, ok := networkIDStrings[n]; ok {
        return net
    }
    return fmt.Sprintf("Unknown NetworkID: %x", uint32(n))
}

// This is a global... where *should* it be?
var (
    P2PCurrentLoggingLevel := Silence
)

const uint8 ( // iota is reset to 0
	Silence = iota    // Say nothing. A log output with level "Silence" is ALWAYS printed.
    Fatal             // Log only fatal errors (fatal errors are always logged even on "Silence")
    Errors            // Log all errors (many errors may be expected)
    Notes             // Log notifications, usually significant events
    Debugging         // Log diagnostic info, pretty low level
    Verbose           // Log everything
)

func log(level uint8, format string, v ...interface{})  {
    if level >= P2PCurrentLoggingLevel {
        if level < Notes { // eg this is an error or higher
            	fmt.Fprintln(os.Stderr, fmt.Sprintf("%d:", os.Getpid()), fmt.Sprintf(format, v...))
        } else {
            	fmt.Fprintln(os.Stdout, fmt.Sprintf("%d:", os.Getpid()), fmt.Sprintf(format, v...))
        }
    }
    if level == Fatal {
        os.Exit(1)
    }
}