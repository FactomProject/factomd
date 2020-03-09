package p2p

import "time"

// functions related to the timed run loop

// run is responsible for everything that involves proactive management
// not based on reactions. runs once a second
func (c *controller) run() {
	c.logger.Debug("Start run()")
	defer c.logger.Error("Stop run()")

	for {
		c.runCatRound()
		c.runMetrics()
		c.runPing()

		select {
		case <-c.net.stopper:
			return
		case <-time.After(time.Second):
		}
	}
}

// the factom network ping behavior is to send a ping message after
// a specific duration has passed
func (c *controller) runPing() {
	for _, p := range c.peers.Slice() {
		if p.LastSendAge() > c.net.conf.PingInterval {
			ping := newParcel(TypePing, []byte("Ping"))
			p.Send(ping)
		}
	}
}

func (c *controller) makeMetrics() map[string]PeerMetrics {
	metrics := make(map[string]PeerMetrics)
	for _, p := range c.peers.Slice() {
		metrics[p.Hash] = p.GetMetrics()
	}
	return metrics
}

func (c *controller) runMetrics() {
	metrics := c.makeMetrics()
	if c.net.metricsHook != nil {
		go c.net.metricsHook(metrics)
	}
	if c.net.prom != nil {
		var MPSDown, MPSUp, BytesDown, BytesUp float64
		for _, m := range metrics {
			MPSDown += float64(m.MPSDown)
			MPSUp += float64(m.MPSUp)
			BytesDown += float64(m.BytesReceived)
			BytesUp += float64(m.BytesSent)
		}

		c.net.prom.ByteRateDown.Set(BytesDown)
		c.net.prom.ByteRateUp.Set(BytesUp)
		c.net.prom.MessageRateUp.Set(MPSUp)
		c.net.prom.MessageRateDown.Set(MPSDown)

		c.net.prom.ToNetworkRatio.Set(float64(len(c.net.toNetwork)) / float64(cap(c.net.toNetwork)))
		c.net.prom.FromNetworkRatio.Set(float64(len(c.net.toNetwork)) / float64(cap(c.net.fromNetwork)))
	}
}
