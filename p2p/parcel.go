// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

// Parcel is the atomic level of communication for the p2p network.  It contains within it the necessary info for
// the networking protocol, plus the message that the Application is sending.
type Parcel struct {
	Header  ParcelHeader
	Payload interface{}
}

// ParcelHeaderSize is the number of bytes in a parcel header
const ParcelHeaderSize = 28

type ParcelHeader struct {
	// Cookie       uint32    // 4 bytes - magic cookie "Fact" //COOKIE - no cookie for now.
	Network      NetworkID         // 4 bytes - the network we are on (eg testnet, main net, etc.)
	Version      uint16            // 2 bytes - the version of the protocol we are running.
	Type         ParcelCommandType // 2 bytes - network level commands (eg: ping/pong)
	Length       uint32            // 4 bytes - length of the payload (that follows this header) in bytes
	ConnectionID uint64            // 8 bytes - Zero for broadcast, otherwise ID of the peer the message came from or is going to.
	Hash         [4]byte           // 4 bytes - data integrity hash (of the payload itself.)
}

type ParcelCommandType uint16

// Parcel commands -- all new commands should be added to the *end* of the list!
const ( // iota is reset to 0
	TypeHeartbeat    ParcelCommandType = iota // "Note, I'm still alive"
	TypePing                                  // "Are you there?"
	TypePong                                  // "yes, I'm here"
	TypeHello                                 // "TBD Share our public key, start talking"
	TypeNetworkError                          // eg: "you sent me a message larger than max payload ParcelHeaderSize""
	TypeAlert                                 // network wide alerts (used in bitcoin to indicate criticalities)
	TypeMessage                               // Application level message
)

// CommandStrings is a Map of command ids to strings for easy printing of network comands
var CommandStrings = map[ParcelCommandType]string{
	TypeHeartbeat:    "Heartbeat",    // "Note, I'm still alive"
	TypePing:         "Ping",         // "Are you there?"
	TypePong:         "Pong",         // "yes, I'm here"
	TypeHello:        "Hello",        // "TBD Share our public key, start talking"
	TypeNetworkError: "NetworkError", // eg: "you sent me a message larger than max payload ParcelHeaderSize""
	TypeAlert:        "Alert",        // network wide alerts (used in bitcoin to indicate criticalities)
	TypeMessage:      "Message",      // Application level message
}

// MaxPayloadSize is the maximum bytes a message can be at the networking level.
const MaxPayloadSize = (1024 * 512) // 512KB

func (p *ParcelHeader) Init(network NetworkID) *ParcelHeader {
	// p.Cookie = ProtocolCookie //COOKIE - no cookie for now.
	p.Network = network
	p.Version = ProtocolVersion
	p.Type = TypeMessage
	p.ConnectionID = uint64(0)
	return p
}
func (p *Parcel) Init(header ParcelHeader) *Parcel {
	p.Header = header
	return p
}

func (p *ParcelHeader) Print() {
	// log(Debugging, true, "\t Cookie: \t%+v", string(p.Cookie))
	log(Debugging, true, "\t Network:\t%+v", NetworkIDStrings[p.Network])
	log(Debugging, true, "\t Payload:\t%+v", p.Version)
	log(Debugging, true, "\t Type:   \t%+v", CommandStrings[p.Type])
	log(Debugging, true, "\t Length:\t%+d", p.Length)
	log(Debugging, true, "\t ConnectionID:\t%+d", p.ConnectionID)
	log(Debugging, true, "\t Hash:\t%+d", p.Hash)
}

func (p *Parcel) Print() {
	p.Header.Print()
	log(Debugging, true, "\t\tPayload: %+v", p.Payload)
}

func (p *Parcel) PrintMessageType() {
	p.Header.Print()
	log(Notes, false, "[%+v]", CommandStrings[p.Header.Type])
}
