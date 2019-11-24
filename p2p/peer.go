package p2p

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var peerLogger = packageLogger.WithField("subpack", "peer")

// Peer is an active connection to an endpoint in the network.
// Represents one lifetime of a connection and should not be restarted
type Peer struct {
	net     *Network
	conn    net.Conn
	metrics ReadWriteCollector
	prot    Protocol

	// current state, read only "constants" after the handshake
	IsIncoming bool
	Endpoint   Endpoint
	NodeID     uint32 // a nonce to distinguish multiple nodes behind one endpoint
	Hash       string // This is more of a connection ID than hash right now.

	stopper sync.Once
	stop    chan bool

	lastPeerRequest  time.Time
	peerShareAsk     bool
	peerShareDeliver chan *Parcel
	lastPeerSend     time.Time

	// communication channels
	send       ParcelChannel   // parcels from Send() are added here
	status     chan peerStatus // the controller's notification channel
	data       chan peerParcel // the controller's data channel
	registered bool

	// Metrics
	metricsMtx           sync.RWMutex
	connected            time.Time
	lastReceive          time.Time // Keep track of how long ago we talked to the peer.
	lastSend             time.Time // Keep track of how long ago we talked to the peer.
	totalParcelsSent     uint64
	totalParcelsReceived uint64
	totalBytesSent       uint64
	totalBytesReceived   uint64
	bpsDown, bpsUp       float64
	mpsDown, mpsUp       float64
	dropped              uint64

	// logging
	logger *log.Entry
}

func newPeer(net *Network, status chan peerStatus, data chan peerParcel) *Peer {
	p := &Peer{}
	p.net = net
	p.status = status
	p.data = data

	p.logger = peerLogger.WithFields(log.Fields{
		"node": net.conf.NodeName,
	})
	p.stop = make(chan bool, 1)

	p.peerShareDeliver = nil

	p.logger.Debugf("Creating blank peer")
	return p
}

func (p *Peer) bootstrapProtocol(hs *Handshake, conn net.Conn, decoder *gob.Decoder, encoder *gob.Encoder) error {
	v := hs.Header.Version
	if v > p.net.conf.ProtocolVersion {
		v = p.net.conf.ProtocolVersion
	}

	///fmt.Printf("@@@ %d %+v %s\n", v, hs.Header, conn.RemoteAddr())
	switch v {
	case 9:
		v9 := new(ProtocolV9)
		v9.init(p, decoder, encoder)
		p.prot = v9

		// v9 starts with a peer request
		hsParcel := new(Parcel)
		hsParcel.Address = hs.Header.TargetPeer
		hsParcel.Payload = hs.Payload
		hsParcel.Type = TypePeerRequest
		/*	TODO at this point, peer is not fully initialized
			if !p.deliver(hsParcel) {
				return fmt.Errorf("unable to deliver peer request to controller")
			}*/
	case 10:
		v10 := new(ProtocolV10)
		v10.init(p, decoder, encoder)
		p.prot = v10
	default:
		return fmt.Errorf("unknown protocol version %d", v)
	}
	return nil
}

// StartWithHandshake performs a basic handshake maneouver to establish the validity
// of the connection. Immediately sends a Peer Request upon connection and waits for the
// response, which can be any parcel. The information in the header is verified, especially
// the port.
//
// The handshake ensures that ALL peers have a valid Port field to start with.
// If there is no reply within the specified HandshakeTimeout config setting, the process
// fails
//
// For outgoing connections, it is possible the endpoint will reject due to being full, in which
// case this function returns an error AND a list of alternate endpoints
func (p *Peer) StartWithHandshake(ep Endpoint, con net.Conn, incoming bool) ([]Endpoint, error) {
	tmplogger := p.logger.WithField("addr", ep.IP)
	timeout := time.Now().Add(p.net.conf.HandshakeTimeout)

	nonce := []byte(fmt.Sprintf("%x", p.net.instanceID))

	// upgrade connection to a metrics connection
	p.metrics = NewMetricsReadWriter(con)
	p.conn = con

	handshake := newHandshake(p.net.conf, nonce)
	decoder := gob.NewDecoder(p.metrics) // pipe gob through the metrics writer
	encoder := gob.NewEncoder(p.metrics)
	con.SetWriteDeadline(timeout)
	con.SetReadDeadline(timeout)
	//fmt.Printf("@@@ %+v %s\n", handshake.Header, con.RemoteAddr())

	failfunc := func(err error) ([]Endpoint, error) {
		tmplogger.WithError(err).Debug("Handshake failed")
		p.conn.Close()
		return nil, err
	}

	err := encoder.Encode(handshake)
	if err != nil {
		return failfunc(fmt.Errorf("Failed to send handshake to incoming connection"))
	}

	var reply Handshake
	err = decoder.Decode(&reply)
	if err != nil {
		return failfunc(fmt.Errorf("Failed to read handshake from incoming connection"))
	}

	// check basic structure
	if err = reply.Valid(p.net.conf); err != nil {
		return failfunc(err)
	}

	// loopback detection
	if string(reply.Payload) == string(nonce) {
		return failfunc(fmt.Errorf("loopback"))
	}

	if err = p.bootstrapProtocol(&reply, con, decoder, encoder); err != nil {
		return failfunc(err)
	}

	if reply.Header.Type == TypeRejectAlternative {
		con.Close()
		tmplogger.Debug("con rejected with alternatives")
		var rawShare []Endpoint
		err := json.Unmarshal(reply.Payload, &rawShare)
		if err != nil {
			return nil, fmt.Errorf("unable to parse alternatives: %s", err.Error())
		}

		filtered := make([]Endpoint, 0, len(rawShare))
		for _, ep := range rawShare {
			if ep.Valid() {
				filtered = append(filtered, ep)
			}
		}
		return filtered, fmt.Errorf("connection rejected")
	}

	// initialize channels
	ep.Port = reply.Header.PeerPort
	p.Endpoint = ep
	p.NodeID = uint32(reply.Header.NodeID)
	p.Hash = fmt.Sprintf("%s:%s %08x", ep.IP, ep.Port, p.NodeID)
	p.send = newParcelChannel(p.net.conf.ChannelCapacity)
	p.IsIncoming = incoming
	p.connected = time.Now()
	p.logger = p.logger.WithFields(log.Fields{
		"hash":    p.Hash,
		"address": p.Endpoint.IP,
		"Port":    p.Endpoint.Port,
		"Version": p.prot.Version(),
	})

	p.status <- peerStatus{peer: p, online: true}
	p.registered = true

	go p.sendLoop()
	go p.readLoop()
	go p.statLoop()

	return nil, nil
}

// Stop disconnects the peer from its active connection
func (p *Peer) Stop() {
	p.stopper.Do(func() {
		p.logger.Debug("Stopping peer")
		sc := p.send

		p.send = nil

		close(p.stop)

		if p.conn != nil {
			p.conn.Close()
		}

		if sc != nil {
			close(sc)
		}

		if p.registered {
			p.status <- peerStatus{peer: p, online: false}
		}
	})
}

func (p *Peer) String() string {
	return p.Hash
}

func (p *Peer) Send(parcel *Parcel) {
	_, dropped := p.send.Send(parcel)
	p.dropped += uint64(dropped)
}

func (p *Peer) statLoop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			p.metricsMtx.Lock()
			mr, ms, br, bs := p.metrics.Collect()
			p.bpsDown = float64(br) / 5
			p.bpsUp = float64(bs) / 5
			p.totalBytesReceived += br
			p.totalBytesSent += bs

			p.mpsDown = float64(mr) / 5
			p.mpsUp = float64(ms) / 5
			p.totalParcelsReceived += mr
			p.totalParcelsSent += ms

			p.metricsMtx.Unlock()
		case <-p.stop:
			return
		}
	}
}

func (p *Peer) readLoop() {
	if p.net.prom != nil {
		p.net.prom.ReceiveRoutines.Inc()
		defer p.net.prom.ReceiveRoutines.Dec()
	}
	defer p.conn.Close() // close connection on fatal error
	for {
		p.conn.SetReadDeadline(time.Now().Add(p.net.conf.ReadDeadline))
		msg, err := p.prot.Receive()
		if err != nil {
			p.logger.WithError(err).Debug("connection error (readLoop)")
			p.Stop()
			return
		}

		if err := msg.Valid(); err != nil {
			p.logger.WithError(err).Warnf("received invalid msg, disconnecting peer")
			p.Stop()
			if p.net.prom != nil {
				p.net.prom.Invalid.Inc()
			}
			return
		}

		// metrics
		p.metricsMtx.Lock()
		p.lastReceive = time.Now()
		p.metricsMtx.Unlock()

		// stats
		if p.net.prom != nil {
			p.net.prom.ParcelsReceived.Inc()
			p.net.prom.ParcelSize.Observe(float64(len(msg.Payload)) / 1024)
			if msg.IsApplicationMessage() {
				p.net.prom.AppReceived.Inc()
			}
		}

		msg.Address = p.Hash // always set sender = peer
		if !p.deliver(msg) {
			return
		}
	}
}

// deliver is a blocking delivery of this peer's messages to the peer manager.
func (p *Peer) deliver(parcel *Parcel) bool {
	select {
	case p.data <- peerParcel{peer: p, parcel: parcel}:
	case <-p.stop:
		return false
	}
	return true
}

// sendLoop listens to the Outgoing channel, pushing all data from there
// to the tcp connection
func (p *Peer) sendLoop() {
	if p.net.prom != nil {
		p.net.prom.SendRoutines.Inc()
		defer p.net.prom.SendRoutines.Dec()
	}

	defer p.conn.Close() // close connection on fatal error
	for {
		select {
		case <-p.stop:
			return
		case parcel := <-p.send:
			if parcel == nil {
				p.logger.Error("Received <nil> pointer from application")
				continue
			}

			p.conn.SetWriteDeadline(time.Now().Add(p.net.conf.WriteDeadline))
			err := p.prot.Send(parcel)
			if err != nil { // no error is recoverable
				p.logger.WithError(err).Debug("connection error (sendLoop)")
				p.Stop()
				return
			}

			// metrics
			p.metricsMtx.Lock()
			p.lastSend = time.Now()
			p.metricsMtx.Unlock()

			// stats
			if p.net.prom != nil {
				p.net.prom.ParcelsSent.Inc()
				p.net.prom.ParcelSize.Observe(float64(len(parcel.Payload)+32) / 1024) // TODO FIX
				if parcel.IsApplicationMessage() {
					p.net.prom.AppSent.Inc()
				}
			}
		}
	}
}

// GetMetrics returns live metrics for this connection
func (p *Peer) GetMetrics() PeerMetrics {
	p.metricsMtx.RLock()
	defer p.metricsMtx.RUnlock()
	pt := "regular"
	if p.net.controller.isSpecial(p.Endpoint) {
		pt = "special_config"
	}
	return PeerMetrics{
		Hash:             p.Hash,
		PeerAddress:      p.Endpoint.IP,
		MomentConnected:  p.connected,
		LastReceive:      p.lastReceive,
		LastSend:         p.lastSend,
		BytesReceived:    p.totalBytesReceived,
		BytesSent:        p.totalBytesSent,
		MessagesReceived: p.totalParcelsReceived,
		MessagesSent:     p.totalParcelsSent,
		Incoming:         p.IsIncoming,
		PeerType:         pt,
		MPSDown:          p.mpsDown,
		MPSUp:            p.mpsUp,
		BPSDown:          p.bpsDown,
		BPSUp:            p.bpsUp,
		ConnectionState:  fmt.Sprintf("v%s", p.prot.Version()),
		Capacity:         p.Capacity(),
		Dropped:          p.dropped,
	}
}

// Capacity is a wrapper for the send channel's Capacity
func (p *Peer) Capacity() float64 {
	return p.send.Capacity()
}
