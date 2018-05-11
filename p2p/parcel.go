// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"hash/crc32"

	log "github.com/sirupsen/logrus"
)

var parcelLogger = packageLogger.WithField("subpack", "connection")

// Parcel is the atomic level of communication for the p2p network.  It contains within it the necessary info for
// the networking protocol, plus the message that the Application is sending.
type Parcel struct {
	Header  ParcelHeader
	Payload []byte
}

// ParcelHeaderSize is the number of bytes in a parcel header
// Note: PartNo and PartsTotal are currently unused, remove at the next protocol version bump
const ParcelHeaderSize = 32

type ParcelHeader struct {
	Network    NetworkID         // 4 bytes - the network we are on (eg testnet, main net, etc.)
	Version    uint16            // 2 bytes - the version of the protocol we are running.
	Type       ParcelCommandType // 2 bytes - network level commands (eg: ping/pong)
	Length     uint32            // 4 bytes - length of the payload (that follows this header) in bytes
	TargetPeer string            // ? bytes - "" or nil for broadcast, otherwise the destination peer's hash.
	Crc32      uint32            // 4 bytes - data integrity hash (of the payload itself.)

	PartNo     uint16 // 2 bytes - unused, remove at the next protocol version bump
	PartsTotal uint16 // 2 bytes - unused, remove at the next protocol version bump

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

func (p *Parcel) LogEntry() *log.Entry {
	return parcelLogger.WithFields(log.Fields{
		"network":     p.Header.Network.String(),
		"version":     p.Header.Version,
		"app_hash":    p.Header.AppHash,
		"app_type":    p.Header.AppType,
		"command":     CommandStrings[p.Header.Type],
		"length":      p.Header.Length,
		"target_peer": p.Header.TargetPeer,
		"crc32":       p.Header.Crc32,
		"node_id":     p.Header.NodeID,
		"part_no":     p.Header.PartNo + 1,
		"parts_total": p.Header.PartsTotal,
	})
}

func (p *Parcel) MessageType() string {
	return (fmt.Sprintf("[%s]", CommandStrings[p.Header.Type]))
}
