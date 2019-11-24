package p2p

import "time"

// functions related to the timed run loop

// run is responsible for everything that involves proactive management
// not based on reactions. runs once a second
func (c *controller) run() {
	c.logger.Debug("Start run()")
	defer c.logger.Error("Stop run()")

	for {
		c.runPersist()
		c.runCatRound()
		c.runMetrics()
		c.runPing()

		select {
		case <-time.After(time.Second):
		}
	}
}

func (c *controller) runPersist() {
	if c.net.conf.PersistFile == "" {
		return
	}

	if time.Since(c.lastPersist) > c.net.conf.PersistInterval {
		c.lastPersist = time.Now()

		data, err := c.persistData()
		if err != nil {
			c.logger.WithError(err).Warn("unable to create peer persist data")
		} else {
			err = c.writePersistFile(data)
			if err != nil {
				c.logger.WithError(err).Warn("unable to persist peer data")
			}
		}
	}
}

func (c *controller) runPing() {
	for _, p := range c.peers.Slice() {
		if time.Since(p.lastSend) > c.net.conf.PingInterval {
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
	if c.net.metricsHook != nil {
		go c.net.metricsHook(c.makeMetrics())
	}
}
