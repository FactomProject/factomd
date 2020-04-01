package p2p

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// PeerCache is the object that gets json-marshalled and written to disk
type PeerCache struct {
	Bans  map[string]time.Time `json:"bans"` // can be ip or ip:port
	Peers []Endpoint           `json:"peers"`
}

func newPeerCache() *PeerCache {
	pc := new(PeerCache)
	pc.Bans = make(map[string]time.Time)
	return pc
}

func loadPeerCache(path string) (*PeerCache, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pc := newPeerCache()
	if err := json.Unmarshal(data, pc); err != nil {
		return nil, err
	}

	// don't load bans that timed out
	for k, v := range pc.Bans {
		if v.Before(time.Now()) {
			delete(pc.Bans, k)
		}
	}

	return pc, nil
}

func (pc *PeerCache) WriteToFile(path string) error {
	data, err := json.Marshal(pc)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644) // rw r r
}

func (c *controller) createPeerCache() *PeerCache {
	pc := newPeerCache()

	c.banMtx.Lock()
	now := time.Now()
	for addr, end := range c.bans {
		if end.Before(now) {
			delete(c.bans, addr)
		} else {
			pc.Bans[addr] = end
		}
	}
	c.banMtx.Unlock()

	peers := c.peers.Slice()
	pc.Peers = make([]Endpoint, len(peers))
	for i, p := range peers {
		pc.Peers[i] = p.Endpoint
	}

	return pc
}

// wrappers for reading and writing the peer file
func (c *controller) writePeerCache() error {
	if c.net.conf.PeerCacheFile == "" {
		return nil
	}
	return c.createPeerCache().WriteToFile(c.net.conf.PeerCacheFile)
}

func (c *controller) loadPeerCache() (*PeerCache, error) {
	if c.net.conf.PeerCacheFile == "" {
		return nil, nil
	}
	return loadPeerCache(c.net.conf.PeerCacheFile)
}
