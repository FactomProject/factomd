package p2p

import "time"

// route Takes messages from the network's ToNetwork channel and routes via the appropriate function
func (c *controller) route() {
	for {
		// blocking read on ToNetwork, and c.stopRoute
		select {
		case message := <-c.net.ToNetwork:
			switch message.Address {
			case FullBroadcast:
				c.Broadcast(message, true)
			case Broadcast:
				c.Broadcast(message, false)
			case RandomPeer:
				c.ToPeer("", message)
			default:
				c.ToPeer(message.Address, message)
			}
		}
	}
}

// manageData processes parcels arriving from peers and responds appropriately.
// application messages are forwarded to the network channel.
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

func (c *controller) selectBroadcastPeers(count uint) []*Peer {
	peers := c.peers.Slice()

	// not enough to randomize
	if uint(len(peers)) <= count {
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

	if uint(len(regular)) < count {
		return append(special, regular...)
	}

	c.net.rng.Shuffle(len(regular), func(i, j int) {
		regular[i], regular[j] = regular[j], regular[i]
	})

	return append(special, regular[:count]...)
}
