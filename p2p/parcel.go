// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
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
	Network     NetworkID         // 4 bytes - the network we are on (eg testnet, main net, etc.)
	Version     uint16            // 2 bytes - the version of the protocol we are running.
	Type        ParcelCommandType // 2 bytes - network level commands (eg: ping/pong)
	Length      uint32            // 4 bytes - length of the payload (that follows this header) in bytes
	TargetPeer  string            // ? bytes - "" or nil for broadcast, otherwise the destination peer's hash.
	Crc32       uint32            // 4 bytes - data integrity hash (of the payload itself.)
	NodeID      uint64
	PeerAddress string // address of the peer set by connection to know who sent message (for tracking source of other peers)
	PeerPort    string // port of the peer , or we are listening on
	AppHash     string // Application specific message hash, for tracing
	AppType     string // Application specific message type, for tracing

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
	TypeHeartbeat:    "Heartbeat",     // "Note, I'm still alive"
	TypePing:         "Ping",          // "Are you there?"
	TypePong:         "Pong",          // "yes, I'm here"
	TypePeerRequest:  "Peer-Request",  // "Please share some peers"
	TypePeerResponse: "Peer-Response", // "Here's some peers I know about."
	TypeAlert:        "Alert",         // network wide alerts (used in bitcoin to indicate criticalities)
	TypeMessage:      "Message",       // Application level message
}

// MaxPayloadSize is the maximum bytes a message can be at the networking level.
const MaxPayloadSize = (1024 * 512) // 512KB

func NewParcel(network NetworkID, payload []byte) *Parcel {
	header := new(ParcelHeader).Init(network)
	header.AppHash = "NetworkMessage"
	header.AppType = "Network"
	parcel := new(Parcel).Init(*header)
	parcel.Payload = payload
	parcel.UpdateHeader() // Updates the header with info about payload.
	return parcel
}

func (p *ParcelHeader) Init(network NetworkID) *ParcelHeader {
	p.Network = network
	p.Version = ProtocolVersion
	p.Type = TypeMessage
	p.TargetPeer = ""              // initially no target
	p.PeerPort = NetworkListenPort // store our listening port
	return p
}
func (p *Parcel) Init(header ParcelHeader) *Parcel {
	p.Header = header
	return p
}

func (p *Parcel) UpdateHeader() {
	p.Header.Crc32 = crc32.Checksum(p.Payload, CRCKoopmanTable)
	p.Header.Length = uint32(len(p.Payload))
}

func (p *Parcel) Trace(location string, sequence string) {
	time := time.Now().Unix()
	fmt.Printf("\nParcelTrace, %s, %s, %s, %s, %s, %d \n", p.Header.AppHash, sequence, p.Header.AppType, CommandStrings[p.Header.Type], location, time)
}

func (p *ParcelHeader) Print() {
	// debug( true, "\t Cookie: \t%+v", string(p.Cookie))
	debug("parcel", "\t Network:\t%+v", NetworkIDStrings[p.Network])
	debug("parcel", "\t Version:\t%+v", p.Version)
	debug("parcel", "\t Type:   \t%+v", CommandStrings[p.Type])
	debug("parcel", "\t Length:\t%d", p.Length)
	debug("parcel", "\t TargetPeer:\t%s", p.TargetPeer)
	debug("parcel", "\t CRC32:\t%d", p.Crc32)
	debug("parcel", "\t NodeID:\t%d", p.NodeID)
}

func (p *Parcel) Print() {
	debug("parcel", "Pretty Printing Parcel:")
	p.Header.Print()
	s := strconv.Quote(string(p.Payload))
	debug("parcel", "\t\tPayload: %s", s)
}

func (p *Parcel) MessageType() string {
	return (fmt.Sprintf("[%s]", CommandStrings[p.Header.Type]))
}

func (p *Parcel) PrintMessageType() {
	fmt.Printf("[%+v]", CommandStrings[p.Header.Type])
}

func (p *Parcel) String() string {
	var output string
	s := strconv.Quote(string(p.Payload))
	fmt.Sprintf(output, "%s\t Network:\t%+v\n", output, NetworkIDStrings[p.Header.Network])
	fmt.Sprintf(output, "%s\t Version:\t%+v\n", output, p.Header.Version)
	fmt.Sprintf(output, "%s\t Type:   \t%+v\n", output, CommandStrings[p.Header.Type])
	fmt.Sprintf(output, "%s\t Length:\t%d\n", output, p.Header.Length)
	fmt.Sprintf(output, "%s\t TargetPeer:\t%s\n", output, p.Header.TargetPeer)
	fmt.Sprintf(output, "%s\t CRC32:\t%d\n", output, p.Header.Crc32)
	fmt.Sprintf(output, "%s\t NodeID:\t%d\n", output, p.Header.NodeID)
	fmt.Sprintf(output, "%s\t Payload: %s\n", output, s)
	return output
}
