// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
)

// Data structures and functions related to peers (eg other nodes in the network)

type Peer struct {
	QualityScore int    // 0 is neutral quality, negative is a bad peer.
	Address      string // Must be in form of x.x.x.x:p
	Hash         string
	Location     uint32 // IP address as an int.
	Type         uint8
}

const ( // iota is reset to 0
	RegularPeer uint8 = iota
	TrustedPeer
)

func (p *Peer) Init(address string, quality int) *Peer {
	p.address = address
	p.qualityScore = 0 // start at zero, zero is neutral, negative is a bad peer, positive is a good peer.
	p.Location = p.locationFromAddress()
	p.Hash = PeerHashFromAddress(address)
	return p
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

// locationFromAddress converts the peers address into a uint32 "location" numeric
func (p *Peer) locationFromAddress() uint32 {
	// Split out the port
	ip_port := strings.Split(p.Location, ":")
	// Split the IPv4 octets
	octets := strings.Split(ip_port[0], ".")
	// Turn into uint32
	var location uint32
	location += uint32(strconv.Atoi(octets[0])) << 24
	location += uint32(strconv.Atoi(octets[1])) << 16
	location += uint32(strconv.Atoi(octets[2])) << 8
	location += uint32(strconv.Atoi(octets[3]))
	return location
}

func PeerHashFromAddress(address string) string {
	raw := sha256.Sum256([]byte(p.address))
	hash = base64.URLEncoding.EncodeToString(raw[0:sha256.Size])
	return hash
}

// merit increases a peers reputation
func (p *Peer) merit() {
	p.qualityScore++
}

// demerit decreases a peers reputation
func (p *Peer) demerit() {
	p.qualityScore--
}
