package p2p

import (
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"time"
)

var _ Protocol = (*ProtocolV9)(nil)

// ProtocolV9 is the legacy format of the old p2p package which sends Parcels
// over the wire using gob. The V9Msg struct is equivalent to the old package's
// "Parcel" and "ParcelHeader" structure
type ProtocolV9 struct {
	network NetworkID
	nodeID  uint32
	port    string
	decoder *gob.Decoder
	encoder *gob.Encoder
}
type V9Handshake V9Msg

func newProtocolV9(netw NetworkID, nodeID uint32, listenPort string, decoder *gob.Decoder, encoder *gob.Encoder) *ProtocolV9 {
	v9 := new(ProtocolV9)
	v9.network = netw
	v9.nodeID = nodeID
	v9.port = listenPort
	v9.decoder = decoder
	v9.encoder = encoder
	return v9
}

func v9SendHandshake(encoder *gob.Encoder, h *Handshake) error {
	var payload []byte
	if len(h.Alternatives) > 0 {
		if data, err := json.Marshal(h.Alternatives); err != nil {
			return err
		} else {
			payload = data
		}
	} else {
		payload = make([]byte, 8)
		binary.LittleEndian.PutUint64(payload, h.Loopback)
	}

	var msg V9Handshake
	msg.Header.Network = h.Network
	msg.Header.Version = h.Version // can be 9 or 10
	msg.Header.Type = h.Type
	msg.Header.TargetPeer = ""

	msg.Header.NodeID = uint64(h.NodeID)
	msg.Header.PeerAddress = ""
	msg.Header.PeerPort = h.ListenPort
	msg.Header.AppHash = "NetworkMessage"
	msg.Header.AppType = "Network"

	msg.Payload = payload
	msg.Header.Crc32 = crc32.Checksum(msg.Payload, crcTable)
	msg.Header.Length = uint32(len(msg.Payload))

	return encoder.Encode(msg)
}

// SendHandshake sends out a v9 structured handshake
// transform handshake into peer request
func (v9 *ProtocolV9) SendHandshake(h *Handshake) error {
	if h.Type == TypeHandshake {
		h.Type = TypePeerRequest
	}
	return v9SendHandshake(v9.encoder, h)
}

func (v9 *ProtocolV9) ReadHandshake() (*Handshake, error) {
	msg, err := v9.read()
	if err != nil {
		return nil, err
	}

	hs := new(Handshake)
	hs.Type = msg.Header.Type

	if msg.Header.Type == TypeRejectAlternative {
		var alternatives []Endpoint
		if err = json.Unmarshal(msg.Payload, &alternatives); err != nil {
			return nil, err
		}
		hs.Alternatives = alternatives
	} else if len(msg.Payload) == 8 {
		hs.Loopback = binary.LittleEndian.Uint64(msg.Payload)
	}

	hs.ListenPort = msg.Header.PeerPort
	hs.Network = msg.Header.Network
	hs.NodeID = uint32(msg.Header.NodeID)
	hs.Version = msg.Header.Version

	return hs, nil
}

// Send a parcel over the connection
func (v9 *ProtocolV9) Send(p *Parcel) error {
	var msg V9Msg
	msg.Header.Network = v9.network
	msg.Header.Version = 9 // hardcoded
	msg.Header.Type = p.Type
	msg.Header.TargetPeer = p.Address

	msg.Header.NodeID = uint64(v9.nodeID)
	msg.Header.PeerAddress = ""
	msg.Header.PeerPort = v9.port
	msg.Header.AppHash = "NetworkMessage"
	msg.Header.AppType = "Network"

	msg.Payload = p.Payload
	msg.Header.Crc32 = crc32.Checksum(p.Payload, crcTable)
	msg.Header.Length = uint32(len(p.Payload))

	return v9.encoder.Encode(msg)
}

func (v9 *ProtocolV9) read() (*V9Msg, error) {
	var msg V9Msg
	err := v9.decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

// Receive a parcel from the network. Blocking.
func (v9 *ProtocolV9) Receive() (*Parcel, error) {
	msg, err := v9.read()
	if err != nil {
		return nil, err
	}

	if err = msg.Valid(); err != nil {
		return nil, err
	}

	p := new(Parcel)
	p.Address = msg.Header.TargetPeer
	p.Payload = msg.Payload
	p.Type = msg.Header.Type
	return p, nil
}

// Version of the protocol
func (v9 *ProtocolV9) Version() uint16 {
	return 9
}

func (v9 *ProtocolV9) String() string { return "9" }

// V9Msg is the legacy format of protocol 9
type V9Msg struct {
	Header  V9Header
	Payload []byte
}

// V9Header carries meta information about the parcel
type V9Header struct {
	Network     NetworkID
	Version     uint16
	Type        ParcelType
	Length      uint32
	TargetPeer  string
	Crc32       uint32
	PartNo      uint16
	PartsTotal  uint16
	NodeID      uint64
	PeerAddress string
	PeerPort    string
	AppHash     string
	AppType     string
}

// Valid checks header for inconsistencies
func (msg V9Msg) Valid() error {
	if msg.Header.Version != 9 {
		return fmt.Errorf("invalid version %v", msg.Header)
	}

	if len(msg.Payload) == 0 {
		return fmt.Errorf("zero-length payload")
	}

	if msg.Header.Length != uint32(len(msg.Payload)) {
		return fmt.Errorf("length in header does not match payload")
	}

	csum := crc32.Checksum(msg.Payload, crcTable)
	if csum != msg.Header.Crc32 {
		return fmt.Errorf("invalid checksum")
	}

	return nil
}

// V9Share is the legacy code's "Peer" struct. Resets QualityScore and Source list when
// decoding, filters out wrong Networks
type V9Share struct {
	QualityScore int32
	Address      string
	Port         string
	NodeID       uint64
	Hash         string
	Location     uint32
	Network      NetworkID
	Type         uint8
	Connections  int
	LastContact  time.Time
	Source       map[string]time.Time
}

// MakePeerShare serializes the given endpoints to a V9Share encoded in json
func (v9 *ProtocolV9) MakePeerShare(ps []Endpoint) ([]byte, error) {
	var conv []V9Share
	src := make(map[string]time.Time)
	for _, ep := range ps {
		loc := IP2LocationQuick(ep.IP)
		conv = append(conv, V9Share{
			Address:      ep.IP,
			Port:         ep.Port,
			QualityScore: 20,
			NodeID:       1,
			Hash:         ep.IP,
			Location:     loc,
			Network:      v9.network,
			Type:         0,
			Connections:  1,
			LastContact:  time.Time{},
			Source:       src,
		})
	}

	return json.Marshal(conv)
}

// ParsePeerShare unserializes the json V9Share
func (v9 *ProtocolV9) ParsePeerShare(payload []byte) ([]Endpoint, error) {
	var list []V9Share

	err := json.Unmarshal(payload, &list)
	if err != nil {
		return nil, err
	}

	var conv []Endpoint
	for _, s := range list {
		conv = append(conv, Endpoint{
			IP:   s.Address,
			Port: s.Port,
		})
	}
	return conv, nil
}
