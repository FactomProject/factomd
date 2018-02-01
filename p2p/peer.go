// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"math/rand"
	"time"
	"net"
)

// Data structures and functions related to peers (eg other nodes in the network)

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
}

const ( // iota is reset to 0
	RegularPeer uint8 = iota
	SpecialPeer
)

func (p *Peer) Init(address string, port string, quality int32, peerType uint8, connections int) *Peer {

	if net.ParseIP(address) == nil {
		ipAddress, err := net.LookupHost(address)
		if err != nil {
			verbose("peer", "Init: LookupHost(%v) failed. %v ", address, err)
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
	ip:= net.ParseIP(p.Address)
	if ip == nil {
		ipAddress, err := net.LookupHost(p.Address)
		if err != nil {
			verbose("peer", "LocationFromAddress(%v) failed. %v ", p.Address, err)
			silence("peer", "Invalid Peer Address: %v", p.Address)
			verbose("peer", "Peer: %s has Location: %d", p.Hash, location)
			return 0 // We use location on 0 to say invalid
		}
		p.Address = ipAddress[0]
		ip = net.ParseIP(p.Address)
	}
	if(len(ip) == 16) { // If we got back an IP6 (16 byte) address, use the last 4 byte
		ip = ip[12:]
	}
		// Turn into uint32
		location += uint32(ip[0]) << 24
		location += uint32(ip[1]) << 16
		location += uint32(ip[2]) << 8
		location += uint32(ip[3])
		verbose("peer", "Peer: %s has Location: %d", p.Hash, location)
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
		//p.QualityScore--
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
