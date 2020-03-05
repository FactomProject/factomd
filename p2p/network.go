package p2p

import (
	"fmt"
	"math/rand"
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

	prom *Prometheus

	metricsHook func(pm map[string]PeerMetrics)

	rng        *rand.Rand // note: not thread safe for Read()
	instanceID uint64
	logger     *log.Entry

	stopper chan interface{}
}

var packageLogger = log.WithField("package", "p2p")

// NewNetwork initializes a new network with the given configuration.
// The passed Configuration is copied and cannot be modified afterwards.
// Does not start the network automatically.
func NewNetwork(conf Configuration) (*Network, error) {
	var err error

	n := new(Network)
	n.conf = &conf
	n.conf.Sanitize()
	n.stopper = make(chan interface{})

	n.logger = packageLogger.WithField("subpackage", "Network").WithField("node", n.conf.NodeName)

	if n.conf.EnablePrometheus {
		n.prom = new(Prometheus)
		n.prom.Setup()
	}

	if src, err := newLockSource(time.Now().UnixNano()); err != nil {
		return nil, err
	} else {
		n.rng = rand.New(src)
	}
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
	n.ToNetwork = newParcelChannel(n.conf.ChannelCapacity)
	n.FromNetwork = newParcelChannel(n.conf.ChannelCapacity)
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
// Listens to incoming connections on the specified port and connects to other peers
func (n *Network) Run() error {
	select {
	case <-n.stopper:
		return fmt.Errorf("unable to restart a network that has been stopped")
	default:
		n.logger.Infof("Starting a P2P Network with configuration %+v", n.conf)
		n.controller.Start()
		return nil
	}

}

// Stop shuts down the network
// Note that the network object will become unusable after it is stopped
func (n *Network) Stop() error {
	select {
	case <-n.stopper:
		return fmt.Errorf("network already stopped")
	default:
		n.logger.Info("Network.Stop() called")
		close(n.stopper)
		return nil
	}
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

// SetSpecial takes a set of ip addresses that should be treated as special.
// Network will always attempt to have a connection to a special peer.
// Format is a single line of ip addresses and ports, separated by semicolon, eg
// "127.0.0.1:8088;8.0.8.8:8088;192.168.0.1:8110"
func (n *Network) SetSpecial(raw string) {
	n.logger.Debugf("Received new list of special peers from application: %s", raw)
	go n.controller.setSpecial(raw)
}

// Total returns the number of active connections
func (n *Network) Total() int {
	return n.controller.peers.Total()
}

// Rounds returns the total number of CAT rounds that have occurred
func (n *Network) Rounds() int {
	return n.controller.rounds
}
