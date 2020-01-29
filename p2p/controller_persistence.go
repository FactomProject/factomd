package p2p

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Persist is the object that gets json-marshalled and written to disk
type Persist struct {
	Bans      map[string]time.Time `json:"bans"` // can be ip or ip:port
	Bootstrap []Endpoint           `json:"bootstrap"`
}

func (c *controller) loadPersist() (*Persist, error) {
	persistData, err := c.loadPersistFile()
	if err != nil {
		return nil, err
	}

	if len(persistData) > 0 {
		return c.parsePersist(persistData)
	}

	return nil, nil
}

// wrappers for reading and writing the peer file
func (c *controller) writePersistFile(data []byte) error {
	if c.net.conf.PersistFile == "" {
		return nil
	}
	return ioutil.WriteFile(c.net.conf.PersistFile, data, 0644) // rw r r
}

func (c *controller) loadPersistFile() ([]byte, error) {
	if c.net.conf.PersistFile == "" {
		return nil, nil
	}
	return ioutil.ReadFile(c.net.conf.PersistFile)
}

func (c *controller) persistData() ([]byte, error) {
	var pers Persist
	pers.Bans = make(map[string]time.Time)

	c.banMtx.Lock()
	now := time.Now()
	for addr, end := range c.bans {
		if end.Before(now) {
			delete(c.bans, addr)
		} else {
			pers.Bans[addr] = end
		}
	}
	c.banMtx.Unlock()

	peers := c.peers.Slice()
	pers.Bootstrap = make([]Endpoint, len(peers))
	for i, p := range peers {
		pers.Bootstrap[i] = p.Endpoint
	}

	return json.Marshal(pers)
}

func (c *controller) parsePersist(data []byte) (*Persist, error) {
	var pers Persist
	err := json.Unmarshal(data, &pers)
	if err != nil {
		return nil, err
	}

	// decoding from a blank or invalid file
	if pers.Bans == nil {
		pers.Bans = make(map[string]time.Time)
	}

	c.logger.Debugf("bootstrapping with %d ips and %d bans", len(pers.Bootstrap), len(pers.Bans))
	return &pers, nil
}

func (c *controller) persistPeerFile() {
	if c.net.conf.PersistFile == "" {
		return
	}

	data, err := c.persistData()
	if err != nil {
		c.logger.WithError(err).Warn("unable to create peer persist data")
	} else {
		err = c.writePersistFile(data)
		if err != nil {
			c.logger.WithError(err).Warn("unable to persist peer data")
		}
	}
}
