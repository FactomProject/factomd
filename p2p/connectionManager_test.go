package p2p

import (
	"testing"
)

func TestConnectionManagerGettingByHash(t *testing.T) {
	cm := new(ConnectionManager).Init()

	nonExisting, present := cm.GetByHash("non existing hash")

	if present || nonExisting != nil {
		t.Error("GetByHash does not handle empty connection manager")
	}

	peer := newPeer("127.0.0.1", "8888", RegularPeer)
	connection := newConnection(peer)
	cm.Add(connection)

	existing, present := cm.GetByHash(peer.Hash)
	if !present || existing != connection {
		t.Error("GetByHash does not find the added connection")
	}

	cm.Remove(connection)
	existing, present = cm.GetByHash(peer.Hash)
	if present || existing != nil {
		t.Error("GetByHash finds the removed connection")
	}
}

func TestConnectionManagerCountConnections(t *testing.T) {
	cm := new(ConnectionManager).Init()

	if cm.Count() != 0 {
		t.Error("Count does not handle empty connection manager")
	}

	cm.Add(newConnection(newPeer("1.2.3.4", "8888", RegularPeer)))
	cm.Add(newConnection(newPeer("2.3.4.5", "8888", RegularPeer)))
	cm.Add(newConnection(newPeer("3.4.5.6", "8888", SpecialPeerCmdLine)))

	totalCount := cm.Count()
	if totalCount != 3 {
		t.Errorf("Connections are not counted correctly: %d", totalCount)
	}

	specialCount := cm.CountIf(func(c *Connection) bool { return c.peer.IsSpecial() })
	if specialCount != 1 {
		t.Errorf("Special connections are not counted correctly: %d", specialCount)
	}
}

func TestConnectionManagerConnectedTo(t *testing.T) {
	cm := new(ConnectionManager).Init()

	conn1 := newConnection(newPeer("1.2.3.4", "8888", RegularPeer))
	conn2 := newConnection(newPeer("2.3.4.5", "8888", RegularPeer))
	conn3 := newConnection(newPeer("1.2.3.4", "8889", RegularPeer))

	if cm.ConnectedTo("some address") {
		t.Error("Connected to reports we're connected, while the connection manager is empty")
	}

	cm.Add(conn1)
	if !cm.ConnectedTo("1.2.3.4") {
		t.Error("Connected to reports we're not connected to an added connection")
	}

	cm.Add(conn2)
	if !cm.ConnectedTo("2.3.4.5") {
		t.Error("Connected to reports we're not connected to an added connection")
	}

	if cm.ConnectedTo("some address") {
		t.Error("Connected to reports we're connected to not added connection")
	}

	cm.Add(conn3)
	if !cm.ConnectedTo("1.2.3.4") {
		t.Error("Added a second connection with the same address, we should still be connected")
	}

	cm.Remove(conn2)
	if cm.ConnectedTo("2.3.4.5") {
		t.Errorf("Removed connection, we shouldn't be connected to this address %v", cm)
	}

	cm.Remove(conn1)
	if !cm.ConnectedTo("1.2.3.4") {
		t.Error("Removed one connection, but the other connection for this address still there, so we should still be connected")
	}

	cm.Remove(conn3)
	if cm.ConnectedTo("1.2.3.4") {
		t.Error("Removed all connections, but we're still connected?")
	}
}

func TestConnectionManagerRemoveIsIdempotent(t *testing.T) {
	cm := new(ConnectionManager).Init()
	conn := newConnection(newPeer("1.2.3.4", "8888", RegularPeer))

	cm.Add(conn)
	existing, present := cm.GetByHash(conn.peer.Hash)
	if !present || existing != conn {
		t.Error("GetByHash does not find the added connection")
	}

	cm.Remove(conn)
	existing, present = cm.GetByHash(conn.peer.Hash)
	if present || existing != nil {
		t.Error("GetByHash found a connection after removal")
	}

	// make sure remove works even if there is no connection
	cm.Remove(conn)
	existing, present = cm.GetByHash(conn.peer.Hash)
	if present || existing != nil {
		t.Error("GetByHash found a connection after second removal")
	}
}

func TestConnectionManagerGetRandomEmpty(t *testing.T) {
	cm := new(ConnectionManager).Init()

	if cm.GetRandom() != nil {
		t.Error("GetRandom should return nil if there is nothing in the manager")
	}
}

func TestConnectionManagerGetRandomSingleRegular(t *testing.T) {
	cm := new(ConnectionManager).Init()
	connection := newActiveConnection(newPeer("1", "1", RegularPeer))

	cm.Add(connection)

	if cm.GetRandom() != connection {
		t.Error("GetRandom should get a single regular peer")
	}
}

func TestConnectionManagerGetRandomSingleSpecial(t *testing.T) {
	cm := new(ConnectionManager).Init()
	connection := newActiveConnection(newPeer("1", "1", RegularPeer))

	cm.Add(connection)

	if cm.GetRandom() != connection {
		t.Error("GetRandom should get a single special peer")
	}
}

func TestConnectionManagerGetRandomRegularSingle(t *testing.T) {
	cm := new(ConnectionManager).Init()
	conn1 := newActiveConnection(newPeer("1", "1", RegularPeer))
	conn2 := newActiveConnection(newPeer("2", "1", SpecialPeerCmdLine))
	conn3 := newActiveConnection(newPeer("3", "1", SpecialPeerCmdLine))

	cm.Add(conn1)
	cm.Add(conn2)
	cm.Add(conn3)

	// we ask for 5 random nodes, while we have only 1 to draw from, but the
	// function should still return the single one
	random := cm.GetRandomRegular(5)

	if len(random) != 1 {
		t.Error("GetRandomRegular returned more than 1 connection")
	}

	if random[0] != conn1 {
		t.Error("GetRandomRegular was expected to return a single regular connection")
	}

}

func newPeer(address string, port string, peerType uint8) *Peer {
	return new(Peer).Init(address, port, 100, peerType, 0)
}

func newConnection(peer *Peer) *Connection {
	return new(Connection).InitWithConn(nil, *peer)
}

func newActiveConnection(peer *Peer) *Connection {
	connection := newConnection(peer)

	// pretend we're online and we've received something
	connection.state = ConnectionOnline
	connection.metrics.BytesReceived += 2

	return connection
}
