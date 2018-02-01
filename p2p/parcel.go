// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"strconv"
	"time"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
)

// Parcel is the atomic level of communication for the p2p network.  It contains within it the necessary info for
// the networking protocol, plus the message that the Application is sending.
type Parcel struct {
	Header  ParcelHeader
	Payload []byte
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

func ParcelsForPayload(network NetworkID, payload []byte) []Parcel {
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
		parcel := NewParcel(network, payload[start:end])
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

func (p *Parcel) Trace(location string, sequence string) {
	if 10 < CurrentLoggingLevel { // lower level means more severe. "Silence" level always printed, overriding silence.
		time := time.Now().Unix()
		fmt.Printf("\nParcelTrace, %s, %s, %s, %s, %s, %d \n", p.Header.AppHash, sequence, p.Header.AppType, CommandStrings[p.Header.Type], location, time)
	}
}

func (p *ParcelHeader) Print() {
	// debug( true, "\t Cookie: \t%+v", string(p.Cookie))
	debug("parcel", "\t Network:\t%+v", p.Network.String())
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
	output = fmt.Sprintf("%s\t Network:\t%+v\n", output, p.Header.Network.String())
	output = fmt.Sprintf("%s\t Version:\t%+v\n", output, p.Header.Version)
	output = fmt.Sprintf("%s\t Type:   \t%+v\n", output, CommandStrings[p.Header.Type])
	output = fmt.Sprintf("%s\t Length:\t%d\n", output, p.Header.Length)
	output = fmt.Sprintf("%s\t TargetPeer:\t%s\n", output, p.Header.TargetPeer)
	output = fmt.Sprintf("%s\t CRC32:\t%d\n", output, p.Header.Crc32)
	output = fmt.Sprintf("%s\t PartNo:\t%d\n", output, p.Header.PartNo)
	output = fmt.Sprintf("%s\t PartsTotal:\t%d\n", output, p.Header.PartsTotal)
	output = fmt.Sprintf("%s\t NodeID:\t%d\n", output, p.Header.NodeID)
	output = fmt.Sprintf("%s\t Payload: %s\n", output, s)
	return output
}

func Parcel2String(p *Parcel) string {
	extraInfo :=""
	r:=""
	switch p.Header.Type {
	case TypeHeartbeat,   // "Note, I'm still alive"
		TypePing,         // "Are you there?"
		TypePong,         // "yes, I'm here"
		TypePeerRequest,  // "Please share some peers"
		TypePeerResponse, // "Here's some peers I know about."
		TypeAlert:        // network wide alerts (used in bitcoin to indicate criticalities)
		r = fmt.Sprintf("%s %26s[%2v]", p.Header.AppHash[:8], CommandStrings[p.Header.Type], p.Header.Type)

	case TypeMessage: // Application level message
		msgType, _ := strconv.Atoi(p.Header.AppType)
		switch byte(msgType) {
		case constants.EOM_MSG:
		case constants.ACK_MSG:
		case constants.AUDIT_SERVER_FAULT_MSG:
		case constants.FED_SERVER_FAULT_MSG:
		case constants.FULL_SERVER_FAULT_MSG:
		case constants.COMMIT_CHAIN_MSG:
		case constants.COMMIT_ENTRY_MSG:
		case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		case constants.EOM_TIMEOUT_MSG:
		case constants.FACTOID_TRANSACTION_MSG:
		case constants.HEARTBEAT_MSG:
		case constants.INVALID_ACK_MSG:
		case constants.INVALID_DIRECTORY_BLOCK_MSG:
		case constants.MISSING_MSG:
		case constants.MISSING_MSG_RESPONSE:
		case constants.MISSING_DATA:
		case constants.DATA_RESPONSE:
		case constants.REVEAL_ENTRY_MSG:
		case constants.REQUEST_BLOCK_MSG:
		case constants.SIGNATURE_TIMEOUT_MSG:
		case constants.DBSTATE_MISSING_MSG:
		case constants.DBSTATE_MSG:
		case constants.BOUNCE_MSG:
		case constants.BOUNCEREPLY_MSG:
		default:
		}
		r = fmt.Sprintf("%s %26s[%2v]", p.Header.AppHash[:8], messages.MessageName(byte(msgType)))

	case TypeMessagePart:                   // Application level message that was split into multiple parts
		debug("Parcel2String","unexpected TypeMessagePart") // we don't do this at this point
	default:
		r = fmt.Sprintf("%s %26s[%2v]", "UnknownParcelType", p.Header.AppHash[:8],  p.Header.Type)
	}

	return r+extraInfo
}
