// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"crypto/sha256"
	"encoding/base64"
)

// Data structures and functions related to peers (eg other nodes in the network)

type Peer struct {
	QualityScore int    // 0 is neutral quality, negative is a bad peer.
	Address      string // Must be in form of x.x.x.x:p
	Hash         string
	Location     uint32 // IP address as an int.
}

func (p *Peer) Init(address string, quality int) *Peer {
	p.address = address
	p.qualityScore = 0 // start at zero, zero is neutral, negative is a bad peer, positive is a good peer.
	// p.Location = p.locationFromAddress()
	p.Hash = PeerHashFromAddress(address)
	return p
}

// BUGBUG Hadn't considered IPV6 addresses.

// locationFromAddress converts the peers address into a uint32 "location" numeric
// func (p *Peer) locationFromAddress() uint32 {
// 	p.qualityScore++
// }

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
