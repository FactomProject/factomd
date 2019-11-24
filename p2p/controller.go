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
	c.addSpecial(conf.Special)

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

func (c *controller) addSpecial(raw string) {
	if len(raw) == 0 {
		return
	}
	specialEndpoints := c.parseSpecial(raw)
	c.specialMtx.Lock()
	for _, ep := range specialEndpoints {
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
}

func (c *controller) manageData() {
	c.logger.Debug("Start manageData()")
	defer c.logger.Debug("Stop manageData()")
	for {
		select {
		case pp := <-c.peerData:
			parcel := pp.parcel
			peer := pp.peer

			if peer == nil && !parcel.IsApplicationMessage() { // peer disconnected between sending message and now
				c.logger.Debugf("Received parcel %s from peer not in system", parcel)
				continue
			}

			//c.logger.Debugf("Received parcel %s from %s", parcel, peer)
			switch parcel.Type {
			case TypePing:
				go func() {
					parcel := newParcel(TypePong, []byte("Pong"))
					peer.Send(parcel)
				}()
			case TypeMessage:
				//c.net.FromNetwork.Send(parcel)
				fallthrough
			case TypeMessagePart:
				parcel.Type = TypeMessage
				c.net.FromNetwork.Send(parcel)
			case TypePeerRequest:
				if time.Since(peer.lastPeerRequest) >= c.net.conf.PeerRequestInterval {
					peer.lastPeerRequest = time.Now()
					share := c.makePeerShare(peer.Endpoint)
					go c.sharePeers(peer, share)
				} else {
					c.logger.Warnf("peer %s sent a peer request too early", peer)
				}
			case TypePeerResponse:
				c.shareMtx.RLock()
				if f, ok := c.shareListener[peer.NodeID]; ok {
					f(parcel)
				}
				c.shareMtx.RUnlock()
			default:
				//not handled
			}
		}
	}
}

// Broadcast delivers a parcel to multiple connections specified by the fanout.
// A full broadcast sends the parcel to ALL connected peers
func (c *controller) Broadcast(parcel *Parcel, full bool) {
	if full {
		for _, p := range c.peers.Slice() {
			p.Send(parcel)
		}
		return
	}
	selection := c.selectBroadcastPeers(c.net.conf.Fanout)
	for _, p := range selection {
		p.Send(parcel)
	}
}

// ToPeer sends a parcel to a single peer, specified by their peer hash.
// If the hash is empty, a random connected peer will be chosen
func (c *controller) ToPeer(hash string, parcel *Parcel) {
	if hash == "" {
		if random := c.randomPeer(); random != nil {
			random.Send(parcel)
		} else {
			c.logger.Warnf("attempted to send parcel %s to a random peer but no peers are connected", parcel)
		}
	} else {
		p := c.peers.Get(hash)
		if p != nil {
			p.Send(parcel)
		}
	}
}

func (c *controller) randomPeers(count uint) []*Peer {
	peers := c.peers.Slice()
	if len(peers) == 0 {
		return nil
	}
	// not enough to randomize
	if uint(len(peers)) <= count {
		return peers
	}

	c.net.rng.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	return peers[:count]
}

func (c *controller) randomPeersConditional(count uint, condition func(*Peer) bool) []*Peer {
	peers := c.peers.Slice()
	if len(peers) == 0 {
		return nil
	}

	filtered := make([]*Peer, 0)
	for _, p := range peers {
		if condition(p) {
			filtered = append(filtered, p)
		}
	}
	// not enough to randomize
	if len(filtered) <= int(count) {
		return filtered
	}

	c.net.rng.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	return filtered[:count]
}

func (c *controller) randomPeer() *Peer {
	peers := c.randomPeers(1)
	if len(peers) == 1 {
		return peers[0]
	}
	return nil
}
