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
	if c.net.prom != nil {
		c.net.prom.CatRounds.Inc()
	}

	if err := c.writePeerCache(); err != nil {
		c.logger.WithError(err).Errorf("unable to write peer cache to disk")
	}

	peers := c.peers.Slice()

	toDrop := len(peers) - int(c.net.conf.DropTo) // current - target amount

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
		return nil
	}

	c.logger.Debugf("Received peer share from %s: %+v", peer, list)

	var res []Endpoint
	for _, ep := range list {
		if !ep.Valid() {
			c.logger.Infof("Peer %s tried to send us peer share with bad data: %s", peer, ep)
			return nil
		}

		if !c.isBannedEndpoint(ep) {
			res = append(res, ep)
		}
	}

	return res
}

func (c *controller) shuffleTrimShare(list []Endpoint) []Endpoint {
	if len(list) == 0 {
		return nil
	}
	list = append(list[:0:0], list...) // don't shuffle passed parameter
	c.net.rng.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	if uint(len(list)) > c.net.conf.PeerShareAmount {
		list = list[:c.net.conf.PeerShareAmount]
	}
	return list
}

func (c *controller) makePeerShare(exclude Endpoint) []Endpoint {
	var list []Endpoint
	peers := c.peers.Slice()

	for _, i := range c.net.rng.Perm(len(peers)) {
		if exclude.Equal(peers[i].Endpoint) {
			continue
		}
		list = append(list, peers[i].Endpoint)
		if uint(len(list)) >= c.net.conf.PeerShareAmount {
			break
		}
	}
	return list
}

// sharePeers shares the list of endpoints with a peer
func (c *controller) sharePeers(peer *Peer, list []Endpoint) {
	// convert to protocol
	payload, err := peer.prot.MakePeerShare(list)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal peer list to json")
		return
	}

	if len(payload) == 0 {
		c.logger.Debugf("No peers to share with %s", peer)
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

	async := make(chan *Parcel, 1)

	c.shareMtx.Lock()
	c.shareListener[peer.Hash] = async
	c.shareMtx.Unlock()

	defer func() {
		c.shareMtx.Lock()
		delete(c.shareListener, peer.Hash)
		c.shareMtx.Unlock()
	}()

	req := newParcel(TypePeerRequest, []byte("Peer Request"))
	peer.lastPeerSend = time.Now()
	peer.Send(req)

	select {
	case parcel := <-async:
		share := c.shuffleTrimShare(c.processPeerShare(peer, parcel))
		return share, nil
	case <-time.After(c.net.conf.PeerShareTimeout):
		return nil, fmt.Errorf("timeout")
	}
}

// catReplenish is the loop that brings the node up to the desired number of connections.
// Does nothing if we have enough peers, otherwise it sends a peer request to a random peer.
// The sources of new peers are, in order of priority:
// (0. Bootstrap peers saved from previous run)
// 1. Special peers
// 2. Seed peers
// 3. Random new peers shared by a random current peer
// 4. Random new peers from peers rejecting our connection
func (c *controller) catReplenish() {
	c.logger.Debug("Replenish loop started")
	defer c.logger.Debug("Replenish loop ended")

	canDial := func(ep Endpoint) bool {
		return !c.peers.Connected(ep) && !c.isBannedEndpoint(ep) && c.dialer.CanDial(ep)
	}

	// bootstrap
	if len(c.bootstrap) > 0 {
		c.logger.Infof("Attempting to connect to %d peers from bootstrap", len(c.bootstrap))
		for _, e := range c.bootstrap {
			if canDial(e) {
				c.Dial(e)
			}
		}
		c.bootstrap = nil
	}

	lastReseed := time.Now()

	for {
		select {
		case <-c.net.stopper:
			return
		default:
		}

		var connect []Endpoint
		if uint(c.peers.Total()) >= c.net.conf.TargetPeers {
			time.Sleep(time.Second)
			continue
		}

		// try special first
		for _, sp := range c.specialEndpoints {
			if canDial(sp) {
				connect = append(connect, sp)
			}
		}

		// reseed if necessary
		minReseed := c.net.conf.MinReseed
		if uint(c.seed.size()) < minReseed {
			minReseed = uint(c.seed.size()) - 1
		}

		if uint(c.peers.Total()) <= minReseed || time.Since(lastReseed) > c.net.conf.PeerReseedInterval {
			seeds := c.seed.retrieve()
			// shuffle to hit different seeds
			c.net.rng.Shuffle(len(seeds), func(i, j int) {
				seeds[i], seeds[j] = seeds[j], seeds[i]
			})
			for _, s := range seeds {
				if canDial(s) {
					connect = append(connect, s)
				}
			}
			lastReseed = time.Now()
		}

		if c.peers.Total() > 0 {
			rand := c.randomPeerConditional(func(p *Peer) bool {
				return time.Since(p.lastPeerSend) >= c.net.conf.PeerRequestInterval
			})
			if rand != nil {
				// error just means timeout of async request
				if eps, err := c.asyncPeerRequest(rand); err == nil {
					// pick random share from peer
					if len(eps) > 0 {
						el := c.net.rng.Intn(len(eps))
						ep := eps[el]
						if canDial(ep) {
							connect = append(connect, ep)
						}
					}
				}
			}
		}

		// if we connect to a peer that's full it gives us some alternatives
		// left unchecked, this can be a very long loop, therefore we are limiting it
		// sum(special, seeds) + 5 more
		attemptsLimit := len(connect) + 5
		var attempts int

		for len(connect) > 0 &&
			attempts < attemptsLimit &&
			uint(c.peers.Total()) < c.net.conf.TargetPeers {
			c.logger.Debugf("replenish loop with %d peers", len(connect))
			ep := connect[0]
			connect = connect[1:]

			if !canDial(ep) {
				continue
			}

			attempts++
			if p, alts := c.Dial(ep); p != nil {
				for _, alt := range alts {
					connect = append(connect, alt)
				}
			}
		}

		if attempts == 0 { // no peers and we exhausted special and seeds
			time.Sleep(time.Second)
		}
	}
}
