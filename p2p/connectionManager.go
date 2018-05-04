// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

// Connection Manager maintains the current list of connections and allows finding
// by peer hash or by the remote address.

type ConnectionManager struct {
	connections          map[string]*Connection // map of the connections indexed by peer hash
	connectionsByAddress map[string]*Connection // map of the connections indexed by peer address
}

func (cm *ConnectionManager) Init() *ConnectionManager {
	cm.connections = make(map[string]*Connection)
	cm.connectionsByAddress = make(map[string]*Connection)

	return cm
}

// TODO get rid of this
func (cm *ConnectionManager) All() map[string]*Connection {
	return cm.connections
}

// Get the number of all connections
func (cm *ConnectionManager) Count() int {
	return len(cm.connections)
}

// Get the number of online outgoing connections
func (cm *ConnectionManager) CountOutgoing() int {
	result := 0
	for _, connection := range cm.connections {
		if connection.IsOutGoing() && connection.IsOnline() {
			result++
		}
	}
	return result
}

// Get the number of online incoming connections
func (cm *ConnectionManager) CountIncoming() int {
	result := 0
	for _, connection := range cm.connections {
		if !connection.IsOutGoing() && connection.IsOnline() {
			result++
		}
	}
	return result
}

// Get the connection for a specified peer hash
func (cm *ConnectionManager) GetByPeerHash(peerHash string) (*Connection, bool) {
	connection, present := cm.connections[peerHash]
	return connection, present
}

// Get the connection for a specified address
func (cm *ConnectionManager) GetByAddress(address string) (*Connection, bool) {
	connection, present := cm.connectionsByAddress[address]
	return connection, present
}

// Add a new connection
func (cm *ConnectionManager) Add(connection *Connection) {
	cm.connections[connection.peer.Hash] = connection
	cm.connectionsByAddress[connection.peer.Address] = connection
}

// Remove an existing connection
func (cm *ConnectionManager) Remove(connection *Connection) {
	delete(cm.connections, connection.peer.Hash)
	delete(cm.connectionsByAddress, connection.peer.Address)
}

// Send a message to all the connections
func (cm *ConnectionManager) SendToAll(message interface{}) {
	for _, connection := range cm.connections {
		BlockFreeChannelSend(connection.SendChannel, message)
	}
}

// Update connection counts in Prometheus
func (cm *ConnectionManager) UpdatePrometheusMetrics() {
	p2pControllerNumConnections.Set(float64(cm.Count()))
	p2pControllerNumConnectionsByAddress.Set(float64(len(cm.connectionsByAddress)))
}

// TODO get rid of this
func (cm *ConnectionManager) UpdateConnectionAddressMap() {
	cm.connectionsByAddress = map[string]*Connection{}
	for _, value := range cm.connections {
		cm.connectionsByAddress[value.peer.Address] = value
	}
}
