// +build all 

package controlPanel_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/p2p"
)

var _ = fmt.Sprintf("")

func TestFormatDuration(t *testing.T) {
	initial := time.Now().Add(-49 * time.Hour)
	if FormatDuration(initial) != "2 days" {
		t.Errorf("Time display incorrect : days")
	}

	initial = time.Now().Add(-25 * time.Hour)
	if FormatDuration(initial) != "1 day" {
		t.Errorf("Time display incorrect : days")
	}

	initial = time.Now().Add(-23 * time.Hour)
	if FormatDuration(initial) != "23 hrs" {
		t.Errorf("Time display incorrect : hrs")
	}

	initial = time.Now().Add(-1 * time.Hour)
	if FormatDuration(initial) != "1 hr" {
		t.Errorf("Time display incorrect : hr")
	}

	initial = time.Now().Add(-59 * time.Minute)
	if FormatDuration(initial) != "59 mins" {
		t.Errorf("Time display incorrect : mins")
	}

	initial = time.Now().Add(-1 * time.Minute)
	if FormatDuration(initial) != "1 min" {
		t.Errorf("Time display incorrect : min")
	}

	initial = time.Now().Add(-30 * time.Second)
	if FormatDuration(initial) != "30 secs" {
		t.Errorf("Time display incorrect : secs")
	}
}

func TestTallyTotals(t *testing.T) {
	cm := NewConnectionsMap()
	var i uint32
	for i = 0; i < 10; i++ {
		cm.Connect(fmt.Sprintf("%d", i), NewP2PConnection(i, i, i, i, fmt.Sprintf("%d", i), i))
	}
	for i = 10; i < 20; i++ {
		cm.Disconnect(fmt.Sprintf("%d", i), NewP2PConnection(i, i, i, i, fmt.Sprintf("%d", i), i))
	}
	cm.TallyTotals()
	cm.Lock.Lock()
	if cm.Totals.BytesSentTotal != 190 {
		t.Errorf("Byte Sent does not match")
	}
	if cm.Totals.BytesReceivedTotal != 190 {
		t.Errorf("Byte Received does not match")
	}
	if cm.Totals.MessagesSent != 190 {
		t.Errorf("Msg Sent does not match")
	}
	if cm.Totals.MessagesReceived != 190 {
		t.Errorf("Msg Received does not match")
	}
	if cm.Totals.PeerQualityAvg != 4 {
		t.Errorf("Peer Quality does not match %d", cm.Totals.PeerQualityAvg)
	}
	cm.Lock.Unlock()

	for key := range cm.GetConnectedCopy() {
		cm.RemoveConnection(key)
	}
	for key := range cm.GetDisconnectedCopy() {
		cm.RemoveConnection(key)
	}
	cm.TallyTotals()
	cm.Lock.Lock()
	if cm.Totals.BytesSentTotal != 0 {
		t.Errorf("Byte Sent does not match")
	}
	if cm.Totals.BytesReceivedTotal != 0 {
		t.Errorf("Byte Received does not match")
	}
	if cm.Totals.MessagesSent != 0 {
		t.Errorf("Msg Sent does not match")
	}
	if cm.Totals.MessagesReceived != 0 {
		t.Errorf("Msg Received does not match")
	}
	if cm.Totals.PeerQualityAvg != 0 {
		t.Errorf("Peer Quality does not match %d", cm.Totals.PeerQualityAvg)
	}
	cm.Lock.Unlock()

	AllConnections = cm
	SortedConnectionString()
	AllConnectionsString()
}

func PopulateConnectionChan(total uint32, connections chan interface{}) {
	time.Sleep(3 * time.Second)
	var i uint32
	temp := make(map[string]p2p.ConnectionMetrics)
	for i = 0; i < total; i++ {
		peer := NewSeededP2PConnection(i)
		if i%2 == 0 {
			peer.MomentConnected = time.Now().Add(-(time.Duration(i)) * time.Hour)
		} else {
			peer.MomentConnected = time.Now().Add(-(time.Duration(i)) * time.Minute)
		}
		temp["{"+peer.PeerAddress+"}"] = *peer
	}
	connections <- temp
}

func TestAccessors(t *testing.T) {
	cm := NewConnectionsMap()

	// Test Disconnect
	for count := uint32(0); count < 2; count++ {
		peer := NewSeededP2PConnection(count)
		cm.Disconnect(peer.PeerAddress, peer)
	}

	for count := uint32(0); count < 2; count++ {
		cp := cm.GetDisconnectedCopy()
		if len(cp) != 2 {
			t.Error("Should have 2 Disconnections")
		}
	}

	// Test Connect
	for count := uint32(0); count < 2; count++ {
		peer := NewSeededP2PConnection(count)
		cm.AddConnection(peer.PeerAddress, *peer)
	}

	for count := uint32(0); count < 2; count++ {
		cp := cm.GetConnectedCopy()
		if len(cp) != 2 {
			t.Error("Should have 2 connections")
		}

		dp := cm.GetDisconnectedCopy()
		if len(dp) != 0 {
			t.Error("Should have 0 Disconnections")
		}
	}

	// Connect with nil, but exists in disconnections
	for count := uint32(0); count < 2; count++ {
		peer := NewSeededP2PConnection(count)
		cm.Disconnect(peer.PeerAddress, peer)
	}

	for count := uint32(0); count < 2; count++ {
		peer := NewSeededP2PConnection(count)
		cm.Connect(peer.PeerAddress, nil)
	}

	for count := uint32(0); count < 2; count++ {
		cp := cm.GetConnectedCopy()
		if len(cp) != 2 {
			t.Error("Should have 2 Connections")
		}
	}

	// Disconnect with nil, but exists in connections
	for count := uint32(0); count < 2; count++ {
		peer := NewSeededP2PConnection(count)
		cm.Disconnect(peer.PeerAddress, nil)
	}

	for count := uint32(0); count < 2; count++ {
		cp := cm.GetDisconnectedCopy()
		if len(cp) != 2 {
			t.Error("Should have 2 Connections")
		}
	}

}

// Absurd map accessing
func TestConcurrency(t *testing.T) {
	cm := NewConnectionsMap()
	connectionMap := make(map[string]p2p.ConnectionMetrics)
	var count uint32
	for count = 0; count < 100; count++ {
		peer := NewSeededP2PConnection(count)

		connectionMap[peer.PeerAddress] = *peer
	}
	var i uint32
	for i = 0; i < 100; i++ {
		// Random Connections
		go func() {
			randPeers := make([]p2p.ConnectionMetrics, 0)
			for ii := 0; ii < 10; ii++ {
				randPeer := NewRandomP2PConnection()
				cm.Connect(randPeer.PeerAddress, randPeer)

				randPeer2 := NewRandomP2PConnection()
				cm.AddConnection(randPeer2.PeerAddress, *randPeer2)

				randPeers = append(randPeers, *randPeer)
				randPeers = append(randPeers, *randPeer2)
			}
			for c, peer := range randPeers {
				switch c % 2 {
				case 0:
					cm.Disconnect(peer.PeerAddress, cm.GetConnection(peer.PeerAddress))
				case 1:
					cm.RemoveConnection(peer.PeerAddress)
				}

			}
			cm.CleanDisconnected()
		}()

		go func() {
			randPeers := make([]p2p.ConnectionMetrics, 0)
			var ii uint32
			for ii = 0; ii < 10; ii++ {
				randPeer1 := NewSeededP2PConnection(ii)
				cm.Connect(randPeer1.PeerAddress, randPeer1)

				randPeer2 := NewSeededP2PConnection(ii)
				cm.AddConnection(randPeer2.PeerAddress, *randPeer2)

				randPeers = append(randPeers, *randPeer1)
				randPeers = append(randPeers, *randPeer2)
			}
			for _, peer := range randPeers {
				cm.Disconnect(peer.PeerAddress, cm.GetConnection(peer.PeerAddress))
			}
			cm.CleanDisconnected()
		}()

		go func() {
			for ii := 0; ii < 50; ii++ {
				cm.TallyTotals()
				cm.SortedConnections()
				cm.GetDisconnectedCopy()
				cm.GetConnectedCopy()
			}
		}()
		go func() {
			var ii uint32
			for ii = 30; ii < 60; ii++ {
				p := NewRandomP2PConnection()
				ps := NewSeededP2PConnection(ii)
				cm.AddConnection(p.PeerAddress, *p)
				cm.AddConnection(ps.PeerAddress, *p)

			}
		}()
		go func() {
			var ii uint32
			for ii = 30; ii < 60; ii++ {
				p := NewRandomP2PConnection()
				ps := NewSeededP2PConnection(ii)
				cm.RemoveConnection(p.PeerAddress)
				cm.RemoveConnection(ps.PeerAddress)

			}
		}()

		go func() {
			var ii uint32
			for ii = 100; ii < 120; ii++ {
				cm.UpdateConnections(connectionMap)
			}
		}()
		// Sharing connections
	}
}

func NewRandomP2PConnection() *p2p.ConnectionMetrics {
	con := NewP2PConnection(rand.Uint32(), rand.Uint32(), rand.Uint32(), rand.Uint32(), string(rand.Uint32()), rand.Uint32())
	return con
}

func NewSeededP2PConnection(seed uint32) *p2p.ConnectionMetrics {
	con := NewP2PConnection(seed, seed, seed, seed, fmt.Sprintf("%d", seed), seed)
	return con
}

func NewP2PConnection(bs uint32, br uint32, ms uint32, mr uint32, addr string, pq uint32) *p2p.ConnectionMetrics {
	pc := new(p2p.ConnectionMetrics)
	pc.MomentConnected = time.Now()
	pc.BytesSent = bs
	pc.BytesReceived = br
	pc.MessagesSent = ms
	pc.MessagesReceived = mr
	pc.PeerAddress = "10.1.1" + addr
	pc.PeerQuality = int32(pq)

	return pc
}

/* For testing peer table sorts. TODO: Remove
if len(cmCopy) > 1 {
		// Testing

		for i := 0; i < 10; i++ {
			s := new(ConnectionInfo)
			s.Connected = true
			s.Hash = hashPeerAddress(fmt.Sprintf("%d min", i))
			s.ConnectionTimeFormatted = fmt.Sprintf("%d min", i)
			s.PeerHash = hashPeerAddress(fmt.Sprintf("%d min", i))
			for key := range cmCopy {
				s.Connection = cmCopy[key]
				s.Connection.PeerAddress = fmt.Sprintf("%d min", i)
				break
			}
			list = append(list, *s)
		}
		for i := 10; i < 15; i++ {
			s := new(ConnectionInfo)
			s.Connected = true
			s.Hash = hashPeerAddress(fmt.Sprintf("%d hr", i))
			s.ConnectionTimeFormatted = fmt.Sprintf("%d hr", i)
			s.PeerHash = hashPeerAddress(fmt.Sprintf("%d hr", i))
			for key := range cmCopy {
				s.Connection = cmCopy[key]
				s.Connection.PeerAddress = fmt.Sprintf("%d hr", i)
				break
			}
			list = append(list, *s)
		}
		for i := 15; i < 20; i++ {
			s := new(ConnectionInfo)
			s.Connected = true
			s.Hash = hashPeerAddress(fmt.Sprintf("%d sec", i))
			s.ConnectionTimeFormatted = fmt.Sprintf("%d sec", i)
			s.PeerHash = hashPeerAddress(fmt.Sprintf("%d sec", i))
			for key := range cmCopy {
				s.Connection = cmCopy[key]
				s.Connection.PeerAddress = fmt.Sprintf("%d sec", i)
				break
			}
			list = append(list, *s)
		}
		for i := 20; i < 25; i++ {
			s := new(ConnectionInfo)
			s.Connected = true
			s.Hash = hashPeerAddress(fmt.Sprintf("%d sec", i))
			s.ConnectionTimeFormatted = fmt.Sprintf("%d sec", i)
			s.PeerHash = hashPeerAddress(fmt.Sprintf("%d sec", i))
			for key := range cmCopy {
				s.Connection = cmCopy[key]
				s.Connection.PeerAddress = fmt.Sprintf("%d sec", i)
				break
			}
			list = append(list, *s)
		}
		//End
	}
*/
