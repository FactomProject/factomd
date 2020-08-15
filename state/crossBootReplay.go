package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/database/boltdb"
	"github.com/PaulSnow/factom2d/database/mapdb"
)

var _ = fmt.Println

// DB Buckets
var (
	heightBucket = []byte("Heights")
	lowest       = []byte("LowestHeight")
	saltBucket   = []byte("AllSalts")
)

// SetupCrossBootReplay will construct the database
func (s *State) SetupCrossBootReplay(path string) {
	// Already initialized
	if s.CrossReplay != nil {
		return
	}
	// Map Database uses a map replay filter
	if path != "Map" {
		path = filepath.Join(s.BoltDBPath, s.Network, "crossreplay.db")
	}
	s.CrossReplay = NewCrossReplayFilter(path)
	// This thread will terminate itself
	go s.CrossReplay.Run()
}

// CrossReplayAddSalt adds the salt to the DB
func (s *State) CrossReplayAddSalt(height uint32, salt [8]byte) error {
	return s.CrossReplay.AddSalt(height, salt)
}

// CrossReplayFilter checks for old messages across reboots based on the salts
// inside the ack messages. It saves all salts of leaders it sees while running.
// On reboot, it will ignore all messages that have an old salt (for a set duration).
// After the duration, no new salts are saved (extra overhead we don't need) and will stop
// ignoring messages based on salts (so a single leader reboot will rejoin the network).
type CrossReplayFilter struct {
	Currentheight int
	LowestHeight  int

	// Indicates we have been running for awhile
	// and should already have the salts
	stopAddingSalts  bool
	endTime          time.Time
	currentSaltCache map[[8]byte]bool
	oldSaltCache     map[[8]byte]bool
	db               interfaces.IDatabase
}

func NewCrossReplayFilter(path string) *CrossReplayFilter {
	c := new(CrossReplayFilter)
	if path == "" || strings.ToLower(path) == "map" {
		c.db = new(mapdb.MapDB)
	} else {
		c.db = boltdb.NewAndCreateBoltDB([][]byte{}, path)
	}

	// Curent is used to not write to disk when not needed
	c.currentSaltCache = make(map[[8]byte]bool)
	// Old is the salts on the previous boot
	c.oldSaltCache = make(map[[8]byte]bool)
	// Load the old salts into the map
	c.loadOldSalts()
	c.endTime = time.Now().Add(constants.CROSSBOOT_SALT_REPLAY_DURATION)

	var m MarshalableUint32
	c.db.Get(heightBucket, lowest, &m)
	c.LowestHeight = int(m)
	c.Currentheight = int(m)

	return c
}

// loadOldSalts loads the db into memory, and clears the db
func (c *CrossReplayFilter) loadOldSalts() {
	keys, _ := c.db.ListAllKeys(saltBucket)
	for _, k := range keys {
		var s [8]byte
		copy(s[:], k[:])
		c.oldSaltCache[s] = true
	}
	// Reset the db
	c.db.Clear(saltBucket)
}

// AddSalt will add the salt to the replay filter that is used on reboot
func (c *CrossReplayFilter) AddSalt(height uint32, salt [8]byte) error {
	// We don't need to add any more salts to the db
	if c.stopAddingSalts {
		return nil
	}

	// The current salts that are in the db
	if _, ok := c.currentSaltCache[salt]; ok {
		return nil
	}

	// Need something to marshal... the data is no longer used
	m := MarshalableUint32(0)
	err := c.db.Put(saltBucket, salt[:], &m)
	if err != nil {
		return err
	}

	return nil
}

// ExistOldSalt checks to see if the salt existed on the previous boot
func (c *CrossReplayFilter) ExistOldSalt(salt [8]byte) bool {
	_, ok := c.oldSaltCache[salt]
	return ok
}

// Exists check if the hash is in the replay filter, and if it encounters a db error, it will report false
//
func (c *CrossReplayFilter) ExistSalt(salt [8]byte) (bool, error) {
	return c.db.DoesKeyExist(saltBucket, salt[:])
}

// Run is a simple loop that ensures we discard old data we do not need.
func (c *CrossReplayFilter) Run() {
	for {
		time.Sleep(time.Second * 5)
		if c.endTime.Before(time.Now()) {
			// We no longer need to add salts
			c.stopAddingSalts = true
			return
		}
		c.collectGarbage()
	}
}

// collectGarbage will delete buckets for older heights that we no longer care about
func (c *CrossReplayFilter) collectGarbage() {
	// fmt.Printf("Cross Replay Cleanup: %d -> %d\n", c.LowestHeight, c.Currentheight)
	// Have more than 5 blocks worth of data
	if c.LowestHeight < c.Currentheight-5 {
		for i := c.LowestHeight; i < c.Currentheight-5; i++ {
			bucket, err := Uint32ToBytes(uint32(i))
			if err != nil {
				continue
			}
			c.db.Clear(bucket)
			c.db.Clear(append([]byte("salt"), bucket...))
			c.LowestHeight = i + 1
		}
	}
	var m MarshalableUint32 = MarshalableUint32(uint32(c.LowestHeight))
	c.db.Put(heightBucket, lowest, &m)
}

func (c *CrossReplayFilter) Close() {
	c.db.Close()
}

// Used for marshal/unmarshal uint32

type MarshalableUint32 uint32

func (m *MarshalableUint32) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MarshalableUint32.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return Uint32ToBytes(uint32(*m))
}

func (m *MarshalableUint32) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MarshalableUint32) UnmarshalBinaryData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("Need at least 4 bytes")
	}
	v, err := BytesToUint32(data[:4])
	if err != nil {
		return nil, err
	}
	*m = MarshalableUint32(v)
	return data[4:], nil
}
