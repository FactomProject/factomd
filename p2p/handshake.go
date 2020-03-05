package p2p

import (
	"fmt"
	"strconv"
)

// Handshake is the protocol independent data that is required to authenticate a peer.
type Handshake struct {
	Network      NetworkID
	Version      uint16
	Type         ParcelType
	NodeID       uint32
	ListenPort   string
	Loopback     uint64
	Alternatives []Endpoint
}

// Valid checks the Handshake's data against a configuration.
// Loopback is checked outside of this function.
func (h *Handshake) Valid(conf *Configuration) error {
	if h.Version < conf.ProtocolVersionMinimum {
		return fmt.Errorf("version %d is below the minimum", h.Version)
	}

	if h.Network != conf.Network {
		return fmt.Errorf("wrong network id %x", h.Network)
	}

	port, err := strconv.Atoi(h.ListenPort)
	if err != nil {
		return fmt.Errorf("unable to parse port %s: %v", h.ListenPort, err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("given port out of range: %d", port)
	}

	for _, ep := range h.Alternatives {
		if !ep.Valid() {
			return fmt.Errorf("invalid list of alternatives provided")
		}
	}

	return nil
}

// create a new handshake
func newHandshake(conf *Configuration, loopback uint64) *Handshake {
	hs := new(Handshake)
	hs.Type = TypeHandshake
	hs.Network = conf.Network
	hs.Version = conf.ProtocolVersion
	hs.NodeID = conf.NodeID
	hs.ListenPort = conf.ListenPort
	hs.Loopback = loopback
	return hs
}
