package p2p

import (
	"fmt"
	"time"
)

// runs a single CAT round that persists peers and drops random connections.
// this function is triggered once a second by the controller.run function
func (c *controller) runCatRound() {
	if time.Since(c.lastRound) < c.net.conf.RoundTime {
		return
	}
	c.lastRound = time.Now()
	c.logger.Debug("Cat Round")
	c.rounds++

	c.persistPeerFile()

	peers := c.peers.Slice()

	toDrop := len(peers) - int(c.net.conf.Drop) // current - target amount

	if toDrop > 0 {
		perm := c.net.rng.Perm(len(peers))

		dropped := 0
		for _, i := range perm {
			if c.isSpecial(peers[i].Endpoint) {
				continue
			}
			peers[i].Stop()
			dropped++
			if dropped >= toDrop {
				break
			}
		}
	}
}

// processPeers processes a peer share response
func (c *controller) processPeerShare(peer *Peer, parcel *Parcel) []Endpoint {
	list, err := peer.prot.ParsePeerShare(parcel.Payload)

	if err != nil {
		c.logger.WithError(err).Warnf("Failed to unmarshal peer share from peer %s", peer)
	}

	c.logger.Debugf("Received peer share from %s: %+v", peer, list)

	var res []Endpoint
	for _, p := range list {
		if !p.Valid() {
			c.logger.Infof("Peer %s tried to send us peer share with bad data: %s", peer, p)
			return nil
		}
		ep, err := NewEndpoint(p.IP, p.Port)
		if err != nil {
			c.logger.WithError(err).Infof("Unable to register endpoint %s:%s from peer %s", p.IP, p.Port, peer)
		} else if !c.isBannedEndpoint(ep) {
			res = append(res, ep)
		}
	}

	if c.net.prom != nil {
		c.net.prom.KnownPeers.Set(float64(c.peers.Total()))
	}

	return res
}

func (c *controller) trimShare(list []Endpoint, shuffle bool) []Endpoint {
	if len(list) == 0 {
		return nil
	}
	if shuffle {
		c.net.rng.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	}
	if uint(len(list)) > c.net.conf.PeerShareAmount {
		list = list[:c.net.conf.PeerShareAmount]
	}
	return list
}

func (c *controller) makePeerShare(ep Endpoint) []Endpoint {
	var list []Endpoint
	tmp := c.peers.Slice()
	var i int

	cmp := ep.String()
	for _, i = range c.net.rng.Perm(len(tmp)) {
		if tmp[i].Endpoint.String() == cmp {
			continue
		}
		list = append(list, tmp[i].Endpoint)
		if uint(len(tmp)) >= c.net.conf.PeerShareAmount {
			break
		}
	}
	return list
}

// sharePeers creates a list of peers to share and sends it to peer
func (c *controller) sharePeers(peer *Peer, list []Endpoint) {
	if peer == nil {
		return
	}
	// CAT select n random active peers
	payload, err := peer.prot.MakePeerShare(list)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal peer list to json")
		return
	}
	c.logger.Debugf("Sharing %d peers with %s", len(list), peer)
	parcel := newParcel(TypePeerResponse, payload)
	peer.Send(parcel)
}

// this function is only intended to be run single-threaded inside the replenish loop
// it works by creating a closure that contains a channel specific for this call
// the closure is called in controller.manageData
// if there is no response from the peer after 5 seconds, it times out
func (c *controller) asyncPeerRequest(peer *Peer) ([]Endpoint, error) {
	c.shareMtx.Lock()

	var share []Endpoint
	async := make(chan bool, 1)
	f := func(parcel *Parcel) {
		share = c.trimShare(c.processPeerShare(peer, parcel), true)
		async <- true
	}
	c.shareListener[peer.NodeID] = f
	c.shareMtx.Unlock()

	defer func() {
		c.shareMtx.Lock()
		delete(c.shareListener, peer.NodeID)
		c.shareMtx.Unlock()
	}()

	req := newParcel(TypePeerRequest, []byte("Peer Request"))
	peer.Send(req)

	select {
	case <-async:
	case <-time.After(time.Second * 5):
		return nil, fmt.Errorf("timeout")
	}

	return share, nil
}

// catReplenish is the loop that brings the node up to the desired number of connections.
// Does nothing if we have enough peers, otherwise it sends a peer request to a random peer.
func (c *controller) catReplenish() {
	c.logger.Debug("Replenish loop started")
	defer c.logger.Debug("Replenish loop ended")

	deny := func(ep Endpoint) bool {
		return c.peers.Connected(ep) || c.isBannedEndpoint(ep) || !c.dialer.CanDial(ep)
	}

	// bootstrap
	if len(c.bootstrap) > 0 {
		c.logger.Infof("Attempting to connect to %d peers from bootstrap", len(c.bootstrap))
		for _, e := range c.bootstrap {
			if !deny(e) {
				_, _ = c.Dial(e)
			}
		}
		c.bootstrap = nil
	}

	lastReseed := time.Now()

	for {
		var connect []Endpoint
		if uint(c.peers.Total()) >= c.net.conf.Target {
			time.Sleep(time.Second)
			continue
		}

		// reseed if necessary
		min := c.net.conf.MinReseed
		if uint(c.seed.size()) < min {
			min = uint(c.seed.size()) - 1
		}

		// try special first
		for _, sp := range c.specialEndpoints {
			if deny(sp) {
				continue
			}
			connect = append(connect, sp)
		}

		if uint(c.peers.Total()) <= min || time.Since(lastReseed) > c.net.conf.PeerReseedInterval {
			seeds := c.seed.retrieve()
			// shuffle to hit different seeds
			c.net.rng.Shuffle(len(seeds), func(i, j int) {
				seeds[i], seeds[j] = seeds[j], seeds[i]
			})
			for _, s := range seeds {
				if deny(s) {
					continue
				}
				connect = append(connect, s)
			}
			lastReseed = time.Now()
		}

		// if we connect to a peer that's full it gives us some alternatives
		// left unchecked, this can be a very long loop, therefore we are limiting it
		// sum(special, seeds) + 5 more
		var attemptsLimit = len(connect) + 5

		if c.peers.Total() > 0 {
			rand := c.randomPeersConditional(1, func(p *Peer) bool {
				return time.Since(p.lastPeerSend) >= c.net.conf.PeerRequestInterval
			})
			if len(rand) > 0 {
				p := rand[0]
				// error just means timeout of async request
				p.lastPeerSend = time.Now()
				if eps, err := c.asyncPeerRequest(p); err == nil {
					// pick random share from peer
					if len(eps) > 0 {
						el := c.net.rng.Intn(len(eps))
						ep := eps[el]
						if !deny(ep) {
							connect = append(connect, ep)
						}
					}
				}
			}
		}

		var ep Endpoint
		var attempts int
		for len(connect) > 0 && attempts < attemptsLimit {
			ep = connect[0]
			connect = connect[1:]

			if deny(ep) {
				continue
			}

			attempts++
			if ok, alts := c.Dial(ep); !ok {
				for _, alt := range alts {
					connect = append(connect, alt)
				}
			}

			if uint(c.peers.Total()) >= c.net.conf.Target {
				break
			}
		}

		connect = nil

		if attempts == 0 { // no peers and we exhausted special and seeds
			time.Sleep(time.Second)
		}
	}
}
