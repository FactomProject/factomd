// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"bufio"
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Discovery struct {
	knownPeers map[string]Peer // peers we know about indexed by hash

	peersFilePath string     // the path to the peers.
	lastPeerSave  time.Time  // Last time we saved known peers.
	rng           *rand.Rand // RNG = random number generator
}

var UpdateKnownPeers sync.Mutex

// Discovery provides the code for sharing and managing peers,
// namely keeping track of all the peers we know about (not just the ones
// we are connected to.)  The discovery "service" is owned by the
// Controller and its routines are called from the Controllers runloop()
// This ensures that all shared memory is accessed from that goroutine.

func (d *Discovery) Init(peersFile string) *Discovery {
	UpdateKnownPeers.Lock()
	d.knownPeers = map[string]Peer{}
	UpdateKnownPeers.Unlock()
	d.peersFilePath = peersFile
	d.LoadPeers()
	d.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	return d
}

// Only controller should be able to read this, but we still got
// a concurrent read/write error, so isolating changes to knownPeers

// UpdatePeer updates the values in our known peers. Creates peer if its not in there.
func (d *Discovery) updatePeer(peer Peer) {
	UpdateKnownPeers.Lock()
	d.knownPeers[peer.Hash] = peer
	UpdateKnownPeers.Unlock()
}

// UpdatePeer updates the values in our known peers. Creates peer if its not in there.
func (d *Discovery) isPeerPresent(peer Peer) bool {
	UpdateKnownPeers.Lock()
	_, present := d.knownPeers[peer.Hash]
	UpdateKnownPeers.Unlock()
	return present
}

// GetFullPeer looks for a peer in the known peers, and if so, returns it  (based on
// the hash of the passed in peer.)  If the peer is unknown , we create it and
// add it to the known peers.
// func (d *Discovery) GetFullPeer(prototype Peer) Peer {
// 	return d.GetPeerByAddress(prototype.Address)
// }

func (d *Discovery) GetPeerByAddress(address string) Peer {
	hash := PeerHashFromAddress(address)
	UpdateKnownPeers.Lock()
	peer, present := d.knownPeers[hash]
	UpdateKnownPeers.Unlock()
	// If it exists, return it, otherwise create and add to knownPeers
	if !present {
		temp := new(Peer).Init(address, 0, RegularPeer)
		peer = *temp
		d.updatePeer(peer)
	}
	return peer
}

// PrintPeers Print details about the known peers
func (d *Discovery) PrintPeers() {
	note("discovery", "\n\n\n\nPeer Report:")
	UpdateKnownPeers.Lock()
	for key, value := range d.knownPeers {
		note("discovery", "%s \t Address: %s \t Quality: %d", key, value.Address, value.QualityScore)
	}
	UpdateKnownPeers.Unlock()
	note("discovery", "End Peer Report\n\n\n\n")
}

// LoadPeers loads the known peers from disk OVERWRITING PREVIOUS VALUES
func (d *Discovery) LoadPeers() {
	file, err := os.Open(d.peersFilePath)
	if nil != err {
		logerror("discovery", "Discover.LoadPeers() File read error on file: %s, Error: %+v", d.peersFilePath, err)
		return
	}
	dec := json.NewDecoder(bufio.NewReader(file))
	UpdateKnownPeers.Lock()
	dec.Decode(&d.knownPeers)
	UpdateKnownPeers.Unlock()
	note("discovery", "LoadPeers() found %d peers in peers.josn", len(d.knownPeers))
	file.Close()
}

// SavePeers just saves our known peers out to disk. Called periodically.
func (d *Discovery) SavePeers() {
	// save known peers to peers.json
	d.lastPeerSave = time.Now()
	file, err := os.Create(d.peersFilePath)
	if nil != err {
		logerror("discovery", "Discover.SavePeers() File write error on file: %s, Error: %+v", d.peersFilePath, err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	UpdateKnownPeers.Lock()
	encoder.Encode(d.knownPeers)
	UpdateKnownPeers.Unlock()
	writer.Flush()
	note("discovery", "SavePeers() saved %d peers in peers.json", len(d.knownPeers))

}

// LearnPeers recieves a set of peers from other hosts
// The unique peers are added to our peer list.
// The peers are in a json encoded string as a byte slice
func (d *Discovery) LearnPeers(payload []byte) {
	dec := json.NewDecoder(bytes.NewReader(payload))
	var peerArray []Peer
	err := dec.Decode(&peerArray)
	if nil != err {
		logfatal("discovery", "Discovery.LearnPeers got an error unmarshalling json. error: %+v json: %+v", err, strconv.Quote(string(payload)))
		return
	}
	for _, value := range peerArray {
		if d.isPeerPresent(value) {
			value.QualityScore = 0
			d.updatePeer(value)
			note("discovery", "Discovery.LearnPeers !!!!!!!!!!!!! Discoverd new PEER!   %+v ", value)
		}
	}
}

// GetStartupPeers gets a set of peers to connect to on startup
// For now, this gives a set of 12 of the total known peers.
// We want peers from diverse networks.  So,method is this:
//	-- generate list of candidates (if exclusive, only special peers)
//	-- sort candidates by distance
//  -- if num canddiates is less than desired set, return all candidates
//  -- Otherwise,repeatedly take candidates at the 0%, %25, %50, %75, %100 points in the list
//  -- remove each candidate from the list.
//  -- continue until there are no candidates left, or we have our set.
func (d *Discovery) GetStartupPeers() []Peer {
	peerPool := []Peer{}
	selectedPeers := []Peer{}
	UpdateKnownPeers.Lock()
	for _, peer := range d.knownPeers {
		if OnlySpecialPeers {
			if SpecialPeer == peer.Type {
				peerPool = append(peerPool, peer)
			}
		} else {
			peerPool = append(peerPool, peer)
		}
	}
	UpdateKnownPeers.Unlock()
	sort.Sort(PeerDistanceSort(peerPool))
	desiredQuantity := NumberPeersToConnect * 3 // Get three times as many as who knows how many will be online
	if len(peerPool) < desiredQuantity {
		return peerPool
	}
	for index := 1; index < desiredQuantity; index++ {
		selectedPeers = append(selectedPeers, peerPool[int(index/desiredQuantity*len(peerPool))])
	}
	return selectedPeers
}

// SharePeers gets a set of peers to send to other hosts
// For now, this gives a random set of 24 of the total known peers.
// The peers are in a json encoded string as byte slice
func (d *Discovery) SharePeers() []byte {
	return d.getPeerSelection()
}

// // Returns a set of peers from the ones we know about.
// // Right now returns the set of all peers we know about.
// // sharePeers is called from the Controllers runloop goroutine
// // The peers are in a json encoded string
// func (d *Discovery) ServePeers() string {
// 	return string(d.getPeerSelection())
// }

// getPeerSelection gets a selection of peers for SHARING.  So we want to share quality peers with the
// network.  Therefore, we sort by quality, and filter out special peers
func (d *Discovery) getPeerSelection() []byte {

	// var peer, currentBest Peer
	// var currentBestDistance float64
	selectedPeers := []Peer{}
	peerPool := []Peer{}
	UpdateKnownPeers.Lock()
	for _, peer := range d.knownPeers {
		peerPool = append(peerPool, peer)
	}
	UpdateKnownPeers.Unlock()
	sort.Sort(PeerQualitySort(peerPool))
	for _, peer := range peerPool {
		if SpecialPeer != peer.Type { // we don't share special peers
			selectedPeers = append(selectedPeers, peer)
		}
	}

	json, err := json.Marshal(selectedPeers)
	if nil != err {
		logerror("discovery", "Discovery.getPeerSelection got an error marshalling json. error: %+v selectedPeers: %+v", err, selectedPeers)
	}
	note("discovery", "peers we are sharing: %+v", string(json))
	return json
}

// Mbe a DDOS resistence mechanism that looks at rate of bad messsages over time.
// Right now, we just get enough demerits and we give up on the peer... forever.
// func (c *Connection) gotBadMessage() {
// 	debug(c.peer.Hash, "Connection.gotBadMessage()")
// 	// TODO Track bad messages to ban bad peers at network level
// 	// Array of in Connection of bad messages
// 	// Add this one to the array with timestamp
// 	// Filter all messages with timestamps over an hour (put value in protocol.go maybe an hour is too logn)
// 	// If count of bad messages in last hour exceeds threshold from protocol.go then we drop connection
// 	// Add this IP address to our banned peers (for an hour or day, also define in protocol.go)
// }
