package p2p

import (
	"fmt"
	"hash/crc32"
)

var (
	crcTable = crc32.MakeTable(crc32.Koopman)
)

// Parcel is the raw data interface between the network, the p2p package, and the application.
//
// Type indicates the network or application type. Messages routed to and from the application
// will only have application types
//
// Address is a unique internal identifier for origin or target of the parcel. For messages from the
// network to the application, the address will the id of the sender. Messages intended to be
// returned to the sender should bear the same address.
//
// There are three special address constants:
// 		Broadcast: The message will be sent to multiple peers as specified by the fanout
//		FullBroadcast: The message will be sent to all peers
//		RandomPeer: The message will be sent to one peer picked at random
//
// The payload is arbitrary data defined at application level
type Parcel struct {
	ptype   ParcelType // 2 bytes - network level commands (eg: ping/pong)
	Address string     // ? bytes - "" or nil for broadcast, otherwise the destination peer's hash.
	Payload []byte
}

// IsApplicationMessage checks if the message is intended for the application
func (p *Parcel) IsApplicationMessage() bool {
	switch p.ptype {
	case TypeMessage, TypeMessagePart:
		return true
	default:
		return false
	}
}

func (p *Parcel) String() string {
	return fmt.Sprintf("[%s] %dB", p.ptype, len(p.Payload))
}

// NewParcel creates a new application message. The target should be either an identifier
// from a previous message, or one of the custom flags: Broadcast, BroadcastFull, RandomPeer
func NewParcel(target string, payload []byte) *Parcel {
	p := newParcel(TypeMessage, payload)
	p.Address = target
	return p
}

func newParcel(command ParcelType, payload []byte) *Parcel {
	parcel := new(Parcel)
	parcel.ptype = command
	parcel.Payload = payload
	return parcel
}

// Valid checks header for inconsistencies
func (p *Parcel) Valid() error {
	if p == nil {
		return fmt.Errorf("nil parcel")
	}

	if p.ptype >= ParcelType(len(typeStrings)) {
		return fmt.Errorf("unknown parcel type %d", p.ptype)
	}

	if len(p.Payload) == 0 {
		return fmt.Errorf("zero-length payload")
	}

	return nil
}
