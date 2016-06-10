// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// Data structures and functions related to peers (eg other nodes in the network)

type Peer struct {
	QualityScore int32  // 0 is neutral quality, negative is a bad peer.
	Address      string // Must be in form of x.x.x.x
	Port         string // Must be in form of xxxx
	NodeID       uint64 // a nonce to distinguish multiple nodes behind one IP address
	Hash         string
	Location     uint32 // IP address as an int.
	Type         uint8
	Connections  int       // Number of successful connections.
	LastContact  time.Time // Keep track of how long ago we talked to the peer.
}

const ( // iota is reset to 0
	RegularPeer uint8 = iota
	SpecialPeer
)

func (p *Peer) Init(address string, port string, quality int32, peerType uint8, connections int) *Peer {
	p.Address = address
	p.Port = port
	p.QualityScore = quality
	p.generatePeerHash()
	p.Type = peerType
	p.Location = p.locationFromAddress()
	return p
}

func (p *Peer) generatePeerHash() {
	buff := make([]byte, 256)
	RandomGenerator.Read(buff)
	raw := sha256.Sum256(buff)
	p.Hash = base64.URLEncoding.EncodeToString(raw[0:sha256.Size])
}

func (p *Peer) AddressPort() string {
	return p.Address + ":" + p.Port
}

func (p *Peer) PeerIdent() string {
	return p.Hash[0:12] + "-" + p.Address + ":" + p.Port
}

// BUGBUG Hadn't considered IPV6 addresses.
// BUGBUG Need to audit all the net code to check IPv6 addresses
// Here's an IPv6 conversion:
// Ref: http://stackoverflow.com/questions/23297141/golang-net-ip-to-ipv6-from-mysql-as-decimal39-0-conversion
// func ipv6ToInt(IPv6Addr net.IP) *big.Int {
//     IPv6Int := big.NewInt(0)
//     IPv6Int.SetBytes(IPv6Addr)
//     return IPv6Int
// }
// Problem is we're working wiht string addresses, may never have made a connection.
// BUGBUG - we might have a DNS address, not iP address and need to resolve it!
// locationFromAddress converts the peers address into a uint32 "location" numeric
func (p *Peer) locationFromAddress() uint32 {
	// Split out the port
	ip_port := strings.Split(p.Address, ":")
	// Split the IPv4 octets
	octets := strings.Split(ip_port[0], ".")
	// Turn into uint32
	var location uint32
	b0, _ := strconv.Atoi(octets[0])
	b1, _ := strconv.Atoi(octets[1])
	b2, _ := strconv.Atoi(octets[2])
	b3, _ := strconv.Atoi(octets[3])
	location += uint32(b0) << 24
	location += uint32(b1) << 16
	location += uint32(b2) << 8
	location += uint32(b3)
	verbose("peer", "Peer: %s with ip_port: %+v and octets: %+v has Location: %d", p.Hash, ip_port, octets, location)
	return location
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
		p.QualityScore--
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
