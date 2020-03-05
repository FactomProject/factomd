package p2p

import (
	"time"
)

// PeerMetrics is the data shared to the metrics hook
type PeerMetrics struct {
	Hash             string
	PeerAddress      string
	MomentConnected  time.Time
	PeerQuality      int32
	LastReceive      time.Time
	LastSend         time.Time
	MessagesSent     uint64
	BytesSent        uint64
	MessagesReceived uint64
	BytesReceived    uint64
	Incoming         bool
	PeerType         string
	ConnectionState  string
	MPSDown          float64
	MPSUp            float64
	BPSDown          float64
	BPSUp            float64
	SendFillRatio    float64
	Dropped          uint64
}

// peerStatus is an indicator for peer manager whether the associated peer is going online or offline
type peerStatus struct {
	peer   *Peer
	online bool
}

type peerParcel struct {
	peer   *Peer
	parcel *Parcel
}
