package p2p

import (
	"crypto/sha1"
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

	resend *PeerHashCache

	// current state, read only "constants" after the handshake
	IsIncoming bool
	Endpoint   Endpoint
	Hash       string // This is more of a connection ID than hash right now.

	stopper sync.Once
	stop    chan bool

	lastPeerRequest time.Time
	lastPeerSend    time.Time

	// communication channels
	send ParcelChannel // parcels from Send() are added here

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

func newPeer(net *Network, id uint32, ep Endpoint, conn net.Conn, protocol Protocol, metrics ReadWriteCollector, incoming bool) *Peer {
	p := new(Peer)
	p.net = net
	p.prot = protocol
	p.Endpoint = ep
	p.metrics = metrics
	p.conn = conn

	p.stop = make(chan bool, 1)
	p.Hash = fmt.Sprintf("%s:%s %08x", ep.IP, ep.Port, id)

	p.logger = peerLogger.WithFields(log.Fields{
		"hash":    p.Hash,
		"address": p.Endpoint.IP,
		"Port":    p.Endpoint.Port,
		"Version": p.prot.Version(),
		"node":    p.net.conf.NodeName,
	})

	// initialize channels
	p.send = newParcelChannel(p.net.conf.ChannelCapacity)
	p.IsIncoming = incoming
	p.connected = time.Now()

	if net.conf.PeerResendFilter {
		p.resend = NewPeerHashCache(net.conf.PeerResendBuckets, net.conf.PeerResendInterval)
	}

	go p.sendLoop()
	go p.readLoop()
	go p.statLoop()

	return p
}

// Stop disconnects the peer from its active connection
func (p *Peer) Stop() {
	p.stopper.Do(func() {
		p.logger.Debug("Stopping peer")
		if p.resend != nil {
			p.resend.Stop()
		}
		close(p.stop) // stops sendLoop and readLoop and statLoop
		p.conn.Close()
		// sendLoop closes p.send in defer
		select {
		case p.net.controller.peerStatus <- peerStatus{peer: p, online: false}:
		case <-p.net.stopper:
		}
	})
}

func (p *Peer) String() string {
	return p.Hash
}

func (p *Peer) Send(parcel *Parcel) {
	select {
	case <-p.stop:
		// don't send when stopped
	default:
		_, dropped := p.send.Send(parcel)
		p.metricsMtx.Lock()
		p.dropped += uint64(dropped)
		p.metricsMtx.Unlock()
	}
}

func (p *Peer) statLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.metricsMtx.Lock()
			mw, mr, bw, br := p.metrics.Collect()
			p.bpsDown = float64(br)
			p.bpsUp = float64(bw)
			p.totalBytesReceived += br
			p.totalBytesSent += bw

			p.mpsDown = float64(mr)
			p.mpsUp = float64(mw)
			p.totalParcelsReceived += mr
			p.totalParcelsSent += mw

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
	defer p.Stop() // close connection on fatal error
	for {
		p.conn.SetReadDeadline(time.Now().Add(p.net.conf.ReadDeadline))
		msg, err := p.prot.Receive()
		if err != nil {
			p.logger.WithError(err).Debug("connection error (readLoop)")
			return
		}

		if err := msg.Valid(); err != nil {
			p.logger.WithError(err).Warnf("received invalid msg, disconnecting peer")
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

		if p.resend != nil && msg.IsApplicationMessage() {
			p.resend.Add(sha1.Sum(msg.Payload))
		}

		msg.Address = p.Hash // always set sender = peer
		if !p.deliver(msg) { // blocking unless peer is already stopped
			return
		}
	}
}

// deliver is a blocking delivery of this peer's messages to the peer manager.
func (p *Peer) deliver(parcel *Parcel) bool {
	select {
	case <-p.stop:
		return false
	case p.net.controller.peerData <- peerParcel{peer: p, parcel: parcel}:
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

	defer close(p.send)
	defer p.Stop() // close connection on fatal error
	for {
		select {
		case <-p.net.stopper:
			return
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
				return // stops in defer
			}

			// metrics
			p.metricsMtx.Lock()
			p.lastSend = time.Now()
			p.metricsMtx.Unlock()

			// stats
			if p.net.prom != nil {
				p.net.prom.ParcelsSent.Inc()
				p.net.prom.ParcelSize.Observe(float64(len(parcel.Payload)+32) / 1024)
				if parcel.IsApplicationMessage() {
					p.net.prom.AppSent.Inc()
				}
			}
		}
	}
}

func (p *Peer) LastSendAge() time.Duration {
	p.metricsMtx.RLock()
	defer p.metricsMtx.RUnlock()
	return time.Since(p.lastSend)
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
		ConnectionState:  fmt.Sprintf("v%s", p.prot),
		SendFillRatio:    p.SendFillRatio(),
		Dropped:          p.dropped,
	}
}

// SendFillRatio is a wrapper for the send channel's FillRatio
func (p *Peer) SendFillRatio() float64 {
	return p.send.FillRatio()
}
