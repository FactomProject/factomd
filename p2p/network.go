package p2p

import (
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

// Network is the main access point for outside applications.
//
// ToNetwork is the channel over which to send parcels to the network layer
//
// FromNetwork is the channel that gets filled with parcels arriving from the network layer
type Network struct {
	ToNetwork   ParcelChannel
	FromNetwork ParcelChannel

	conf       *Configuration
	controller *controller
	seed       *seed

	prom *Prometheus

	metricsHook func(pm map[string]PeerMetrics)
	measure     *Measure

	rng        *rand.Rand
	instanceID uint64
	logger     *log.Entry

	globalCloser chan interface{}
	fatalError   chan error
}

var packageLogger = log.WithField("package", "p2p")

// DebugMessage is temporary
func (n *Network) DebugMessage() (string, string, int) {
	hv := ""
	s := n.controller.peers.Slice()
	r := fmt.Sprintf("\nONLINE: (%d/%d/%d)\n", len(s), n.conf.Target, n.conf.Max)
	count := len(s)
	for _, p := range s {

		metrics := p.GetMetrics()
		r += fmt.Sprintf("\tPeer %s (MPS %.2f/%.2f) (BPS %.2f/%.2f) (Cap %.2f)\n", p.String(), metrics.MPSDown, metrics.MPSUp, metrics.BPSDown, metrics.BPSUp, metrics.Capacity)
		edge := ""
		if n.conf.NodeID < 4 || p.NodeID < 4 {
			min := n.conf.NodeID
			if p.NodeID < min {
				min = p.NodeID
			}
			if min != 0 {
				color := []string{"red", "green", "blue"}[min-1]
				edge = fmt.Sprintf(" {color:%s, weight=3}", color)
			}
		}
		if p.IsIncoming {
			hv += fmt.Sprintf("%s -> %s:%s%s\n", p.Endpoint, n.conf.BindIP, n.conf.ListenPort, edge)
		} else {
			hv += fmt.Sprintf("%s:%s -> %s%s\n", n.conf.BindIP, n.conf.ListenPort, p.Endpoint, edge)
		}
	}
	known := ""
	/*for _, ip := range n.controller.endpoints.IPs() {
		known += ip.Address + " "
	}*/
	r += "\nKNOWN:\n" + known

	/*banned := ""
	for ip, time := range n.controller.endpoints.Bans {
		banned += fmt.Sprintf("\t%s %s\n", ip, time)
	}
	r += "\nBANNED:\n" + banned*/
	return r, hv, count
}

// DebugServer is temporary
func DebugServer(n *Network) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug", func(rw http.ResponseWriter, req *http.Request) {
		a, _, _ := n.DebugMessage()
		rw.Write([]byte(a))
	})

	mux.HandleFunc("/stats", func(rw http.ResponseWriter, req *http.Request) {
		out := ""
		out += fmt.Sprintf("Channels\n")
		out += fmt.Sprintf("\tToNetwork: %d / %d\n", len(n.ToNetwork), cap(n.ToNetwork))
		out += fmt.Sprintf("\tFromNetwork: %d / %d\n", len(n.FromNetwork), cap(n.FromNetwork))
		out += fmt.Sprintf("\tpeerData: %d / %d\n", len(n.controller.peerData), cap(n.controller.peerData))
		out += fmt.Sprintf("\nPeers (%d)\n", n.controller.peers.Total())

		slice := n.controller.peers.Slice()
		sort.Slice(slice, func(i, j int) bool {
			return slice[i].connected.Before(slice[j].connected)
		})

		for _, p := range slice {
			out += fmt.Sprintf("\t%s\n", p.Endpoint)
			out += fmt.Sprintf("\t\tsend: %d / %d\n", len(p.send), cap(p.send))
			m := p.GetMetrics()
			out += fmt.Sprintf("\t\tBytesReceived: %d\n", m.BytesReceived)
			out += fmt.Sprintf("\t\tBytesSent: %d\n", m.BytesSent)
			out += fmt.Sprintf("\t\tMessagesSent: %d\n", m.MessagesSent)
			out += fmt.Sprintf("\t\tMessagesReceived: %d\n", m.MessagesReceived)
			out += fmt.Sprintf("\t\tMomentConnected: %s\n", m.MomentConnected)
			out += fmt.Sprintf("\t\tPeerQuality: %d\n", m.PeerQuality)
			out += fmt.Sprintf("\t\tIncoming: %v\n", m.Incoming)
			out += fmt.Sprintf("\t\tLastReceive: %s\n", m.LastReceive)
			out += fmt.Sprintf("\t\tLastSend: %s\n", m.LastSend)
			out += fmt.Sprintf("\t\tMPS Down: %.2f\n", m.MPSDown)
			out += fmt.Sprintf("\t\tMPS Up: %.2f\n", m.MPSUp)
			out += fmt.Sprintf("\t\tBPS Down: %.2f\n", m.BPSDown)
			out += fmt.Sprintf("\t\tBPS Up: %.2f\n", m.BPSUp)
			out += fmt.Sprintf("\t\tCapacity: %.2f\n", m.Capacity)
			out += fmt.Sprintf("\t\tDropped: %d\n", m.Dropped)
		}

		rw.Write([]byte(out))
	})

	go http.ListenAndServe("localhost:8070", mux)
}

// NewNetwork initializes a new network with the given configuration.
// The passed Configuration is copied and cannot be modified afterwards.
// Does not start the network automatically.
func NewNetwork(conf Configuration) (*Network, error) {
	var err error
	myconf := conf // copy
	myconf.Sanitize()

	n := new(Network)
	n.fatalError = make(chan error)

	n.logger = packageLogger.WithField("subpackage", "Network").WithField("node", conf.NodeName)

	n.conf = &myconf
	if n.conf.EnablePrometheus {
		n.prom = new(Prometheus)
		n.prom.Setup()
	}
	n.measure = NewMeasure(time.Second * 15)
	n.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	// generate random instanceid for loopback detection
	n.instanceID = n.rng.Uint64()

	// turn nodename into nodeid
	if n.conf.NodeID == 0 {
		n.conf.NodeID = StringToUint32(n.conf.NodeName)
	}

	n.controller, err = newController(n)
	if err != nil {
		return nil, err
	}
	n.ToNetwork = newParcelChannel(conf.ChannelCapacity)
	n.FromNetwork = newParcelChannel(conf.ChannelCapacity)
	return n, nil
}

func (n *Network) GetInfo() Info {
	peers := n.controller.peers.Slice()
	pDown, pUp, rDown, rUp := 0.0, 0.0, 0.0, 0.0
	for _, p := range peers {
		metrics := p.GetMetrics()
		pDown += metrics.MPSDown
		pUp += metrics.MPSUp
		rDown += metrics.BPSDown
		rUp += metrics.BPSUp
	}
	return Info{
		Peers:     n.controller.peers.Total(),
		Receiving: pDown,
		Sending:   pUp,
		Download:  rDown,
		Upload:    rUp,
	}
}

func (n *Network) GetPeerMetrics() map[string]PeerMetrics {
	return n.controller.makeMetrics()
}

// SetMetricsHook allows you to read peer metrics.
// Gets called approximately once a second and transfers the metrics
// of all CONNECTED peers in the format "identifying hash" -> p2p.PeerMetrics
func (n *Network) SetMetricsHook(f func(pm map[string]PeerMetrics)) {
	n.metricsHook = f
}

// Run starts the network.
// Listens to incoming connections on the specified port
// and connects to other peers
func (n *Network) Run() {
	n.logger.Infof("Starting a P2P Network with configuration %+v", n.conf)

	n.controller.Start() // this will get peer manager ready to handle incoming connections
	//DebugServer(n)
	go n.route()
}

func (n *Network) Stop() {
	// TODO implement
	// close stop channel
	// stop all peers
}

// Ban removes a peer as well as any other peer from that address
// and prevents any connection being established for the amount of time
// set in the configuration (default one week)
func (n *Network) Ban(hash string) {
	n.logger.Debugf("Received ban for %s from application", hash)
	go n.controller.ban(hash, n.conf.ManualBan)
}

// Disconnect severs connection for a specific peer. They are free to
// connect again afterward
func (n *Network) Disconnect(hash string) {
	n.logger.Debugf("Received disconnect for %s from application", hash)
	go n.controller.disconnect(hash)
}

// AddSpecial takes a set of ip addresses that should be treated as special.
// Network will always attempt to have a connection to a special peer.
// Format is a single line of ip addresses and ports, separated by semicolon, eg
// "127.0.0.1:8088;8.8.8.8;192.168.0.1:8110"
//
// The port is optional and the entire ip will be considered special if no port is
// provided
func (n *Network) AddSpecial(raw string) {
	n.logger.Debugf("Received new list of special peers from application: %s", raw)
	go n.controller.addSpecial(raw)
}

// Total returns the number of active connections
func (n *Network) Total() int {
	return n.controller.peers.Total()
}

// Rounds returns the total number of CAT rounds that have occurred
func (n *Network) Rounds() int {
	return n.controller.rounds
}

// route Takes messages from the network's ToNetwork channel and routes it
// to the controller via the appropriate function
func (n *Network) route() {
	for {
		// blocking read on ToNetwork, and c.stopRoute
		select {
		case message := <-n.ToNetwork:
			switch message.Address {
			case FullBroadcast:
				n.controller.Broadcast(message, true)
			case Broadcast:
				n.controller.Broadcast(message, false)
			case RandomPeer:
				n.controller.ToPeer("", message)
			default:
				n.controller.ToPeer(message.Address, message)
			}
		}
	}
}
