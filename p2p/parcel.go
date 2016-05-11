// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"hash/crc32"
	"strconv"
	"time"
)

// Parcel is the atomic level of communication for the p2p network.  It contains within it the necessary info for
// the networking protocol, plus the message that the Application is sending.
type Parcel struct {
	Header  ParcelHeader
	Payload []byte
}

// ParcelHeaderSize is the number of bytes in a parcel header
const ParcelHeaderSize = 28

type ParcelHeader struct {
	Network      NetworkID         // 4 bytes - the network we are on (eg testnet, main net, etc.)
	Version      uint16            // 2 bytes - the version of the protocol we are running.
	Type         ParcelCommandType // 2 bytes - network level commands (eg: ping/pong)
	Length       uint32            // 4 bytes - length of the payload (that follows this header) in bytes
	ConnectionID string            // ? bytes - "" or nil for broadcast, otherwise the destination peer's hash.
	Crc32        uint32            // 4 bytes - data integrity hash (of the payload itself.)
	Timestamp    time.Time
}

type ParcelCommandType uint16

// Parcel commands -- all new commands should be added to the *end* of the list!
const ( // iota is reset to 0
	TypeHeartbeat    ParcelCommandType = iota // "Note, I'm still alive"
	TypePing                                  // "Are you there?"
	TypePong                                  // "yes, I'm here"
	TypePeerRequest                           // "Please share some peers"
	TypePeerResponse                          // "Here's some peers I know about."
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

func NewParcel(network NetworkID, payload []byte) *Parcel {
	header := new(ParcelHeader).Init(network)
	parcel := new(Parcel).Init(*header)
	parcel.Payload = payload
	parcel.UpdateHeader() // Updates the header with info about payload.
	return parcel
}

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

func (p *Parcel) UpdateHeader() *Parcel {
	p.Header.Crc32 = crc32.Checksum(p.Payload, CRCKoopmanTable)
	p.Header.Length = uint32(len(p.Payload))
}

func (p *ParcelHeader) Print() {
	// debug( true, "\t Cookie: \t%+v", string(p.Cookie))
	debug(true, "\t Network:\t%+v", NetworkIDStrings[p.Network])
	debug(true, "\t Version:\t%+v", p.Version)
	debug(true, "\t Type:   \t%+v", CommandStrings[p.Type])
	debug(true, "\t Length:\t%+d", p.Length)
	debug(true, "\t ConnectionID:\t%+d", p.ConnectionID)
	debug(true, "\t Hash:\t%+d", p.Hash)
}

func (p *Parcel) Print() {
	debug(true, "Pretty Printing Parcel:")
	p.Header.Print()
	switch p.Payload.(type) {
	case string:
		debug(true, "\t\tPayload: %s", p.Payload)
	case []byte:
		s := strconv.Quote(string(p.Payload.([]byte)))
		debug(true, "\t\tPayload: %s", s)
	default:
		debug(true, "\t\tPayload: %+v", p.Payload)

	}
}

func (p *Parcel) PrintMessageType() {
	log(Notes, false, "[%+v]", CommandStrings[p.Header.Type])
}
