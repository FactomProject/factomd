// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

var peerLogger = packageLogger.WithField("subpack", "peer")

// Data structures and functions related to peers (eg other nodes in the network)

// Keep a short history of messages
type Last100 struct {
	Msgs     map[[32]byte]bool // Look up messages by hash
	MsgOrder [100][32]byte     // keep a list of the order they were added
	N        int
}

func (l *Last100) Add(hash [32]byte) {
	if l.Msgs == nil {
		l.Msgs = make(map[[32]byte]bool, 0)
	}
	previous := l.MsgOrder[l.N] // get the oldest message
	delete(l.Msgs, previous)    // Delete it for the map
	l.MsgOrder[l.N] = hash      // replace it with the new message
	l.Msgs[hash] = true         // Add new the message to the map
	l.N = (l.N + 1) % 100       // move and wrap the index
}

//Check if we have a message in the short history
func (l *Last100) Get(hash [32]byte) bool {
	_, exists := l.Msgs[hash]
	return exists
}

type Peer struct {
	QualityScore int32     // 0 is neutral quality, negative is a bad peer.
	Address      string    // Must be in form of x.x.x.x
	Port         string    // Must be in form of xxxx
	NodeID       uint64    // a nonce to distinguish multiple nodes behind one IP address
	Hash         string    // This is more of a connection ID than hash right now.
	Location     uint32    // IP address as an int.
	Network      NetworkID // The network this peer reference lives on.
	Type         uint8
	Connections  int                  // Number of successful connections.
	LastContact  time.Time            // Keep track of how long ago we talked to the peer.
	Source       map[string]time.Time // source where we heard from the peer.

	// logging
	logger *log.Entry

	PrevMsgs Last100 `json:"-"`
}

const (
	RegularPeer        uint8 = iota
	SpecialPeerConfig        // special peer defined in the config file
	SpecialPeerCmdLine       // special peer defined via the cmd line params
)

func (p *Peer) Init(address string, port string, quality int32, peerType uint8, connections int) *Peer {

	p.logger = peerLogger.WithFields(log.Fields{
		"address":  address,
		"port":     port,
		"peerType": peerType,
	})
	if net.ParseIP(address) == nil {
		ipAddress, err := net.LookupHost(address)
		if err != nil {
			p.logger.Errorf("Init: LookupHost(%v) failed. %v ", address, err)
			// is there a way to abandon this peer at this point? -- clay
		} else {
			address = ipAddress[0]
		}
	}

	p.Address = address
	p.Port = port
	p.QualityScore = quality
	p.generatePeerHash()
	p.Type = peerType
	p.Location = p.LocationFromAddress()
	p.Source = map[string]time.Time{}
	p.Network = CurrentNetwork
	return p
}

func (p *Peer) generatePeerHash() {
	p.Hash = fmt.Sprintf("%s:%s %x", p.Address, p.Port, rand.Int63())
}

func (p *Peer) AddressPort() string {
	return p.Address + ":" + p.Port
}

func (p *Peer) PeerIdent() string {
	return p.Hash[0:12] + "-" + p.Address + ":" + p.Port
}

func (p *Peer) PeerFixedIdent() string {
	address := fmt.Sprintf("%16s", p.Address)
	return p.Hash[0:12] + "-" + address + ":" + p.Port
}

func (p *Peer) PeerLogFields() log.Fields {
	return log.Fields{
		"address":   p.Address,
		"port":      p.Port,
		"peer_type": p.PeerTypeString(),
	}
}

// gets the last source where this peer was seen
func (p *Peer) LastSource() (result string) {
	var maxTime time.Time

	for source, lastSeen := range p.Source {
		if lastSeen.After(maxTime) {
			maxTime = lastSeen
			result = source
		}
	}

	return
}

// TODO Hadn't considered IPV6 address support.
// TODO Need to audit all the net code to check IPv6 addresses
// Here's an IPv6 conversion:
// Ref: http://stackoverflow.com/questions/23297141/golang-net-ip-to-ipv6-from-mysql-as-decimal39-0-conversion
// func ipv6ToInt(IPv6Addr net.IP) *big.Int {
//     IPv6Int := big.NewInt(0)
//     IPv6Int.SetBytes(IPv6Addr)
//     return IPv6Int
// }
// Problem is we're working with string addresses, may never have made a connection.
// TODO - we might have a DNS address, not iP address and need to resolve it!
// locationFromAddress converts the peers address into a uint32 "location" numeric
func (p *Peer) LocationFromAddress() (location uint32) {
	location = 0
	// Split the IPv4 octets
	ip := net.ParseIP(p.Address)
	if ip == nil {
		ipAddress, err := net.LookupHost(p.Address)
		if err != nil {
			p.logger.Debugf("LocationFromAddress(%v) failed. %v ", p.Address, err)
			p.logger.Debugf("Invalid Peer Address: %v", p.Address)
			p.logger.Debugf("Peer: %s has Location: %d", p.Hash, location)
			return 0 // We use location on 0 to say invalid
		}
		p.Address = ipAddress[0]
		ip = net.ParseIP(p.Address)
	}
	if len(ip) == 16 { // If we got back an IP6 (16 byte) address, use the last 4 byte
		ip = ip[12:]
	}
	// Turn into uint32
	location += uint32(ip[0]) << 24
	location += uint32(ip[1]) << 16
	location += uint32(ip[2]) << 8
	location += uint32(ip[3])
	p.logger.Debugf("Peer: %s has Location: %d", p.Hash, location)
	return location
}

func (p *Peer) IsSamePeerAs(netAddress net.Addr) bool {
	address, _, err := net.SplitHostPort(netAddress.String())
	if err != nil {
		return false
	}
	return address == p.Address
}

// merit increases a peers reputation
func (p *Peer) merit() {
	if 2147483000 > p.QualityScore {
		p.QualityScore++
	}
}

// demerit decreases a peers reputation
func (p *Peer) demerit() {
	if -2147483000 < p.QualityScore {
		//p.QualityScore--
	}
}

func (p *Peer) IsSpecial() bool {
	return p.Type == SpecialPeerConfig || p.Type == SpecialPeerCmdLine
}

func (p *Peer) PeerTypeString() string {
	switch p.Type {
	case RegularPeer:
		return "regular"
	case SpecialPeerConfig:
		return "special_config"
	case SpecialPeerCmdLine:
		return "special_cmdline"
	default:
		return "unknown"
	}
}

// sort.Sort interface implementation
type PeerQualitySort []Peer

func (p PeerQualitySort) Len() int {
	return len(p)
}
func (p PeerQualitySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p PeerQualitySort) Less(i, j int) bool {
	return p[i].QualityScore < p[j].QualityScore
}

// sort.Sort interface implementation
type PeerDistanceSort []Peer

func (p PeerDistanceSort) Len() int {
	return len(p)
}
func (p PeerDistanceSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p PeerDistanceSort) Less(i, j int) bool {
	return p[i].Location < p[j].Location
}
