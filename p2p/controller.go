package p2p

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var controllerLogger = packageLogger.WithField("subpack", "controller")

// controller is responsible for managing Peers and Endpoints
type controller struct {
	net *Network

	peerStatus chan peerStatus
	peerData   chan peerParcel

	peers    *PeerStore
	dialer   *Dialer
	listener *LimitedListener

	specialMtx   sync.RWMutex
	specialCount int

	banMtx           sync.RWMutex
	bans             map[string]time.Time // (ip|ip:port) => time the ban ends
	special          map[string]bool      // (ip|ip:port) => bool
	specialEndpoints []Endpoint
	bootstrap        []Endpoint

	shareListener map[uint32]func(*Parcel)
	shareMtx      sync.RWMutex

	lastPeerDial time.Time
	lastPersist  time.Time

	counterMtx sync.RWMutex
	online     int
	connecting int

	lastRound    time.Time
	seed         *seed
	replenishing bool
	rounds       int // TODO make prometheus

	logger *log.Entry
}

// newController creates a new controller
// configuration is shared between the two
func newController(network *Network) (*controller, error) {
	var err error
	c := &controller{}
	c.net = network
	conf := network.conf // local var to reduce amount to type

	c.logger = controllerLogger.WithFields(log.Fields{
		"node":    conf.NodeName,
		"port":    conf.ListenPort,
		"network": conf.Network})
	c.logger.Debugf("Initializing Controller")

	c.dialer, err = NewDialer(conf.BindIP, conf.RedialInterval, conf.DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dialer: %v", err)
	}
	c.lastPersist = time.Now()

	c.peerStatus = make(chan peerStatus, 10) // TODO reconsider this value
	c.peerData = make(chan peerParcel, conf.ChannelCapacity)

	c.special = make(map[string]bool)
	c.shareListener = make(map[uint32]func(*Parcel))

	// CAT
	c.lastRound = time.Now()
	c.seed = newSeed(conf.SeedURL, conf.PeerReseedInterval)

	c.peers = NewPeerStore()
	c.setSpecial(conf.Special)

	if persist, err := c.loadPersist(); err != nil || persist == nil {
		c.logger.Infof("no valid bootstrap file found")
		c.bans = make(map[string]time.Time)
		c.bootstrap = nil
	} else if persist != nil {
		c.bans = persist.Bans
		c.bootstrap = persist.Bootstrap
	}

	if c.net.prom != nil {
		c.net.prom.KnownPeers.Set(float64(c.peers.Total()))
	}

	return c, nil
}

// ban bans the peer indicated by the hash as well as any other peer from that ip
// address
func (c *controller) ban(hash string, duration time.Duration) {
	peer := c.peers.Get(hash)
	if peer != nil {
		c.banMtx.Lock()

		end := time.Now().Add(duration)

		// there's a stronger ban in place already
		if existing, ok := c.bans[peer.Endpoint.IP]; ok && end.Before(existing) {
			end = existing
		}

		// ban both ip and ip:port
		c.bans[peer.Endpoint.IP] = end
		c.bans[peer.Endpoint.String()] = end

		for _, p := range c.peers.Slice() {
			if p.Endpoint.IP == peer.Endpoint.IP {
				peer.Stop()
			}
		}
		c.banMtx.Unlock()
	}
}

// ban a specific endpoint for a duration.
// to nullify a ban, use a duration of zero.
func (c *controller) banEndpoint(ep Endpoint, duration time.Duration) {
	c.banMtx.Lock()
	c.bans[ep.String()] = time.Now().Add(duration)
	c.banMtx.Unlock()

	if duration > 0 {
		for _, p := range c.peers.Slice() {
			if p.Endpoint == ep {
				p.Stop()
			}
		}
	}
}

func (c *controller) isBannedEndpoint(ep Endpoint) bool {
	c.banMtx.RLock()
	defer c.banMtx.RUnlock()
	return time.Now().Before(c.bans[ep.IP]) || time.Now().Before(c.bans[ep.String()])
}

func (c *controller) isBannedIP(ip string) bool {
	c.banMtx.RLock()
	defer c.banMtx.RUnlock()
	return time.Now().Before(c.bans[ip])
}

func (c *controller) isSpecial(ep Endpoint) bool {
	c.specialMtx.RLock()
	defer c.specialMtx.RUnlock()
	return c.special[ep.String()]
}

func (c *controller) isSpecialIP(ip string) bool {
	c.specialMtx.RLock()
	defer c.specialMtx.RUnlock()
	return c.special[ip]
}

func (c *controller) disconnect(hash string) {
	peer := c.peers.Get(hash)
	if peer != nil {
		peer.Stop()
	}
}

func (c *controller) setSpecial(raw string) {
	if len(raw) == 0 {
		c.specialEndpoints = nil
		return
	}
	c.specialEndpoints = c.parseSpecial(raw)
	c.specialMtx.Lock()
	for _, ep := range c.specialEndpoints {
		c.logger.Debugf("Registering special endpoint %s", ep)
		c.special[ep.String()] = true
		c.special[ep.IP] = true
	}
	c.specialCount = len(c.special)
	c.specialMtx.Unlock()
}

func (c *controller) parseSpecial(raw string) []Endpoint {
	var eps []Endpoint
	split := strings.Split(raw, ",")
	for _, item := range split {
		ep, err := ParseEndpoint(item)
		if err != nil {
			c.logger.Warnf("unable to determine host and port of special entry \"%s\"", item)
			continue
		}
		eps = append(eps, ep)
	}
	return eps
}

// Start starts the controller
// reads from the seed and connect to peers
func (c *controller) Start() {
	c.logger.Info("Starting the Controller")

	go c.run()          // cycle every 1s
	go c.manageData()   // blocking on data
	go c.manageOnline() // blocking on peer status changes
	go c.listen()       // blocking on tcp connections
	go c.catReplenish() // cycle every 1s
	go c.route()        // route data
}
