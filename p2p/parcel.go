// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"bytes"
	"fmt"
	"hash/crc32"

	"github.com/FactomProject/factomd/common/interfaces"
	log "github.com/sirupsen/logrus"
)

var parcelLogger = packageLogger.WithField("subpack", "connection")

// Parcel is the atomic level of communication for the p2p network.  It contains within it the necessary info for
// the networking protocol, plus the message that the Application is sending.
type Parcel struct {
	Header  ParcelHeader
	Payload []byte
	Msg     interfaces.IMsg `json:"-"` // Keep the message for debugging
}

// ParcelHeaderSize is the number of bytes in a parcel header
const ParcelHeaderSize = 32

type ParcelHeader struct {
	Network     NetworkID         // 4 bytes - the network we are on (eg testnet, main net, etc.)
	Version     uint16            // 2 bytes - the version of the protocol we are running.
	Type        ParcelCommandType // 2 bytes - network level commands (eg: ping/pong)
	Length      uint32            // 4 bytes - length of the payload (that follows this header) in bytes
	TargetPeer  string            // ? bytes - "" or nil for broadcast, otherwise the destination peer's hash.
	Crc32       uint32            // 4 bytes - data integrity hash (of the payload itself.)
	PartNo      uint16            // 2 bytes - in case of multipart parcels, indicates which part this corresponds to, otherwise should be 0
	PartsTotal  uint16            // 2 bytes - in case of multipart parcels, indicates the total number of parts that the receiver should expect
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
	TypeMessagePart                           // Application level message that was split into multiple parts
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
	TypeMessagePart:  "MessagePart",   // Application level message that was split into multiple parts
}

// MaxPayloadSize is the maximum bytes a message can be at the networking level.
const MaxPayloadSize = 1000000000

func NewParcel(network NetworkID, payload []byte) *Parcel {
	header := new(ParcelHeader).Init(network)
	header.AppHash = "NetworkMessage"
	header.AppType = "Network"
	parcel := new(Parcel).Init(*header)
	parcel.Payload = payload
	parcel.UpdateHeader() // Updates the header with info about payload.
	return parcel
}
func NewParcelMsg(network NetworkID, payload []byte, msg interfaces.IMsg) *Parcel {
	header := new(ParcelHeader).Init(network)
	header.AppHash = "NetworkMessage"
	header.AppType = "Network"
	parcel := new(Parcel).Init(*header)
	parcel.Payload = payload
	parcel.Msg = msg      // Keep the message for debugging
	parcel.UpdateHeader() // Updates the header with info about payload.
	return parcel
}

func ParcelsForPayload(network NetworkID, payload []byte, msg interfaces.IMsg) []Parcel {
	parcelCount := (len(payload) / MaxPayloadSize) + 1
	parcels := make([]Parcel, parcelCount)

	for i := 0; i < parcelCount; i++ {
		start := i * MaxPayloadSize
		next := (i + 1) * MaxPayloadSize
		var end int
		if next < len(payload) {
			end = next
		} else {
			end = len(payload)
		}
		parcel := NewParcelMsg(network, payload[start:end], msg)
		parcel.Header.Type = TypeMessagePart
		parcel.Header.PartNo = uint16(i)
		parcel.Header.PartsTotal = uint16(parcelCount)
		parcels[i] = *parcel
	}
	return parcels
}

func ReassembleParcel(parcels []*Parcel) *Parcel {
	var payload bytes.Buffer

	for _, parcel := range parcels {
		payload.Write(parcel.Payload)
	}

	// create a new message parcel from the reassembled payload, but
	// copy all the relevant header fields from one of the original
	// messages
	origHeader := parcels[0].Header

	assembledParcel := NewParcel(origHeader.Network, payload.Bytes())
	assembledParcel.Header.NodeID = origHeader.NodeID
	assembledParcel.Header.Type = TypeMessage
	assembledParcel.Header.TargetPeer = origHeader.TargetPeer
	assembledParcel.Header.PeerAddress = origHeader.PeerAddress
	assembledParcel.Header.PeerPort = origHeader.PeerPort

	return assembledParcel
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
