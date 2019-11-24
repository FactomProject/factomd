package p2p

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
)

var _ Protocol = (*ProtocolV10)(nil)

// ProtocolV10 is the protocol introduced by p2p 2.0.
// It is a slimmed down version of V9, reducing overhead
type ProtocolV10 struct {
	net     *Network
	decoder *gob.Decoder
	encoder *gob.Encoder
	peer    *Peer
}

// V10Msg is the barebone message
type V10Msg struct {
	Type    ParcelType
	Crc32   uint32
	Payload []byte
}

func (v10 *ProtocolV10) init(peer *Peer, decoder *gob.Decoder, encoder *gob.Encoder) {
	v10.peer = peer
	v10.net = peer.net
	v10.decoder = decoder
	v10.encoder = encoder
}

// Send encodes a Parcel as V10Msg, calculates the crc and encodes it as gob
func (v10 *ProtocolV10) Send(p *Parcel) error {
	var msg V10Msg
	msg.Type = p.Type
	msg.Crc32 = crc32.Checksum(p.Payload, crcTable)
	msg.Payload = p.Payload
	return v10.encoder.Encode(msg)
}

// Version 10
func (v10 *ProtocolV10) Version() string {
	return "10"
}

// Receive converts a V10Msg back to a Parcel
func (v10 *ProtocolV10) Receive() (*Parcel, error) {
	var msg V10Msg
	err := v10.decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}

	if len(msg.Payload) == 0 {
		return nil, fmt.Errorf("nul payload")
	}

	csum := crc32.Checksum(msg.Payload, crcTable)
	if csum != msg.Crc32 {
		return nil, fmt.Errorf("invalid checksum")
	}

	p := newParcel(msg.Type, msg.Payload)
	return p, nil
}

// V10Share is an alias of PeerShare
type V10Share Endpoint

// MakePeerShare serializes a list of ips via json
func (v10 *ProtocolV10) MakePeerShare(share []Endpoint) ([]byte, error) {
	var peershare []V10Share
	for _, ep := range share {
		peershare = append(peershare, V10Share{IP: ep.IP, Port: ep.Port})
	}
	return json.Marshal(peershare)
}

// ParsePeerShare parses a peer share payload
func (v10 *ProtocolV10) ParsePeerShare(payload []byte) ([]Endpoint, error) {
	var share []Endpoint
	err := json.Unmarshal(payload, &share)
	return share, err
}
