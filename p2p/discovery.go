// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"encoding/json"
	"time"
)


type Discovery struct {
	knownPeers    map[string]Peer // peers we know about indexed by hash
	peersFilePath string          // the path to the peers.
}

// Discovery provides the code for sharing and managing peers,
// namely keeping track of all the peers we know about (not just the ones
// we are connected to.)  The discovery "service" is owned by the
// Controller and its routines are called from the Controllers runloop()
// This ensures that all shared memory is accessed from that goroutine.

func (d *Discovery) Init(peersFile string) {
	d.peersFilePath = peersFile
	d.knownPeers = map[string]Peer{}
	d.LoadPeers()
}

// SavePeers Merges the passed in peers with known peers and saves to disk.
func (d *Discovery) SavePeers(peers []Peer) {
	// BUGBUG TODO
	// Iterate thru the peers an reconcile iwth known peers
	// Update quality scores, etc.
	// save known peers to peers.json
  file, err := ioutil.WriteFile(d.peersFilePath)
        if nil != err {
			logerror(true, "Discover.SavePeers() File read error on file: %s, Error: %+v", d.peersFilePath, err)
			return
        }
	encoder := json.NewEncoder(file)
	encoder.Encode(d.knownPeers)
}

// LoadPeers loads the known peers from disk OVERWRITING PREVIOUS VALUES
func (d *Discovery) LoadPeers() {
  file, err := ioutil.ReadFile(d.peersFilePath)
        if nil != err {
			logerror(true, "Discover.LoadPeers() File read error on file: %s, Error: %+v", d.peersFilePath, err)
			return
        }
	// BUGBUG TODO IMPLEMENT JAY

        dec := json.NewDecoder(bytes.NewReader(file))
        var d myjson
        dec.Decode(&d)	decoder := json.NewDecoder()
}

// SharePeers gets a set of peers to send to other hosts
// For now, this gives a random set of 24 of the total known peers.
// The peers are in a json encoded string as byte slice
func (d *Discovery) SharePeers() []byte {
	return []byte(d.getPeerSelection())
}
// Returns a set of peers from the ones we know about.
// Right now returns the set of all peers we know about.
// sharePeers is called from the Controllers runloop goroutine
func (d *Discovery) ServePeers() {
	return d.getPeerSelection()
}

// Returns a set of peers from the ones we know about.
// Right now returns the set of all peers we know about.
// Should maybe return the top 24 peers, based on distance from a randomly chosen peer (maximizing distance)
// Returns the peers as a JSON string (maybe this should be []Peers, but right now only SharePeers
// and ServePeers call this.)
func (d *Discovery) getPeerSelection() string {
	// BUGBUG TODO IMPLEMENT JAY
	// Pick a random peer
	// Iterate thru the peers and find the peer that is furthest away by location
	// Add it to the list/
	// If the list is less than 24, keep going unless the list has all the peers
	// For the next peer use the previous peer and find the peer furthest away.
	// Take the list of peers and conver it to JSON
	json := json.Marshl(d.knownPeers)
	// Put it in the connections channel to send out
return json
}


// LearnPeers recieves a set of peers from other hosts
// The unique peers are added to our peer list.
// The peers are in a json encoded string as a byte slice
func (d *Discovery) LearnPeers(payload []byte) {
	// BUGBUG TODO IMPLEMENT JAY
}



// GetStartupPeers gets a set of peers to connect to on startup
// For now, this gives a random set of 12 of the total known peers.
func (d *Discovery) GetStartupPeers(peer Peer) []Peer {
	// BUGBUG TODO IMPLEMENT JAY
}


// GetFullPeer looks for a peer in the known peers, and if so, returns it  (based on
// the hash of the passed in peer.)  If the peer is unknown , we create it and
// add it to the known peers.
func (d *Discovery) GetFullPeer(prototype Peer) Peer {
	// Look up Peer
	// If it exists, return it
	// IF it doesn't exist, make ones
	// Add the new one to the known peers.
		// BUGBUG TODO IMPLEMENT JAY

}

// GetPeerByHash looks for a peer in the known peers, and if so, returns it 
// If the peer is unknown , we create it and add it to the known peers.
func (d *Discovery) GetPeerByHash(hash string) Peer {
	// Look up Peer
	// If it exists, return it
	// IF it doesn't exist, make ones
	// Add the new one to the known peers.
		// BUGBUG TODO IMPLEMENT JAY

}

// UpdatePeer updates the values in our known peers. Creates peer if its not in there.
func (d *Discovery) UpdatePeer(peer Peer) {
 // If peer is in database, then update its qualtiy score, if lower (peers
 // that dial into us start with a zero score befor)
	// BUGBUG TODO IMPLEMENT JAY
}


