package p2p

import (
	"time"
)

const (
	// Broadcast sends a parcel to multiple peers (randomly selected based on fanout and special peers)
	Broadcast = "<BROADCAST>"
	// FullBroadcast sends a parcel to all peers
	FullBroadcast = "<FULLBORADCAST>"
	// RandomPeer sends a parcel to one randomly selected peer
	RandomPeer = "<RANDOMPEER>"
)

// Configuration defines the behavior of the gossip network protocol
type Configuration struct {
	// Network is the NetworkID of the network to use, eg. MainNet, TestNet, etc
	Network NetworkID

	// NodeID is this node's id
	NodeID uint32
	// NodeName is the internal name of the node
	NodeName string

	// === Peer Management Settings ===
	// PeerRequestInterval dictates how often neighbors should be asked for an
	// updated peer list
	PeerRequestInterval time.Duration
	// PeerReseedInterval dictates how often the seed file should be accessed
	// to check for changes
	PeerReseedInterval time.Duration
	// PeerIPLimit specifies the maximum amount of peers to accept from a single
	// ip address
	// 0 for unlimited
	PeerIPLimitIncoming uint
	PeerIPLimitOutgoing uint

	// Special is a list of special peers, separated by comma. If no port is specified, the entire
	// ip is considered special
	Special string

	// PeerCacheFile is the filepath to the file to save peers. It is persisted in every CAT round
	PeerCacheFile string
	// PeerCacheAge is the maximum age of the peer file to try and bootstrap peers from
	PeerCacheAge time.Duration

	// to count as being connected
	// PeerShareAmount is the number of peers we share
	PeerShareAmount uint

	// CAT Settings
	// How often to do cat rounds
	RoundTime time.Duration
	// Desired amount of peers
	TargetPeers uint
	// Hard cap of connections
	MaxPeers uint
	// Amount of peers to drop down to
	DropTo uint
	// Reseed if there are fewer than this peers
	MinReseed   uint
	MaxIncoming uint // maximum inbound connections, 0 <= MaxIncoming <= MaxPeers

	// === Gossip Behavior ===

	// Fanout controls how many random peers are selected for propagating messages
	// Higher values increase fault tolerance but also increase network congestion
	Fanout uint

	// SeedURL is the URL of the remote seed file
	SeedURL string // URL to a source of peer info

	// === Connection Settings ===

	// BindIP is the ip address to bind to for listening and connecting
	//
	// leave blank to bind to all
	BindIP string
	// ListenPort is the port to listen to incoming tcp connections on
	ListenPort string
	// ListenLimit is the lockout period of accepting connections from a single
	// ip after having a successful connection from that ip
	ListenLimit time.Duration

	// PingInterval dictates the maximum amount of time a connection can be
	// silent (no writes) before sending a Ping
	PingInterval time.Duration

	// RedialInterval dictates how long to wait between connection attempts
	RedialInterval time.Duration

	// ManualBan is the duration to ban an address for when banned manually
	ManualBan time.Duration

	// HandshakeDeadline is the maximum acceptable time for an incoming conneciton
	// to send the first parcel after connecting
	HandshakeTimeout time.Duration
	DialTimeout      time.Duration

	// ReadDeadline is the maximum acceptable time to read a single parcel
	// if a connection takes longer, it is disconnected
	ReadDeadline time.Duration

	// WriteDeadline is the maximum acceptable time to send a single parcel
	// if a connection takes longer, it is disconnected
	WriteDeadline time.Duration

	ProtocolVersion uint16
	// ProtocolVersionMinimum is the earliest version this package supports
	ProtocolVersionMinimum uint16

	// ChannelCapacity dictates how large each peer's send channel is.
	// Should be large enough to accomodate bursts of traffic.
	ChannelCapacity uint

	EnablePrometheus bool // Enable prometheus logging. Disable if you run multiple instances
}

// DefaultP2PConfiguration returns a network configuration with base values
// These should be overwritten with command line and config parameters
func DefaultP2PConfiguration() (c Configuration) {
	c.Network = MainNet
	c.NodeID = 0
	c.NodeName = "FNode0"
	c.ListenPort = "8108"

	c.PeerRequestInterval = time.Second * 5
	c.PeerReseedInterval = time.Hour * 4
	c.PeerIPLimitIncoming = 0
	c.PeerIPLimitOutgoing = 0
	c.ManualBan = time.Hour * 24 * 7 // a week

	c.PeerCacheFile = ""
	c.PeerCacheAge = time.Hour //

	c.MaxIncoming = 36
	c.Fanout = 8
	c.PeerShareAmount = 3 // CAT share
	c.RoundTime = time.Minute * 15
	c.TargetPeers = 32
	c.MaxPeers = 36
	c.DropTo = 30
	c.MinReseed = 10

	c.BindIP = "" // bind to all
	c.ListenPort = "8108"
	c.ListenLimit = time.Second
	c.PingInterval = time.Second * 15
	c.RedialInterval = time.Minute * 2

	c.ReadDeadline = time.Minute * 5     // high enough to accomodate large packets
	c.WriteDeadline = time.Minute * 5    // but fail eventually
	c.HandshakeTimeout = time.Second * 5 // can be quite low
	c.DialTimeout = time.Second * 5      // can be quite low

	c.ProtocolVersion = 10
	c.ProtocolVersionMinimum = 9

	c.ChannelCapacity = 1000

	c.EnablePrometheus = true
	return
}

// Sanitize automatically adjusts some variables that are dependent on others
func (c *Configuration) Sanitize() {
	if c.MaxIncoming > c.MaxPeers {
		c.MaxIncoming = c.MaxPeers
	}
}
