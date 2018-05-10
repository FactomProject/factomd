// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

// Connection Manager maintains the current list of connections and allows finding
// by peer hash or by the remote address.

import (
	"fmt"
	"math/rand"
)

type ConnectionManager struct {
	connections          map[string]*Connection     // connections indexed by peer hash
	connectionsByAddress map[string]map[string]bool // peer hashes indexed by the address (we can have multiple connections to the same address)
}

func (cm *ConnectionManager) Init() *ConnectionManager {
	cm.connections = make(map[string]*Connection)
	cm.connectionsByAddress = make(map[string]map[string]bool)

	return cm
}

// Get the map of all connections by the peer hash.
func (cm *ConnectionManager) All() map[string]*Connection {
	return cm.connections
}

// Get the number of all connections
func (cm *ConnectionManager) Count() int {
	return len(cm.connections)
}

// Get the number of connections meeting the predicate
func (cm *ConnectionManager) CountIf(pred func(*Connection) bool) int {
	result := 0
	for _, connection := range cm.connections {
		if pred(connection) {
			result++
		}
	}
	return result
}

// Get the connection for a specified peer hash.
func (cm *ConnectionManager) GetByHash(peerHash string) (*Connection, bool) {
	connection, present := cm.connections[peerHash]
	return connection, present
}

// Checks if we are already connected to a specified address.
func (cm *ConnectionManager) ConnectedTo(address string) bool {
	_, present := cm.connectionsByAddress[address]
	return present
}

// Add a new connection.
func (cm *ConnectionManager) Add(connection *Connection) {
	if _, present := cm.connections[connection.peer.Hash]; present {
		// we should be checking whether we are already connected to this peer,
		// so something went wrong
		panic(fmt.Sprintf("Duplicated peer in connection manager: %s", connection.peer.Hash))
	}
	cm.connections[connection.peer.Hash] = connection
	cm.addToConnectionsByAddress(connection)
}

// Remove an existing connection.
func (cm *ConnectionManager) Remove(connection *Connection) {
	delete(cm.connections, connection.peer.Hash)
	cm.removeFromConnectionsByAddress(connection)
}

// Send a message to all the connections.
func (cm *ConnectionManager) SendToAll(message interface{}) {
	for _, connection := range cm.connections {
		BlockFreeChannelSend(connection.SendChannel, message)
	}
}

// Update connection counts in Prometheus.
func (cm *ConnectionManager) UpdatePrometheusMetrics() {
	p2pControllerNumConnections.Set(float64(cm.Count()))
	p2pControllerNumConnectionsByAddress.Set(float64(len(cm.connectionsByAddress)))
}

// Get a single random connection from all the online and active connection we have,
// returns nil if none are found.
func (cm *ConnectionManager) GetRandom() *Connection {
	onlineActive := cm.getMatching(func(c *Connection) bool {
		return c.IsOnline() && c.metrics.BytesReceived > 0
	})

	if len(onlineActive) == 0 {
		// no peer to send to
		return nil
	}

	return onlineActive[rand.Intn(len(onlineActive))]
}

// Get connections for all online, active regular peers, but in random order.
func (cm *ConnectionManager) GetAllRegular() []*Connection {
	selection := cm.getMatching(func(c *Connection) bool {
		return c.IsOnline() && !c.peer.IsSpecial() && c.metrics.BytesReceived > 0
	})

	shuffle(len(selection), func(i, j int) {
		selection[i], selection[j] = selection[j], selection[i]
	})

	return selection
}

// Get a set of random connections from all the online, active regular peers we have.
func (cm *ConnectionManager) GetRandomRegular(sampleSize int) []*Connection {
	if sampleSize <= 0 {
		return make([]*Connection, 0)
	}

	selection := cm.GetAllRegular()
	resultSize := min(sampleSize, len(selection))
	return selection[:resultSize]
}

func (cm *ConnectionManager) addToConnectionsByAddress(connection *Connection) {
	addressBucket, exists := cm.connectionsByAddress[connection.peer.Address]

	if !exists {
		addressBucket = make(map[string]bool)
		cm.connectionsByAddress[connection.peer.Address] = addressBucket
	}

	addressBucket[connection.peer.Hash] = true
}

func (cm *ConnectionManager) removeFromConnectionsByAddress(connection *Connection) {
	addressBucket, exists := cm.connectionsByAddress[connection.peer.Address]

	if !exists {
		return
	}

	delete(addressBucket, connection.peer.Hash)

	if len(addressBucket) == 0 {
		delete(cm.connectionsByAddress, connection.peer.Address)
	}
}

func (cm *ConnectionManager) getMatching(pred func(*Connection) bool) []*Connection {
	matching := make([]*Connection, 0, len(cm.connections))

	for _, connection := range cm.connections {
		if pred(connection) {
			matching = append(matching, connection)
		}
	}

	return matching
}

// This is the implementation of Shuffle with go 1.10, but included here to allow go 1.9 to
// still compile our code.
func shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(rand.Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(rand.Int31n(int32(i + 1)))
		swap(i, j)
	}
}

func min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}
