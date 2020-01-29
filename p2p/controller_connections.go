package p2p

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// manageOnline listens to peerStatus updates sent out by peers
// if a peer notifies it's going offline, it will be removed
// if a peer notifies it's coming online, existing peers with the same hash are removed
func (c *controller) manageOnline() {
	c.logger.Debug("Start manageOnline()")
	defer c.logger.Debug("Stop manageOnline()")
	for {
		select {
		case pc := <-c.peerStatus:
			if pc.online {
				old := c.peers.Get(pc.peer.Hash)
				if old != nil {
					old.Stop()
					c.logger.Debugf("removing old peer %s", pc.peer.Hash)
					c.peers.Remove(old)
				}
				err := c.peers.Add(pc.peer)
				if err != nil {
					c.logger.Errorf("Unable to add peer %s to peer store because an old peer still exists", pc.peer)
				}
			} else {
				c.peers.Remove(pc.peer)
			}
			if c.net.prom != nil {
				c.net.prom.Connections.Set(float64(c.peers.Total()))
				//c.net.prom.Unique.Set(float64(c.peers.Unique()))
				c.net.prom.Incoming.Set(float64(c.peers.Incoming()))
				c.net.prom.Outgoing.Set(float64(c.peers.Outgoing()))
			}
		}
	}
}

// preliminary check to see if we should accept an unknown connection
func (c *controller) allowIncoming(addr string) error {
	if c.isBannedIP(addr) {
		return fmt.Errorf("Address %s is banned", addr)
	}

	if uint(c.peers.Total()) >= c.net.conf.Incoming && !c.isSpecialIP(addr) {
		return fmt.Errorf("Refusing incoming connection from %s because we are maxed out (%d of %d)", addr, c.peers.Total(), c.net.conf.Incoming)
	}

	if c.net.conf.PeerIPLimitIncoming > 0 && uint(c.peers.Count(addr)) >= c.net.conf.PeerIPLimitIncoming {
		return fmt.Errorf("Rejecting %s due to per ip limit of %d", addr, c.net.conf.PeerIPLimitIncoming)
	}

	return nil
}

// what to do with a new tcp connection
func (c *controller) handleIncoming(con net.Conn) {
	if c.net.prom != nil {
		c.net.prom.Connecting.Inc()
		defer c.net.prom.Connecting.Dec()
	}

	host, _, err := net.SplitHostPort(con.RemoteAddr().String())
	if err != nil {
		c.logger.WithError(err).Debugf("Unable to parse address %s", con.RemoteAddr().String())
		con.Close()
		return
	}

	// port is overriden during handshake, use default port as temp port
	ep, err := NewEndpoint(host, c.net.conf.ListenPort)
	if err != nil { // should never happen for incoming
		c.logger.WithError(err).Debugf("Unable to decode address %s", host)
		con.Close()
		return
	}

	// if we're full, give them alternatives
	if err = c.allowIncoming(host); err != nil {
		c.logger.WithError(err).Infof("Rejecting connection")
		share := c.makePeerShare(ep)  // they're not connected to us, so we don't have them in our system
		c.RejectWithShare(con, share) // closes con
		return
	}

	peer := newPeer(c.net, c.peerStatus, c.peerData)
	// we are never expecting a reject-alternate for incoming connections
	if _, err := peer.StartWithHandshake(ep, con, true); err != nil {
		c.logger.WithError(err).Debugf("Handshake failed for address %s, stopping", ep)
		peer.Stop()
		return
	}

	c.logger.Debugf("Incoming handshake success for peer %s, version %s", peer.Hash, peer.prot.Version())

	if c.isBannedEndpoint(peer.Endpoint) {
		c.logger.Debugf("Peer %s is banned, disconnecting", peer.Hash)
		peer.Stop()
	}
}

// RejectWithShare rejects an incoming connection by sending them a handshake that provides
// them with alternative peers to connect to
func (c *controller) RejectWithShare(con net.Conn, share []Endpoint) error {
	defer con.Close() // we're rejecting, so always close

	payload, err := json.Marshal(share)
	if err != nil {
		return err
	}

	handshake := newHandshake(c.net.conf, payload)
	handshake.Header.Type = TypeRejectAlternative

	// only push the handshake, don't care what they send us
	encoder := gob.NewEncoder(con)
	con.SetWriteDeadline(time.Now().Add(c.net.conf.HandshakeTimeout))
	err = encoder.Encode(handshake)
	if err != nil {
		return err
	}

	return nil
}

// Dial attempts to connect to a remote endpoint.
// If the dial was not successful, it may return a list of alternate endpoints
// given by the remote host.
func (c *controller) Dial(ep Endpoint) (bool, []Endpoint) {
	if c.net.prom != nil {
		c.net.prom.Connecting.Inc()
		defer c.net.prom.Connecting.Dec()
	}

	if ep.Port == "" {
		ep.Port = c.net.conf.ListenPort
		c.logger.Debugf("Dialing to %s (with no previously known port)", ep)
	} else {
		c.logger.Debugf("Dialing to %s", ep)
	}

	con, err := c.dialer.Dial(ep)
	if err != nil {
		c.logger.WithError(err).Infof("Failed to dial to %s", ep)
		return false, nil
	}

	peer := newPeer(c.net, c.peerStatus, c.peerData)
	if share, err := peer.StartWithHandshake(ep, con, false); err != nil {
		if err.Error() == "loopback" {
			c.logger.Debugf("Banning ourselves for 50 years")
			c.banEndpoint(ep, time.Hour*24*365*50) // ban for 50 years
		} else if len(share) > 0 {
			c.logger.Debugf("Connection declined with alternatives from %s", ep)
			return false, share
		} else {
			c.logger.WithError(err).Debugf("Handshake fail with %s", ep)
		}
		peer.Stop()
		return false, nil
	}

	c.logger.Debugf("Handshake success for peer %s, version %s", peer.Hash, peer.prot.Version())
	return true, nil
}

// listen listens for incoming TCP connections and passes them off to handshake maneuver
func (c *controller) listen() {
	tmpLogger := c.logger.WithFields(log.Fields{"address": c.net.conf.BindIP, "port": c.net.conf.ListenPort})
	tmpLogger.Debug("controller.listen() starting up")

	addr := fmt.Sprintf("%s:%s", c.net.conf.BindIP, c.net.conf.ListenPort)

	l, err := NewLimitedListener(addr, c.net.conf.ListenLimit)
	if err != nil {
		tmpLogger.WithError(err).Error("controller.Start() unable to start limited listener")
		return
	}

	c.listener = l

	// start permanent loop
	// terminates on program exit or when listener is closed
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			if ne, ok := err.(*net.OpError); ok && !ne.Timeout() {
				if !ne.Temporary() {
					tmpLogger.WithError(err).Warn("controller.acceptLoop() error accepting")
				}
			}
			continue
		}

		go c.handleIncoming(conn)
	}
}
