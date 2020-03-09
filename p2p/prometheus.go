package p2p

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var once sync.Once

// Prometheus holds all of the prometheus recording instruments
type Prometheus struct {
	Connections      prometheus.Gauge
	Connecting       prometheus.Gauge
	Incoming         prometheus.Gauge
	Outgoing         prometheus.Gauge
	ToNetwork        prometheus.Gauge
	ToNetworkRatio   prometheus.Gauge
	FromNetwork      prometheus.Gauge
	FromNetworkRatio prometheus.Gauge
	CatRounds        prometheus.Counter

	SendRoutines    prometheus.Gauge
	ReceiveRoutines prometheus.Gauge

	ParcelsSent     prometheus.Counter
	ParcelsReceived prometheus.Counter
	Invalid         prometheus.Counter
	AppSent         prometheus.Counter
	AppReceived     prometheus.Counter
	ByteRateDown    prometheus.Gauge
	ByteRateUp      prometheus.Gauge
	MessageRateDown prometheus.Gauge
	MessageRateUp   prometheus.Gauge

	ParcelSize prometheus.Histogram
}

// Setup registers all of the instruments with prometheus once
func (p *Prometheus) Setup() {
	once.Do(func() {
		ng := func(name, help string) prometheus.Gauge {
			g := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: name,
				Help: help,
			})
			prometheus.MustRegister(g)
			return g
		}

		p.Connections = ng("factomd_p2p_peers_online", "Number of established connections")
		p.Connecting = ng("factomd_p2p_peers_connecting", "Number of connections currently dialing or awaiting handshake")
		p.Incoming = ng("factomd_p2p_peers_incoming", "Number of peers that have dialed to this node")
		p.Outgoing = ng("factomd_p2p_peers_outgoing", "Number of peers that this node has dialed to")
		p.ToNetwork = ng("factomd_p2p_to_network", "The number of parcels in the ToNetwork channel")
		p.ToNetworkRatio = ng("factomd_p2p_to_network_ratio", "The fill ratio of the ToNetwork channel")
		p.FromNetwork = ng("factomd_p2p_from_network", "The number of parcels in the FromNetwork channel")
		p.FromNetworkRatio = ng("factomd_p2p_from_network_ratio", "The fill ratio of the FromNetwork channel")
		p.CatRounds = ng("factomd_p2p_cat_rounds", "Number of CAT rounds")

		p.SendRoutines = ng("factomd_p2p_tech_sendroutines", "Number of active send routines")
		p.ReceiveRoutines = ng("factomd_p2p_tech_receiveroutines", "Number of active receive routines")

		p.ParcelsSent = ng("factomd_p2p_parcels_sent", "Total number of parcels sent out")
		p.ParcelsReceived = ng("factomd_p2p_parcels_received", "Total number of parcels received")
		p.Invalid = ng("factomd_p2p_parcels_invalid", "Total number of invalid parcels received")
		p.AppSent = ng("factomd_p2p_messages_sent", "Total number of application messages sent")
		p.AppReceived = ng("factomd_p2p_messages_received", "Total number of application messages received")
		p.ByteRateDown = ng("factomd_p2p_byte_rate_down", "Current rate of incoming traffic (in bytes/sec)")
		p.ByteRateUp = ng("factomd_p2p_byte_rate_up", "Current rate of outgoing traffic (in bytes/sec)")
		p.MessageRateDown = ng("factomd_p2p_msg_rate_down", "Current rate of incoming traffic (in messages/sec)")
		p.MessageRateUp = ng("factomd_p2p_msg_rate_up", "Current rate of outgoing traffic (in messages/sec)")

		p.ParcelSize = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "factomd_p2p_parcels_size",
			Help:    "Number of parcels encountered for specific sizes (in KiBi)",
			Buckets: prometheus.ExponentialBuckets(1, 2, 16),
		})
		prometheus.MustRegister(p.ParcelSize)
	})
}
