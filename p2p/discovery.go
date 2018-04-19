// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"bufio"
	"bytes"
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var discoLogger = packageLogger.WithField("subpack", "discovery")

type Discovery struct {
	knownPeers map[string]Peer // peers we know about indexed by hash

	peersFilePath string     // the path to the peers.
	lastPeerSave  time.Time  // Last time we saved known peers.
	rng           *rand.Rand // RNG = random number generator
	seedURL       string     // URL to the source of a list of peers

	// logging
	logger *log.Entry
}

var UpdateKnownPeers sync.Mutex

// Discovery provides the code for sharing and managing peers,
// namely keeping track of all the peers we know about (not just the ones
// we are connected to.)  The discovery "service" is owned by the
// Controller and its routines are called from the Controllers runloop()
// This ensures that all shared memory is accessed from that goroutine.

func (d *Discovery) Init(peersFile string, seed string) *Discovery {
	d.logger = discoLogger
	UpdateKnownPeers.Lock()
	d.knownPeers = map[string]Peer{}
	UpdateKnownPeers.Unlock()
	d.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	d.peersFilePath = peersFile
	d.seedURL = seed
	//d.LoadPeers()
	d.DiscoverPeersFromSeed()
	return d
}

// Only controller should be able to read this, but we still got
// a concurrent read/write error, so isolating changes to knownPeers

// UpdatePeer updates the values in our known peers. Creates peer if its not in there.
func (d *Discovery) updatePeer(peer Peer) {
	d.logger.Debugf("Updating peer: %v", peer)
	UpdateKnownPeers.Lock()

	_, ok := d.knownPeers[peer.Address]
	if !ok {
		d.logger.WithFields(log.Fields{
			"address":     peer.Address,
			"last_source": peer.LastSource()}).Infof("Discovered new peer")
	}

	d.knownPeers[peer.Address] = peer
	UpdateKnownPeers.Unlock()
}

// getPeer returns a known peer, if present
func (d *Discovery) getPeer(address string) Peer {
	UpdateKnownPeers.Lock()
	thePeer := d.knownPeers[address]
	UpdateKnownPeers.Unlock()
	return thePeer
}

// UpdatePeer updates the values in our known peers. Creates peer if its not in there.
func (d *Discovery) isPeerPresent(peer Peer) bool {
	UpdateKnownPeers.Lock()
	_, present := d.knownPeers[peer.Address]
	UpdateKnownPeers.Unlock()
	return present
}

// LoadPeers loads the known peers from disk OVERWRITING PREVIOUS VALUES
func (d *Discovery) LoadPeers() {
	file, err := os.Open(d.peersFilePath)
	if nil != err {
		d.logger.Errorf("Discover.LoadPeers() File read error on file: %s, Error: %+v", d.peersFilePath, err)
		return
	}
	dec := json.NewDecoder(bufio.NewReader(file))
	UpdateKnownPeers.Lock()
	dec.Decode(&d.knownPeers)
	// since this is run at startup, reset quality scores.
	for _, peer := range d.knownPeers {
		peer.QualityScore = 0
		peer.Location = peer.LocationFromAddress()
		d.knownPeers[peer.Address] = peer
	}
	UpdateKnownPeers.Unlock()
	d.logger.Debugf("LoadPeers() found %d peers in peers.josn", len(d.knownPeers))
	file.Close()
}

// SavePeers just saves our known peers out to disk. Called periodically.
func (d *Discovery) SavePeers() {
	// save known peers to peers.json
	d.lastPeerSave = time.Now()
	file, err := os.Create(d.peersFilePath)
	if nil != err {
		d.logger.Errorf("Discover.SavePeers() File write error on file: %s, Error: %+v", d.peersFilePath, err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	var qualityPeers = map[string]Peer{}
	UpdateKnownPeers.Lock()
	for _, peer := range d.knownPeers {
		switch {
		case peer.Type == SpecialPeerConfig || peer.Type == SpecialPeerCmdLine: // always save special peers, even if we haven't talked in awhile.
			qualityPeers[peer.AddressPort()] = peer
			d.logger.Debugf("SavePeers() saved peer in peers.json: %+v", peer)

		case time.Since(peer.LastContact) > time.Hour*168:
			d.logger.Debugf("SavePeers() DID NOT SAVE peer in peers.json. Last Contact greater than 168 hours. Peer: %+v", peer)
			break
		case MinumumQualityScore > peer.QualityScore:
			d.logger.Debugf("SavePeers() DID NOT SAVE peer in peers.json. MinumumQualityScore: %d > Peer quality score.  Peer: %+v", MinumumQualityScore, peer)
			break
		default:
			qualityPeers[peer.AddressPort()] = peer
		}
	}
	UpdateKnownPeers.Unlock()
	encoder.Encode(qualityPeers)
	writer.Flush()
	d.logger.Debugf("SavePeers() saved %d peers in peers.json. \n They were: %+v", len(qualityPeers), qualityPeers)
}

// LearnPeers receives a set of peers from other hosts
// The unique peers are added to our peer list.
// The peers are in a json encoded string as a byte slice
func (d *Discovery) LearnPeers(parcel Parcel) {
	dec := json.NewDecoder(bytes.NewReader(parcel.Payload))
	var peerArray []Peer
	err := dec.Decode(&peerArray)
	if nil != err {
		d.logger.Errorf("Discovery.LearnPeers got an error unmarshalling json. error: %+v json: %+v", err, strconv.Quote(string(parcel.Payload)))
		return
	}
	filteredArray := d.filterPeersFromOtherNetworks(peerArray)
	for _, value := range filteredArray {
		value.QualityScore = 0
		switch d.isPeerPresent(value) {
		case true:
			alreadyKnownPeer := d.getPeer(value.Address)
			d.updatePeer(d.updatePeerSource(alreadyKnownPeer, parcel.Header.PeerAddress))
		default:
			value.Source = map[string]time.Time{parcel.Header.PeerAddress: time.Now()}
			d.updatePeer(value)
			d.logger.Debugf("Discovery.LearnPeers !!!!!!!!!!!!! Discovered new PEER!   %+v ", value)
		}
	}
	d.SavePeers()
}

// updatePeerSource checks to see if source is in peer's sources, and if not puts it in there with a value equal to time.Now()
func (d *Discovery) updatePeerSource(peer Peer, source string) Peer {
	if nil == peer.Source {
		peer.Source = map[string]time.Time{}
	}
	_, sp := peer.Source[source]
	if !sp {
		peer.Source[source] = time.Now()
	}
	return peer
}

func (d *Discovery) filterPeersFromOtherNetworks(peers []Peer) (filtered []Peer) {
	for _, peer := range peers {
		if CurrentNetwork == peer.Network {
			filtered = append(filtered, peer)
		}
	}
	return
}

func (d *Discovery) filterForUniqueIPAdresses(peers []Peer) (filtered []Peer) {
	unique := map[string]Peer{}
	for _, peer := range peers {
		_, present := unique[peer.Address]
		if !present {
			filtered = append(filtered, peer)
			unique[peer.Address] = peer
		}
	}
	return
}

// GetOutgoingPeers gets a set of peers to connect to on startup
// For now, this gives a set of 12 of the total known peers.
// We want peers from diverse networks.  So,method is this:
//	-- generate list of candidates (if exclusive, only special peers)
//	-- sort candidates by distance
//  -- if num canddiates is less than desired set, return all candidates
//  -- Otherwise,repeatedly take candidates at the 0%, %25, %50, %75, %100 points in the list
//  -- remove each candidate from the list.
//  -- continue until there are no candidates left, or we have our set.
func (d *Discovery) GetOutgoingPeers() []Peer {
	firstPassPeers := []Peer{}
	selectedPeers := map[string]Peer{}
	UpdateKnownPeers.Lock()
	for _, peer := range d.knownPeers {
		switch {
		case OnlySpecialPeers && (peer.Type == SpecialPeerConfig || peer.Type == SpecialPeerCmdLine):
			firstPassPeers = append(firstPassPeers, peer)
		case !OnlySpecialPeers:
			firstPassPeers = append(firstPassPeers, peer)
		default:
		}
	}
	UpdateKnownPeers.Unlock()
	secondPass := d.filterPeersFromOtherNetworks(firstPassPeers)
	peerPool := d.filterForUniqueIPAdresses(secondPass)
	sort.Sort(PeerDistanceSort(peerPool))
	// Get four times as many as who knows how many will be online
	desiredQuantity := NumberPeersToConnect * 4
	// If the peer pool isn't at least twice the size of what we need, then location diversity is meaningless.
	if len(peerPool) < desiredQuantity*2 {
		return peerPool
	}
	// Algo is to divide peers up into buckets, sorted by distance.
	// Number of buckets is the number of peers we want to get.
	// Then given the size of each bucket, pick a random peer in the bucket.
	bucketSize := 1 + int(len(peerPool)-1/desiredQuantity)
	for index := 0; index < int(desiredQuantity); index++ {
		bucketIndex := int(index / desiredQuantity * len(peerPool))
		newPeerIndex := bucketIndex + rand.Intn(bucketSize)
		if newPeerIndex > len(peerPool)-1 {
			newPeerIndex = len(peerPool) - 1
		}
		newPeer := peerPool[newPeerIndex]
		selectedPeers[newPeer.Address] = newPeer
	}

	// Now derive a slice of peers to return
	finalSet := []Peer{}
	for _, v := range selectedPeers {
		finalSet = append(finalSet, v)
	}
	d.logger.Debugf("discovery.GetOutgoingPeers() got the following peers: %+v", finalSet)
	return finalSet
}

// SharePeers gets a set of peers to send to other hosts
// For now, this gives a random set of  the total known peers.
// The peers are in a json encoded string as byte slice
func (d *Discovery) SharePeers() []byte {
	return d.getPeerSelection()
}

// For now we use 4 * NumberPeersToConnect to share, which if connection
// rate is %25 will result in NumberPeersToConnect connections.

// getPeerSelection gets a selection of peers for SHARING.  So we want to share quality peers with the
// network.  Therefore, we sort by quality, and filter out special peers
func (d *Discovery) getPeerSelection() []byte {
	// var peer, currentBest Peer
	// var currentBestDistance float64
	selectedPeers := []Peer{}
	firstPassPeers := []Peer{}
	specialPeersByLocation := map[uint32]Peer{}
	UpdateKnownPeers.Lock()
	for _, peer := range d.knownPeers {
		if peer.QualityScore > MinumumSharingQualityScore { // Only share peers that have earned positive reputation
			firstPassPeers = append(firstPassPeers, peer)
		}
	}
	UpdateKnownPeers.Unlock()
	peerPool := d.filterPeersFromOtherNetworks(firstPassPeers)
	sort.Sort(PeerQualitySort(peerPool))
	// Pull out special peers by location.  Use location because it should more accurately reflect IP address.
	// we check by location to keep from sharing special peers when they dial into us (in which case we wouldn't realize
	// they were special by the flag.)
	for _, peer := range peerPool {
		if (peer.Type == SpecialPeerConfig || peer.Type == SpecialPeerCmdLine) && peer.Location != 0 { // only include special peers that have IP address
			specialPeersByLocation[peer.Location] = peer
		}
	}
	for _, peer := range peerPool {
		_, present := specialPeersByLocation[peer.Location]
		switch {
		case peer.Type == SpecialPeerConfig || peer.Type == SpecialPeerCmdLine:
			break
		case present:
			break
		default:
			selectedPeers = append(selectedPeers, peer)
		}
		if 4*NumberPeersToConnect <= len(selectedPeers) {
			break
		}
	}

	json, err := json.Marshal(selectedPeers)
	if nil != err {
		d.logger.Errorf("Discovery.getPeerSelection got an error marshalling json. error: %+v selectedPeers: %+v", err, selectedPeers)
	}
	d.logger.Debugf("peers we are sharing: %+v", string(json))
	return json
}

// DiscoverPeers gets a set of peers from a DNS Seed
func (d *Discovery) DiscoverPeersFromSeed() {
	d.logger.Info("Contacting seed URL to get peers")
	resp, err := http.Get(d.seedURL)
	if nil != err {
		d.logger.Errorf("DiscoverPeersFromSeed getting peers from %s produced error %+v", d.seedURL, err)
		return
	}
	defer resp.Body.Close()
	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	bad := 0
	for _, line := range lines {
		address, port, err := net.SplitHostPort(line)
		if err == nil {
			peerp := new(Peer).Init(address, port, 0, RegularPeer, 0)
			peer := *peerp
			peer.LastContact = time.Now()
			d.updatePeer(d.updatePeerSource(peer, "DNS-Seed"))
		} else {
			bad++
			d.logger.Errorf("Bad peer in " + d.seedURL + " [" + line + "]")
		}
	}
	d.logger.Debugf("DiscoverPeersFromSeed got peers: %+v", lines)
}
