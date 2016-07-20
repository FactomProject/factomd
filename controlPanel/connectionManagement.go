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
	Connection p2p.ConnectionMetrics
}

// Used to send to front ent
func (cm *ConnectionsMap) SortConnections() ConnectionInfoArray {
	cm.Lock()
	defer cm.Unlock()
	list := make([]ConnectionInfo, 0)
	for key := range cm.connected {
		item := new(ConnectionInfo)
		item.Connection = cm.connected[key]
		item.Connected = true
		list = append(list, *item)
	}
	for key := range cm.disconnected {
		item := new(ConnectionInfo)
		item.Connection = cm.disconnected[key]
		item.Connected = false
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
			// fmt.Printf("Channel Metrics: %+v", metrics)
			time.Sleep(500 * time.Millisecond)
		default:
			time.Sleep(5 * time.Second)
		}
	}
}
