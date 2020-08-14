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
	BytesSentTotal     uint64
	BytesReceivedTotal uint64
	MessagesSent       uint64
	MessagesReceived   uint64
	BPSDown            float64
	BPSUp              float64
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
	connected    map[string]p2p.PeerMetrics
	disconnected map[string]p2p.PeerMetrics

	Totals AllConnectionsTotals
	Lock   sync.Mutex
}

func NewConnectionsMap() *ConnectionsMap {
	newCM := new(ConnectionsMap)
	newCM.Lock.Lock()
	defer newCM.Lock.Unlock()
	newCM.connected = map[string]p2p.PeerMetrics{}
	newCM.disconnected = map[string]p2p.PeerMetrics{}
	newCM.Totals = *(NewAllConnectionTotals())
	return newCM
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
		totals.BPSDown += peer.BPSDown
		totals.BPSUp += peer.BPSUp
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
		totals.BPSDown += peer.BPSDown
		totals.BPSUp += peer.BPSUp
	}

	cm.Lock.Lock()
	cm.Totals = *totals
	cm.Lock.Unlock()
}

func (cm *ConnectionsMap) UpdateConnections(connections map[string]p2p.PeerMetrics) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	cm.connected = connections
}

func hashPeerAddress(addr string) string {
	hash := sha256.Sum256([]byte(addr))
	hashStr := fmt.Sprintf("%x", hash)
	return hashStr
}

func (cm *ConnectionsMap) AddConnection(key string, val p2p.PeerMetrics) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	if _, ok := cm.disconnected[key]; ok {
		delete(cm.disconnected, key)
	}
	cm.connected[key] = val
}

func (cm *ConnectionsMap) RemoveConnection(key string) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	delete(cm.disconnected, key)
	delete(cm.connected, key)
}

func (cm *ConnectionsMap) Connect(key string, val *p2p.PeerMetrics) bool {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	disVal, ok := cm.disconnected[key]
	if ok {
		delete(cm.disconnected, key)
	}
	if val == nil && ok {
		cm.connected[key] = disVal
	} else if val != nil {
		cm.connected[key] = *val
	}
	return true
}

func (cm *ConnectionsMap) GetConnection(key string) *p2p.PeerMetrics {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	var ok bool
	var ret p2p.PeerMetrics
	ret, ok = cm.connected[key]
	if !ok {
		ret, ok = cm.disconnected[key]
		if !ok {
			return nil
		}

	}
	return &ret
}

func (cm *ConnectionsMap) GetConnectedCopy() map[string]p2p.PeerMetrics {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	newMap := map[string]p2p.PeerMetrics{}
	for k, v := range cm.connected {
		newMap[k] = v
	}
	return newMap
}

func (cm *ConnectionsMap) GetDisconnectedCopy() map[string]p2p.PeerMetrics {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	newMap := map[string]p2p.PeerMetrics{}
	for k, v := range cm.disconnected {
		newMap[k] = v
	}
	return newMap
}

func (cm *ConnectionsMap) Disconnect(key string, val *p2p.PeerMetrics) bool {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	conVal, ok := cm.connected[key]
	if ok {
		delete(cm.connected, key)
	}
	if val == nil && ok {
		cm.disconnected[key] = conVal
	} else if val != nil {
		cm.disconnected[key] = *val
	}
	return true
}

func (cm *ConnectionsMap) CleanDisconnected() int {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	count := 0
	for key := range cm.disconnected {
		delete(cm.disconnected, key)
		count++
		_ = key
	}
	return count
}

type ConnectionInfoArray []ConnectionInfo

func (slice ConnectionInfoArray) Len() int {
	return len(slice)
}

func (slice ConnectionInfoArray) Less(i, j int) bool {
	if slice[i].Connection.MomentConnected.Before(slice[j].Connection.MomentConnected) {
		return true
	}
	return false
}

func (slice ConnectionInfoArray) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type ConnectionInfo struct {
	Connected               bool
	Hash                    string // Hash of PeerHash (Peerhash contains illegal characters for html ID)
	Connection              p2p.PeerMetrics
	ConnectionTimeFormatted string
	PeerHash                string
}

// Used to send to front ent
func (cm *ConnectionsMap) SortedConnections() ConnectionInfoArray {
	list := make([]ConnectionInfo, 0)
	cmCopy := cm.GetConnectedCopy()
	for key := range cmCopy {
		item := new(ConnectionInfo)
		if newCon := cm.GetConnection(key); newCon == nil {
			continue
		} else {
			item.Connection = *newCon
			item.ConnectionTimeFormatted = FormatDuration(newCon.MomentConnected)
			item.Hash = hashPeerAddress(key)
			item.PeerHash = key
		}
		item.Connected = true
		list = append(list, *item)
	}
	disCopy := cm.GetDisconnectedCopy()
	for key := range disCopy {
		item := new(ConnectionInfo)
		if newCon := cm.GetConnection(key); newCon == nil {
			continue
		} else {
			item.Connection = *newCon
			item.ConnectionTimeFormatted = FormatDuration(newCon.MomentConnected)
			item.Hash = hashPeerAddress(key)
			item.PeerHash = key
		}
		item.Connected = false
		list = append(list, *item)

	}

	var sortedList ConnectionInfoArray
	sortedList = list
	sort.Sort(sort.Reverse(sortedList))
	cm.CleanDisconnected()
	return sortedList
}

func FormatDuration(initial time.Time) string {
	dif := time.Since(initial)
	if dif.Hours() > 24 {
		if int(dif.Hours()/24) == 1 {
			return fmt.Sprintf("%d%s", int(dif.Hours()/24), " day")
		}
		return fmt.Sprintf("%d%s", int(dif.Hours()/24), " days")
	} else if int(dif.Hours()) > 0 {
		if int(dif.Hours()) == 1 {
			return fmt.Sprintf("%d%s", int(dif.Hours()), " hr")
		}
		return fmt.Sprintf("%d%s", int(dif.Hours()), " hrs")
	} else if int(dif.Minutes()) > 0 {
		if int(dif.Minutes()) == 1 {
			return fmt.Sprintf("%d%s", int(dif.Minutes()), " min")
		}
		return fmt.Sprintf("%d%s", int(dif.Minutes()), " mins")
	} else {
		return fmt.Sprintf("%d%s", int(dif.Seconds()), " secs")
	}
}

// map[string]p2p.PeerMetrics
func manageConnections(connections chan map[string]p2p.PeerMetrics) {
	for {
		select {
		case connectionsMessage := <-connections:
			AllConnections.UpdateConnections(connectionsMessage)
			AllConnections.TallyTotals()
		default:
			time.Sleep(400 * time.Millisecond)
		}
	}
}
