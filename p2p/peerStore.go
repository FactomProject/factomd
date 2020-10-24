package p2p

import (
	"fmt"
	"sync"
)

// PeerStore holds active Peers, managing them in a concurrency safe
// manner and providing lookup via various functions
type PeerStore struct {
	mtx       sync.RWMutex
	peers     map[string]*Peer // hash -> peer
	connected map[string]int   // (ip|ip:port) -> count
	curSlice  []*Peer          // temporary slice that gets reset when changes are made
	incoming  int
	outgoing  int
}

// NewPeerStore initializes a new peer store
func NewPeerStore() *PeerStore {
	ps := new(PeerStore)
	ps.peers = make(map[string]*Peer)
	ps.connected = make(map[string]int)
	return ps
}

// Add a peer to be managed. Returns an error if a peer with that hash
// is already tracked
func (ps *PeerStore) Add(p *Peer) error {
	if p == nil {
		return fmt.Errorf("trying to add nil")
	}
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if _, ok := ps.peers[p.Hash]; ok {
		return fmt.Errorf("peer already exists")
	}
	ps.curSlice = nil
	ps.peers[p.Hash] = p
	ps.connected[p.Endpoint.IP]++
	ps.connected[p.Endpoint.String()]++

	if p.IsIncoming {
		ps.incoming++
	} else {
		ps.outgoing++
	}
	return nil
}

// Remove a specific peer if it exists. This checks by pointer reference and not by hash.
// If you have two distinct peer instances (A and B) with the same hash and add A, removing B has no
// effect, even if they have the same values
func (ps *PeerStore) Remove(p *Peer) {
	if p == nil {
		return
	}
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	if old, ok := ps.peers[p.Hash]; ok && old == p { // pointer comparison
		ps.connected[p.Endpoint.IP]--
		if ps.connected[p.Endpoint.IP] == 0 {
			delete(ps.connected, p.Endpoint.IP)
		}
		ps.connected[p.Endpoint.String()]--
		if ps.connected[p.Endpoint.String()] == 0 {
			delete(ps.connected, p.Endpoint.String())
		}
		if old.IsIncoming {
			ps.incoming--
		} else {
			ps.outgoing--
		}
		ps.curSlice = nil
		delete(ps.peers, p.Hash)
	}
}

// Total amount of peers connected
func (ps *PeerStore) Total() int {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return len(ps.peers)
}

// Outgoing is the amount of outgoing peers connected
func (ps *PeerStore) Outgoing() int {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return ps.outgoing
}

// Incoming is the amount of incoming peers connected
func (ps *PeerStore) Incoming() int {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return ps.incoming
}

// Get retrieves a Peer with a specific hash, nil if it doesn't exist
func (ps *PeerStore) Get(hash string) *Peer {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return ps.peers[hash]
}

// Connections tests whether there is a peer connected from a specified ip address
func (ps *PeerStore) Connections(addr string) int {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return ps.connected[addr]
}

// Connected returns if there is a peer from that specific endpoint
func (ps *PeerStore) Connected(ep Endpoint) bool {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return ps.connected[ep.String()] > 0
}

// Count returns the amount of peers connected from a specified ip address
func (ps *PeerStore) Count(addr string) int {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return ps.connected[addr]
}

// Slice returns a slice of the current peers that is considered concurrency
// safe for reading operations. The slice should not be modified. Peers are randomly
// ordered
func (ps *PeerStore) Slice() []*Peer {
	ps.mtx.RLock()
	if ps.curSlice != nil {
		defer ps.mtx.RUnlock()
		return append(ps.curSlice[:0:0], ps.curSlice...)
	}
	ps.mtx.RUnlock()

	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	r := make([]*Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		r = append(r, p)
	}
	ps.curSlice = r
	return r
}
