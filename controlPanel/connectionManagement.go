package controlPanel

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/FactomProject/factomd/p2p"
)

var AllConnections *ConnectionsMap

type AllConnectionsTotals struct {
	PeerQualityAvg     int32
	BytesSentTotal     uint32
	BytesReceivedTotal uint32
	MessagesSent       uint32
	MessagesReceived   uint32
}

func NewAllConnectionTotals() *AllConnectionsTotals {
	a := new(AllConnectionsTotals)
	a.PeerQualityAvg = 0
	a.BytesSentTotal = 0
	a.BytesReceivedTotal = 0
	a.MessagesReceived = 0
	a.MessagesSent = 0

	return a
}

type ConnectionsMap struct {
	connected    map[string]p2p.ConnectionMetrics
	disconnected map[string]p2p.ConnectionMetrics

	totals AllConnectionsTotals
	sync.RWMutex
}

func (cm *ConnectionsMap) TallyTotals() {
	cons := cm.GetConnectedCopy()
	dis := cm.GetDisconnectedCopy()
	totals := NewAllConnectionTotals()
	var count int32 = 0
	var totalQuality int32 = 0
	for key := range cons {
		peer := cons[key]
		totals.BytesSentTotal += peer.BytesSent
		totals.BytesReceivedTotal += peer.BytesReceived
		totals.MessagesSent += peer.MessagesSent
		totals.MessagesReceived += peer.MessagesReceived
		count++
		totalQuality += peer.PeerQuality
	}
	if count == 0 {
		totals.PeerQualityAvg = 0
	} else {
		totals.PeerQualityAvg = totalQuality / count
	}
	for key := range dis {
		peer := dis[key]
		totals.BytesSentTotal += peer.BytesSent
		totals.BytesReceivedTotal += peer.BytesReceived
		totals.MessagesSent += peer.MessagesSent
		totals.MessagesReceived += peer.MessagesReceived
	}

	cm.totals = *totals
}

func (cm *ConnectionsMap) UpdateConnections(connections map[string]p2p.ConnectionMetrics) {
	cm.Lock()
	defer cm.Unlock()
	for key := range cm.connected { // Update Connected
		val, ok := connections[key]
		if ok {
			cm.connected[key] = val // Update Exisiting
		} else {
			delete(cm.connected, key)
			cm.Disconnect(key, &val) // No longer connected
		}
	}
	for key := range cm.disconnected { // Update Disconnected
		val, ok := connections[key]
		if ok {
			cm.Connect(key, &val) // Reconnected
		}
	}
	for key := range connections { // New Connections
		val, ok := cm.connected[key]
		if !ok {
			cm.connected[key] = val
		}
	}
}

func hashPeerAddress(addr string) string {
	hash := sha256.Sum256([]byte(addr))
	hashStr := fmt.Sprintf("%x", hash)
	return hashStr
}

func (cm *ConnectionsMap) AddConnection(key string, val p2p.ConnectionMetrics) {
	cm.Lock()
	defer cm.Unlock()
	if _, ok := cm.disconnected[key]; ok {
		delete(cm.disconnected, key)
	}
	cm.connected[key] = val
}

func (cm *ConnectionsMap) RemoveConnection(key string) {
	cm.Lock()
	defer cm.Unlock()
	delete(cm.disconnected, key)
	delete(cm.connected, key)
}

func (cm *ConnectionsMap) Connect(key string, val *p2p.ConnectionMetrics) bool {
	cm.Lock()
	defer cm.Unlock()
	dis, ok := cm.disconnected[key]
	if !ok {
		return false
	}
	delete(cm.disconnected, key)
	if val == nil {
		cm.connected[key] = dis
	} else {
		cm.connected[key] = *val

	}
	return true
}

func (cm *ConnectionsMap) GetConnection(key string) *p2p.ConnectionMetrics {
	cm.Lock()
	defer cm.Unlock()
	var ok bool
	var ret p2p.ConnectionMetrics
	ret, ok = cm.connected[key]
	if !ok {
		ret, ok = cm.disconnected[key]
		if !ok {
			return nil
		}

	}
	return &ret
}

func (cm *ConnectionsMap) GetConnectedCopy() map[string]p2p.ConnectionMetrics {
	cm.Lock()
	defer cm.Unlock()
	newMap := map[string]p2p.ConnectionMetrics{}
	for k, v := range cm.connected {
		newMap[k] = v
	}
	return newMap
}

func (cm *ConnectionsMap) GetDisconnectedCopy() map[string]p2p.ConnectionMetrics {
	cm.Lock()
	defer cm.Unlock()
	newMap := map[string]p2p.ConnectionMetrics{}
	for k, v := range cm.disconnected {
		newMap[k] = v
	}
	return newMap
}

func (cm *ConnectionsMap) Disconnect(key string, val *p2p.ConnectionMetrics) bool {
	cm.Lock()
	defer cm.Unlock()
	cm.disconnected[key] = *val
	return true
}

func (cm *ConnectionsMap) CleanDisconnected() int {
	cm.Lock()
	defer cm.Unlock()
	count := 0
	for key := range cm.disconnected {
		delete(cm.disconnected, key)
		count++
	}
	return count
}

type ConnectionInfoArray []ConnectionInfo

func (slice ConnectionInfoArray) Len() int {
	return len(slice)
}

func (slice ConnectionInfoArray) Less(i, j int) bool {
	if slice[i].Connected == true && slice[j].Connected == false {
		return true
	}
	return false
}

func (slice ConnectionInfoArray) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type ConnectionInfo struct {
	Connected               bool
	Hash                    string
	Connection              p2p.ConnectionMetrics
	ConnectionTimeFormatted string
}

// Used to send to front ent
func (cm *ConnectionsMap) SortedConnections() ConnectionInfoArray {
	list := make([]ConnectionInfo, 0)
	for key := range cm.GetConnectedCopy() {
		item := new(ConnectionInfo)
		if newCon := cm.GetConnection(key); newCon == nil {
			continue
		} else {
			item.Connection = *newCon
			hour, minute, second := newCon.MomentConnected.Clock()
			item.ConnectionTimeFormatted = fmt.Sprintf("%d:%d:%d", hour, minute, second)
			item.Hash = hashPeerAddress(key)
		}
		item.Connected = true
		list = append(list, *item)
	}
	for key := range cm.GetDisconnectedCopy() {
		item := new(ConnectionInfo)
		if newCon := cm.GetConnection(key); newCon == nil {
			continue
		} else {
			item.Connection = *newCon
			hour, minute, second := newCon.MomentConnected.Clock()
			item.ConnectionTimeFormatted = fmt.Sprintf("%d:%d:%d", hour, minute, second)
			item.Hash = hashPeerAddress(key)
		}
		item.Connected = false
		list = append(list, *item)

	}
	var sortedList ConnectionInfoArray
	sortedList = list
	sort.Sort(sortedList)
	cm.CleanDisconnected()
	return sortedList
}

// map[string]p2p.ConnectionMetrics
func manageConnections(connections chan interface{}) {
	for {
		select {
		case connectionsMessage := <-connections:
			switch connectionsMessage.(type) {
			case map[string]p2p.ConnectionMetrics:
				newConnections := connectionsMessage.(map[string]p2p.ConnectionMetrics)
				AllConnections.UpdateConnections(newConnections)
				AllConnections.TallyTotals()
			default: // drop that garbage
				fmt.Printf("Got garbage data on metrics channel: %+v", connectionsMessage)
			}
		default:
			time.Sleep(400 * time.Millisecond)
		}
	}
}
