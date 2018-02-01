package state

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/mapdb"
)

var _ = fmt.Println

var (
	heightBucket = []byte("Heights")
	lowest       = []byte("LowestHeight")
	saltBucket   = []byte("AllSalts")
)

func (s *State) SetupCrossBootReplay(path string) {
	// Alrady initialized
	if s.CrossReplay != nil {
		return
	}
	if path != "Map" {
		path = filepath.Join(s.BoltDBPath, s.Network, "crossreplay.db")
		//path = s.BoltDBPath + "/" + s.Network + "/" + "crossreplay.db"
	}
	fmt.Println(s.FactomNodeName)
	s.CrossReplay = NewCrossReplayFilter(path)
	go s.CrossReplay.Run()
}

func (s *State) CrossReplayAddHash(height uint32, hash interfaces.IHash) error {
	return s.CrossReplay.AddMsgHash(height, hash)
}

func (s *State) CrossReplayAddSalt(height uint32, salt [8]byte) error {
	return s.CrossReplay.AddSalt(height, salt)
}

func (s *State) CrossReplayExists(height uint32, hash interfaces.IHash) (bool, error) {
	return s.CrossReplay.Exists(height, hash)
}

type CrossReplayFilter struct {
	Currentheight int
	LowestHeight  int

	oldSaltCache map[[8]byte]bool
	db           interfaces.IDatabase
}

func NewCrossReplayFilter(path string) *CrossReplayFilter {
	c := new(CrossReplayFilter)
	if path == "" || strings.ToLower(path) == "map" {
		c.db = new(mapdb.MapDB)
	} else {
		fmt.Println(path)
		// path := s.LdbPath + "/" + s.Network + "/" + "factoid_level.db"
		c.db = boltdb.NewAndCreateBoltDB([][]byte{}, path)
	}

	c.oldSaltCache = make(map[[8]byte]bool)
	var m MarshalableUint32
	c.db.Get(heightBucket, lowest, &m)
	c.LowestHeight = int(m)
	c.Currentheight = int(m)
	c.loadOldSalts()

	return c
}

func (c *CrossReplayFilter) loadOldSalts() {
	keys, _ := c.db.ListAllKeys(saltBucket)
	for _, k := range keys {
		var s [8]byte
		copy(s[:], k[:])
		c.oldSaltCache[s] = true
	}
	c.db.Clear(saltBucket)
}

// AddMsgHash will add the hash to the replay filter
func (c *CrossReplayFilter) AddMsgHash(height uint32, hash interfaces.IHash) error {
	if int(height) > c.Currentheight {
		c.Currentheight = int(height)
		if c.LowestHeight == 0 {
			c.Currentheight = int(height)
		}
	}
	bucket, err := Uint32ToBytes(height)
	if err != nil {
		return err
	}

	err = c.db.Put(bucket, hash.Bytes(), hash)
	if err != nil {
		return err
	}

	return nil
}

// AddMsgHash will add the hash to the replay filter
func (c *CrossReplayFilter) AddSalt(height uint32, salt [8]byte) error {
	if int(height) > c.Currentheight {
		c.Currentheight = int(height)
		if c.LowestHeight == 0 {
			c.Currentheight = int(height)
		}
	}

	m := MarshalableUint32(height)
	err := c.db.Put(saltBucket, salt[:], &m)
	if err != nil {
		return err
	}

	return nil
}

func (c *CrossReplayFilter) ExistOldSalt(salt [8]byte) bool {
	_, ok := c.oldSaltCache[salt]
	return ok
}

// Exists check if the hash is in the replay filter, and if it encounters a db error, it will report false
func (c *CrossReplayFilter) ExistSalt(height uint32, salt [8]byte) (bool, error) {
	return c.db.DoesKeyExist(saltBucket, salt[:])
}

// Exists check if the hash is in the replay filter, and if it encounters a db error, it will report false
func (c *CrossReplayFilter) Exists(height uint32, hash interfaces.IHash) (bool, error) {
	bucket, err := Uint32ToBytes(height)
	if err != nil {
		return false, nil
	}

	return c.db.DoesKeyExist(bucket, hash.Bytes())
}

// Run is a simple loop that ensures we discard old data we do not need.
func (c *CrossReplayFilter) Run() {
	for {
		time.Sleep(time.Second * 5)
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

func (m *MarshalableUint32) MarshalBinary() ([]byte, error) {
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
