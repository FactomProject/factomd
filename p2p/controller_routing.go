package p2p

import "time"

// route takes messages from ToNetwork and routes it to the appropriate peers
func (c *controller) route() {
	c.logger.Debug("Start controller.route()")
	defer c.logger.Debug("Stop controller.route()")
	for {
		// blocking read on ToNetwork, and c.stopRoute
		select {
		case <-c.net.stopper:
			return
		case parcel := <-c.net.toNetwork:
			switch parcel.Address {
			case FullBroadcast:
				for _, p := range c.peers.Slice() {
					p.Send(parcel)
				}

			case Broadcast:
				selection := c.selectBroadcastPeers(c.net.conf.Fanout)
				for _, p := range selection {
					p.Send(parcel)
				}

			case "", RandomPeer:
				if random := c.randomPeer(); random != nil {
					random.Send(parcel)
				} else {
					c.logger.Debugf("attempted to send parcel %s to a random peer but no peers are connected", parcel)
				}

			default:
				if p := c.peers.Get(parcel.Address); p != nil {
					p.Send(parcel)
				}
			}
		}
	}
}

// manageData processes parcels arriving from peers and responds appropriately.
// application messages are forwarded to the network channel.
func (c *controller) manageData() {
	c.logger.Debug("Start controller.manageData()")
	defer c.logger.Debug("Stop controller.manageData()")
	for {
		select {
		case <-c.net.stopper:
			return
		case pp := <-c.peerData:
			parcel := pp.parcel
			peer := pp.peer

			if peer == nil && !parcel.IsApplicationMessage() { // peer disconnected between sending message and now
				c.logger.Debugf("Received parcel %s from peer not in system", parcel)
				continue
			}

			//c.logger.Debugf("Received parcel %s from %s", parcel, peer)
			switch parcel.ptype {
			case TypePing:
				go func() {
					parcel := newParcel(TypePong, []byte("Pong"))
					peer.Send(parcel)
				}()
			case TypeMessage, TypeMessagePart:
				parcel.ptype = TypeMessage
				c.net.fromNetwork.Send(parcel)
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
				if async, ok := c.shareListener[peer.Hash]; ok {
					async <- parcel
				}
				c.shareMtx.RUnlock()
			default:
				//not handled
			}
		}
	}
}

func (c *controller) randomPeerConditional(condition func(*Peer) bool) *Peer {
	peers := c.peers.Slice()

	filtered := make([]*Peer, 0)
	for _, p := range peers {
		if condition(p) {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) > 0 {
		return filtered[c.net.rng.Intn(len(filtered))]
	}
	return nil
}

func (c *controller) randomPeer() *Peer {
	peers := c.peers.Slice()
	if len(peers) > 0 {
		return peers[c.net.rng.Intn(len(peers))]
	}
	return nil
}

func (c *controller) selectBroadcastPeers(fanout uint) []*Peer {
	peers := c.peers.Slice()

	// not enough to randomize
	if uint(len(peers)) <= fanout {
		return peers
	}

	var special []*Peer
	var regular []*Peer

	for _, p := range peers {
		if c.isSpecial(p.Endpoint) {
			special = append(special, p)
		} else {
			regular = append(regular, p)
		}
	}

	if uint(len(regular)) < fanout {
		return append(special, regular...)
	}

	c.net.rng.Shuffle(len(regular), func(i, j int) {
		regular[i], regular[j] = regular[j], regular[i]
	})

	return append(special, regular[:fanout]...)
}
