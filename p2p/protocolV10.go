package p2p

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
)

var _ Protocol = (*ProtocolV10)(nil)

// ProtocolV10 is the protocol introduced by p2p 2.0.
// It is a slimmed down version of V9, reducing overhead
type ProtocolV10 struct {
	decoder *gob.Decoder
	encoder *gob.Encoder
}

// V10Msg is the barebone message
type V10Msg struct {
	Type    ParcelType
	Payload []byte
}

func newProtocolV10(decoder *gob.Decoder, encoder *gob.Encoder) *ProtocolV10 {
	v10 := new(ProtocolV10)
	v10.decoder = decoder
	v10.encoder = encoder
	return v10
}

func (v10 *ProtocolV10) SendHandshake(h *Handshake) error {
	return v9SendHandshake(v10.encoder, h)
}

// ReadHandshake for v10 is using the identical format to V9 for backward compatibility.
// It can't be easily told apart without first decoding the message, so the code is only
// implemented in v9, then upgraded to V10 based on the values
func (v10 *ProtocolV10) ReadHandshake() (*Handshake, error) {
	return nil, fmt.Errorf("V10 doesn't have its own handshake")
}

// Send encodes a Parcel as V10Msg, calculates the crc and encodes it as gob
func (v10 *ProtocolV10) Send(p *Parcel) error {
	var msg V10Msg
	msg.Type = p.ptype
	msg.Payload = p.Payload
	return v10.encoder.Encode(msg)
}

// Version 10
func (v10 *ProtocolV10) Version() uint16 {
	return 10
}

// String 10
func (v10 *ProtocolV10) String() string {
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
		return nil, fmt.Errorf("null payload")
	}

	p := newParcel(msg.Type, msg.Payload)
	return p, nil
}

// V10Share is an alias of Endpoint
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
