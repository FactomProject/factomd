package p2p

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

// Handshake is an alias of V9MSG for backward compatibility
type Handshake V9Msg

// Valid checks if the other node is compatible
func (h *Handshake) Valid(conf *Configuration) error {
	if h.Header.Version < conf.ProtocolVersionMinimum {
		return fmt.Errorf("version %d is below the minimum", h.Header.Version)
	}

	if h.Header.Network != conf.Network {
		return fmt.Errorf("wrong network id %x", h.Header.Network)
	}

	if len(h.Payload) == 0 {
		return fmt.Errorf("zero-length payload")
	}

	if h.Header.Length != uint32(len(h.Payload)) {
		return fmt.Errorf("length in header does not match payload")
	}

	csum := crc32.Checksum(h.Payload, crcTable)
	if csum != h.Header.Crc32 {
		return fmt.Errorf("invalid checksum")
	}

	port, err := strconv.Atoi(h.Header.PeerPort)
	if err != nil {
		return fmt.Errorf("unable to parse port %s: %v", h.Header.PeerPort, err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("given port out of range: %d", port)
	}
	return nil
}

func newHandshake(conf *Configuration, payload []byte) *Handshake {
	hs := new(Handshake)
	hs.Header = V9Header{
		Network:  conf.Network,
		Version:  conf.ProtocolVersion,
		Type:     TypeHandshake,
		NodeID:   uint64(conf.NodeID),
		PeerPort: conf.ListenPort,
		AppHash:  "NetworkMessage",
		AppType:  "Network",
	}
	hs.SetPayload(payload)
	return hs
}

// SetPayload adds a payload to the handshake and updates the header with metadata
func (h *Handshake) SetPayload(payload []byte) {
	h.Payload = payload
	h.Header.Crc32 = crc32.Checksum(h.Payload, crcTable)
	h.Header.Length = uint32(len(h.Payload))
}
