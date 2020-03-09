package p2p

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
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
				if old := c.peers.Get(pc.peer.Hash); old != nil {
					old.Stop()
					c.logger.Debugf("replacing old peer %s", pc.peer.Hash)
					c.peers.Remove(old)
				}
				err := c.peers.Add(pc.peer)
				if err != nil {
					c.logger.WithError(err).Errorf("Unable to add peer %s", pc.peer)
				}
			} else {
				c.peers.Remove(pc.peer)
			}
			if c.net.prom != nil {
				c.net.prom.Connections.Set(float64(c.peers.Total()))
				c.net.prom.Incoming.Set(float64(c.peers.Incoming()))
				c.net.prom.Outgoing.Set(float64(c.peers.Outgoing()))
			}
		case <-c.net.stopper:
			return
		}
	}
}

// preliminary check to see if we should accept an unknown connection
func (c *controller) allowIncoming(addr string) error {
	if c.isBannedIP(addr) {
		return fmt.Errorf("Address %s is banned", addr)
	}

	if uint(c.peers.Total()) >= c.net.conf.MaxIncoming && !c.isSpecialIP(addr) {
		return fmt.Errorf("Refusing incoming connection from %s because we are maxed out (%d of %d)", addr, c.peers.Total(), c.net.conf.MaxIncoming)
	}

	if c.net.conf.PeerIPLimitIncoming > 0 && uint(c.peers.Count(addr)) >= c.net.conf.PeerIPLimitIncoming {
		return fmt.Errorf("Rejecting %s due to per ip limit of %d", addr, c.net.conf.PeerIPLimitIncoming)
	}

	return nil
}

// handshakeIncoming performs the handshake maneouver for incoming connections.
// 	1. Determine their protocol from the first message they send
//	2. If we understand that protocol, validate that handshake
//	3. Reply with a handshake
//	4. Create a peer with that protocol
//
// If the incoming endpoint is banned, the connection is closed without alternatives.
// If the node is full, the connection is closed with alternatives.
//
// For more information, see the README
func (c *controller) handshakeIncoming(con net.Conn, ep Endpoint) error {
	if c.net.prom != nil {
		c.net.prom.Connecting.Inc()
		defer c.net.prom.Connecting.Dec()
	}

	con.SetDeadline(time.Now().Add(c.net.conf.HandshakeTimeout))

	// reject incoming connections based on host
	// the ep's port is our local port so we can't check that yet
	if err := c.allowIncoming(ep.IP); err != nil {
		share := c.makePeerShare(ep)  // they're not connected to us, so we don't have them in our system
		c.RejectWithShare(con, share) // closes con
		return fmt.Errorf("rejecting connection: %s", ep.IP)
	}

	// upgrade connection to a metrics connection
	metrics := NewMetricsReadWriter(con)
	prot, handshake, err := c.detectProtocolFromFirstMessage(metrics)
	if err != nil {
		con.Close()
		return fmt.Errorf("error detecting protocol: %v", err)
	}

	reply := newHandshake(c.net.conf, c.net.instanceID)
	reply.Version = handshake.Version
	if err := prot.SendHandshake(reply); err != nil {
		con.Close()
		return fmt.Errorf("unable to send handshake reply: %v", err)
	}

	// listenport has been validated in handshake.Valid
	ep.Port = handshake.ListenPort

	peer := newPeer(c.net, handshake.NodeID, ep, con, prot, metrics, true)
	c.peerStatus <- peerStatus{peer: peer, online: true}

	// a p2p1 node sends a peer request, so it needs to be processed
	if handshake.Type == TypePeerRequest {
		req := newParcel(TypePeerRequest, []byte("Peer Request"))
		req.Address = peer.Hash
		c.peerData <- peerParcel{peer: peer, parcel: req}
	}

	c.logger.Debugf("Incoming handshake success for peer %s, version %s", peer.Hash, peer.prot)
	return nil
}

// detectProtocolFromFirstMessage will listen for data to arrive on the ReadWriter and then attempt to interpret it.
// the existing protocol is only needed for nodes running v9 in order to bring
func (c *controller) detectProtocolFromFirstMessage(rw io.ReadWriter) (Protocol, *Handshake, error) {
	var prot Protocol
	var handshake *Handshake

	buffy := bufio.NewReader(rw)

	sig, err := buffy.Peek(4) // don't consume in case it's gob
	if err != nil {
		return nil, nil, err
	}

	if bytes.Equal(sig, V11Signature) {
		// pass the unread contents of buffy to the protocol so it's responsible for its own signature
		rw = struct {
			io.Reader
			io.Writer
		}{buffy, rw}
		prot = newProtocolV11(rw)
		hs, err := prot.ReadHandshake()
		if err != nil {
			return nil, nil, err
		}

		if err := hs.Valid(c.net.conf); err != nil {
			return nil, nil, err
		}
		handshake = hs
	} else { // default = gob
		decoder := gob.NewDecoder(buffy)
		encoder := gob.NewEncoder(rw)

		v9test := newProtocolV9(c.net.conf.Network, c.net.conf.NodeID, c.net.conf.ListenPort, decoder, encoder)
		hs, err := v9test.ReadHandshake()
		if err != nil {
			return nil, nil, err
		}

		if err := hs.Valid(c.net.conf); err != nil {
			return nil, nil, err
		}

		if hs.Version < c.net.conf.ProtocolVersionMinimum {
			return nil, nil, fmt.Errorf("protocol version %d below minimum of %d", hs.Version, c.net.conf.ProtocolVersionMinimum)
		}

		if hs.Version == 10 && c.net.conf.ProtocolVersion == 9 {
			hs.Version = 9
		}

		handshake = hs

		switch hs.Version {
		case 9:
			prot = v9test
		case 10:
			prot = newProtocolV10(decoder, encoder)
		default:
			return nil, nil, fmt.Errorf("unsupported protocol version %d", hs.Version)
		}
	}

	return prot, handshake, nil
}

// selectProtocol chooses the protocol based on the configuration.
// used to send the initial handshake when no other information is present.
func (c *controller) selectProtocol(rw io.ReadWriter) Protocol {
	switch c.net.conf.ProtocolVersion {
	case 11:
		return newProtocolV11(rw)
	case 10:
		decoder := gob.NewDecoder(rw)
		encoder := gob.NewEncoder(rw)
		return newProtocolV10(decoder, encoder)
	default:
		decoder := gob.NewDecoder(rw)
		encoder := gob.NewEncoder(rw)
		return newProtocolV9(c.net.conf.Network, c.net.conf.NodeID, c.net.conf.ListenPort, decoder, encoder)
	}
}

// handshakeOutgoing performs the handshake maneouver when dialing to remote nodes.
// 	1. Pick our desired protocol
//	2. Send a handshake
//	3. Figure out which protocol to use from the reply
//	4. Create a peer if a compatible protocol is established
//
// It is possible the endpoint will reject due to being full, in which
// case this function returns an error AND a list of alternate endpoints
//
// For more information, see the README
func (c *controller) handshakeOutgoing(con net.Conn, ep Endpoint) (*Peer, []Endpoint, error) {
	tmplogger := c.logger.WithField("endpoint", ep)
	timeout := time.Now().Add(c.net.conf.HandshakeTimeout)
	con.SetDeadline(timeout)

	handshake := newHandshake(c.net.conf, c.net.instanceID)
	metrics := NewMetricsReadWriter(con)
	desiredProt := c.selectProtocol(metrics)

	failfunc := func(err error) (*Peer, []Endpoint, error) {
		tmplogger.WithError(err).Debug("Handshake failed")
		con.Close()
		return nil, nil, err
	}
	if err := desiredProt.SendHandshake(handshake); err != nil {
		return failfunc(err)
	}

	prot, reply, err := c.detectProtocolFromFirstMessage(metrics)
	if err != nil {
		return failfunc(err)
	}

	if reply.Loopback == handshake.Loopback {
		return failfunc(fmt.Errorf("loopback"))
	}

	// this is required because a new protocol is instantiated in the above call
	// since V9Msg is already registered in the other end's gob, a new gob encoder
	// would try to register it again, causing an error on the other side
	// v10 is fine since it switches to a new V10Msg
	if v9new, ok := prot.(*ProtocolV9); ok {
		switch desiredProt.(type) {
		case *ProtocolV10:
			v9new.encoder = desiredProt.(*ProtocolV10).encoder
		case *ProtocolV9:
			v9new.encoder = desiredProt.(*ProtocolV9).encoder
		}
	}

	// dialed a node that's full
	if reply.Type == TypeRejectAlternative {
		con.Close()
		tmplogger.Debug("con rejected with alternatives")
		return nil, reply.Alternatives, fmt.Errorf("connection rejected")
	}

	peer := newPeer(c.net, reply.NodeID, ep, con, prot, metrics, false)
	c.peerStatus <- peerStatus{peer: peer, online: true}

	// a p2p1 node sends a peer request, so it needs to be processed
	if reply.Type == TypePeerRequest {
		req := newParcel(TypePeerRequest, []byte("Peer Request"))
		req.Address = peer.Hash
		c.peerData <- peerParcel{peer: peer, parcel: req}
	}

	c.logger.Debugf("Outgoing handshake success for peer %s, version %s", peer.Hash, peer.prot)

	return peer, nil, nil
}

// RejectWithShare rejects an incoming connection by sending them a handshake that provides
// them with alternative peers to connect to
func (c *controller) RejectWithShare(con net.Conn, share []Endpoint) error {
	defer con.Close() // we're rejecting, so always close

	prot := c.selectProtocol(con)

	handshake := newHandshake(c.net.conf, 0)
	handshake.Type = TypeRejectAlternative
	handshake.Alternatives = share

	return prot.SendHandshake(handshake)
}

// Dial attempts to connect to a remote endpoint.
// If the dial was not successful, it may return a list of alternate endpoints
// given by the remote host.
func (c *controller) Dial(ep Endpoint) (bool, []Endpoint) {
	if c.net.prom != nil {
		c.net.prom.Connecting.Inc()
		defer c.net.prom.Connecting.Dec()
	}

	c.logger.Debugf("Dialing to %s", ep)
	con, err := c.dialer.Dial(ep)
	if err != nil {
		c.logger.WithError(err).Infof("Failed to dial to %s", ep)
		return false, nil
	}

	peer, alternatives, err := c.handshakeOutgoing(con, ep)
	if err != nil { // handshake closes connection
		if err.Error() == "loopback" {
			c.logger.Debugf("Banning ourselves for 50 years")
			c.banEndpoint(ep, time.Hour*24*365*50) // ban for 50 years
			return false, nil
		}

		if len(alternatives) > 0 {
			c.logger.Debugf("Connection declined with alternatives from %s", ep)
			return false, alternatives
		}
		c.logger.WithError(err).Debugf("Handshake fail with %s", ep)
		return false, nil
	}

	c.logger.Debugf("Handshake success for peer %s, version %s", peer.Hash, peer.prot)
	return true, nil
}

// listen listens for incoming TCP connections and passes them off to handshake maneuver
func (c *controller) listen() {
	tmpLogger := c.logger.WithFields(log.Fields{"host": c.net.conf.BindIP, "port": c.net.conf.ListenPort})
	tmpLogger.Debug("controller.listen() starting up")

	addr := fmt.Sprintf("%s:%s", c.net.conf.BindIP, c.net.conf.ListenPort)

	l, err := NewLimitedListener(addr, c.net.conf.ListenLimit)
	if err != nil {
		tmpLogger.WithError(err).Error("controller.Start() unable to start limited listener")
		return
	}
	defer tmpLogger.Debug("controller.listen() stopping")
	c.listener = l

	go func() { // the listener doesn't play well with immediately stopping
		<-c.net.stopper
		c.listener.Close()
	}()

	// start permanent loop
	// terminates on program exit or when listener is closed
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			if ne, ok := err.(*net.OpError); ok && !ne.Timeout() {
				if !ne.Temporary() {
					tmpLogger.WithError(err).Error("controller.acceptLoop() error accepting")
					return
				}
			}
			continue
		}

		host, port, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			c.logger.WithError(err).Debugf("unable to parse address %s", conn.RemoteAddr().String())
			conn.Close()
			continue
		}

		ep, err := NewEndpoint(host, port) // this is the randomly assigned local port
		if err != nil {                    // should never happen for incoming
			c.logger.WithError(err).Errorf("failure to decode net address %s", conn.RemoteAddr().String())
			conn.Close()
			continue
		}

		go func() {
			if err := c.handshakeIncoming(conn, ep); err != nil {
				c.logger.WithError(err).Debug("incoming handshake failed")
			}
		}()
	}
}
