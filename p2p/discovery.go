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
	"time"
)

type Discovery struct {
	knownPeers map[string]Peer // peers we know about indexed by hash

	peersFilePath string     // the path to the peers.
	lastPeerSave  time.Time  // Last time we saved known peers.
	rng           *rand.Rand // RNG = random number generator
}

// Discovery provides the code for sharing and managing peers,
// namely keeping track of all the peers we know about (not just the ones
// we are connected to.)  The discovery "service" is owned by the
// Controller and its routines are called from the Controllers runloop()
// This ensures that all shared memory is accessed from that goroutine.

func (d *Discovery) Init(peersFile string) *Discovery {
	d.peersFilePath = peersFile
	d.knownPeers = map[string]Peer{}
	d.LoadPeers()
	d.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	return d
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
	encoder.Encode(d.knownPeers)
	writer.Flush()
	note("discovery", "SavePeers() saved %d peers in peers.josn", len(d.knownPeers))

}

// LoadPeers loads the known peers from disk OVERWRITING PREVIOUS VALUES
func (d *Discovery) LoadPeers() {
	file, err := os.Open(d.peersFilePath)
	if nil != err {
		logerror("discovery", "Discover.LoadPeers() File read error on file: %s, Error: %+v", d.peersFilePath, err)
		return
	}
	dec := json.NewDecoder(bufio.NewReader(file))
	dec.Decode(&d.knownPeers)
	note("discovery", "LoadPeers() found %d peers in peers.josn", len(d.knownPeers))
	file.Close()
}

// SharePeers gets a set of peers to send to other hosts
// For now, this gives a random set of 24 of the total known peers.
// The peers are in a json encoded string as byte slice
func (d *Discovery) SharePeers() []byte {
	return d.getPeerSelection()
}

// Returns a set of peers from the ones we know about.
// Right now returns the set of all peers we know about.
// sharePeers is called from the Controllers runloop goroutine
// The peers are in a json encoded string
func (d *Discovery) ServePeers() string {
	return string(d.getPeerSelection())
}

// BUGBUG - we need to filter on special peers, and rewrite share peers and serve peers for this
// Also need to update the controller stuff for them.
// Exclusive flag is:  OnlySpecialPeers

// Think the way to do this is to have known peers be peers in the general peer to peer network.
// Exclusive peers are not shared, and stored in a seperate file.

// Returns a set of peers from the ones we know about.
// Right now returns the set of all peers we know about.
// Should maybe return the top 24 peers, based on distance from a randomly chosen peer (maximizing distance)
// Returns the peers as a JSON string (maybe this should be []Peers, but right now only SharePeers
// and ServePeers call this.)
func (d *Discovery) getPeerSelection() []byte {
	// BUGBUG TODO
	// BUGBUG doesn't take into account peer type or exclusive flag
	// BUGBUG couldn't we implemetn a distance sort, the n take the first and last of the sorted peers as furthest away?

	// var peer, currentBest Peer
	// var currentBestDistance float64
	selectedPeers := []Peer{}
	peerPool := []Peer{}
	for _, peer := range d.knownPeers {
		// if SpecialPeer != peer.Type {
		peerPool = append(peerPool, peer)
		// }
	}
	// Temporarily return all peers:
	selectedPeers = peerPool

	// This needs to be refactored:
	// numPeers := len(peerPool)
	// numToSelect := 24
	// if numToSelect > numPeers {
	// 	// then we return all of the peers
	// 	selectedPeers = peerPool
	// } else {
	// 	selectedPeers = peerPool
	// BUGBUG Rewrite this to sort by location, then take first, middle, last repeatedly. More deterministic
	// 	// Pick a random peer
	// 	if numToSelect >= 1 && numPeers > 0 {
	// 		peer = peers[d.rng.Intn(numPeers-1)]
	// 	} else {
	// 		return []byte{}
	// 	}
	// 	currentBest = peer
	// 	currentBestDistance = 0.0
	// 	selectedPeers := []Peer{peer}
	// 	verbose("discovery", "getPeerSelection() numToSelect %d, numPeers: %d", numToSelect, numPeers)

	// 	for numToSelect > len(selectedPeers) {
	// 		verbose("discovery", "getPeerSelection() numToSelect %d, selected: %d", numToSelect, len(selectedPeers))
	// 		// Iterate thru the peers and find the peer that is furthest away by location
	// 		for _, target := range peers {
	// 			distance := math.Abs(float64(target.Location) - float64(peer.Location))
	// 			if distance > currentBestDistance { // better peer found
	// 				currentBest = target
	// 				currentBestDistance = distance
	// 			}
	// 		}
	// 		if currentBest != peer {
	// 			selectedPeers = append(selectedPeers, currentBest)
	// 			currentBestDistance = 0.0
	// 			currentBest = peer
	// 		}
	// 	}
	// }

	json, err := json.Marshal(selectedPeers)
	if nil != err {
		logerror("discovery", "Discovery.getPeerSelection got an error marshalling json. error: %+v selectedPeers: %+v", err, selectedPeers)
	}
	note("discovery", "peers we are sharing: %+v", string(json))
	return json
}

// LearnPeers recieves a set of peers from other hosts
// The unique peers are added to our peer list.
// The peers are in a json encoded string as a byte slice
func (d *Discovery) LearnPeers(payload []byte) {
	dec := json.NewDecoder(bytes.NewReader(payload))
	var peerArray []Peer
	err := dec.Decode(&peerArray)
	if nil != err {
		logerror("discovery", "Discovery.LearnPeers got an error unmarshalling json. error: %+v json: %+v", err, strconv.Quote(string(payload)))
		return
	}
	for _, value := range peerArray {
		_, present := d.knownPeers[value.Hash]
		if !present {
			value.QualityScore = 0
			d.knownPeers[value.Hash] = value
			note("discovery", "Discovery.LearnPeers !!!!!!!!!!!!! Discoverd new PEER!   %+v ", value)
		}
	}
}

// BUGBUG Have PeerQualitySort and PeerDistanceSort implemented.
// TODO:
// 	-- Get peers by quality on startup (pass in number you want)
// 	-- Get peers by distance (When adding additional?)

// GetStartupPeers gets a set of peers to connect to on startup
// For now, this gives a set of 12 of the total known peers.
func (d *Discovery) GetStartupPeers() []Peer {
	peers := []Peer{}
	// reverse := []Peer{}
	selectedPeers := []Peer{}
	for _, peer := range d.knownPeers {
		peers = append(peers, peer)
		// reverse = append(peers, peer)
	}
	// Generate a sort of the known peers by quality score
	sort.Sort(PeerDistanceSort(peers))
	// sort.Sort(sort.Reverse(PeerDistanceSort(peers)))
	// Do we only connect to special peers?
	// need a flag for special peers and logic
	// Otherwise take the N best peers from known peers.
	// BUGBUG this doesn't find diverse distances.
	for _, peer := range peers {
		if len(selectedPeers) < NumberPeersToConnect {
			if OnlySpecialPeers {
				if SpecialPeer == peer.Type {
					selectedPeers = append(selectedPeers, peer)
				}
			} else {
				selectedPeers = append(selectedPeers, peer)
			}
		}
	}
	return selectedPeers
}

// GetFullPeer looks for a peer in the known peers, and if so, returns it  (based on
// the hash of the passed in peer.)  If the peer is unknown , we create it and
// add it to the known peers.
// func (d *Discovery) GetFullPeer(prototype Peer) Peer {
// 	return d.GetPeerByAddress(prototype.Address)
// }

func (d *Discovery) GetPeerByAddress(address string) Peer {
	hash := PeerHashFromAddress(address)
	peer, present := d.knownPeers[hash]
	// If it exists, return it, otherwise create and add to knownPeers
	if !present {
		temp := new(Peer).Init(address, 0, RegularPeer)
		peer = *temp
		d.knownPeers[hash] = peer
	}
	return peer
}

// UpdatePeer updates the values in our known peers. Creates peer if its not in there.
func (d *Discovery) UpdatePeer(peer Peer) {
	d.knownPeers[peer.Hash] = peer
}

// PrintPeers Print details about the known peers
func (d *Discovery) PrintPeers() {
	note("discovery", "\n\n\n\nPeer Report:")
	for key, value := range d.knownPeers {
		note("discovery", "%s \t Address: %s \t Quality: %d", key, value.Address, value.QualityScore)
	}
	note("discovery", "\n\n\n\n")
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
