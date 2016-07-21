package controlPanel

import (
	"sort"
	"sync"
	"time"

	"github.com/FactomProject/factomd/p2p"
)

var AllConnections *ConnectionsMap

type ConnectionsMap struct {
	connected    map[string]p2p.ConnectionMetrics
	disconnected map[string]p2p.ConnectionMetrics

	sync.RWMutex
}

func (cm *ConnectionsMap) UpdateConnections(connections map[string]p2p.ConnectionMetrics) {
	cm.Lock()
	defer cm.Unlock()
	for key := range cm.connected { // Update Connected
		val, ok := connections[key]
		if ok {
			cm.connected[key] = val // Update Exisiting
		} else {
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
	var newMap map[string]p2p.ConnectionMetrics
	for k, v := range cm.connected {
		newMap[k] = v
	}
	return newMap
}

func (cm *ConnectionsMap) GetDisconnectedCopy() map[string]p2p.ConnectionMetrics {
	cm.Lock()
	defer cm.Unlock()
	newMap := make(map[string]p2p.ConnectionMetrics)
	for k, v := range cm.disconnected {
		newMap[k] = v
	}
	return newMap
}

func (cm *ConnectionsMap) Disconnect(key string, val *p2p.ConnectionMetrics) bool {
	cm.Lock()
	defer cm.Unlock()
	con, ok := cm.connected[key]
	if !ok {
		return false
	}
	delete(cm.connected, key)
	if val == nil {
		cm.disconnected[key] = con
	} else {
		cm.disconnected[key] = *val

	}
	return true
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
	Connected  bool
	Hash       string
	Connection p2p.ConnectionMetrics
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
		}
		item.Connected = true
		item.Hash = key
		list = append(list, *item)
	}
	for key := range cm.GetDisconnectedCopy() {
		item := new(ConnectionInfo)
		if newCon := cm.GetConnection(key); newCon == nil {
			continue
		} else {
			item.Connection = *newCon
		}
		item.Connected = false
		item.Hash = key
		list = append(list, *item)
	}
	var sortedList ConnectionInfoArray
	sortedList = list
	sort.Sort(sortedList)

	return sortedList
}

func manageConnections(connections chan map[string]p2p.ConnectionMetrics) {
	for {
		select {
		case newConnections := <-connections:
			AllConnections.UpdateConnections(newConnections)
			time.Sleep(1000 * time.Millisecond)
		default:
			time.Sleep(5 * time.Second)
		}
	}
}
