package p2p

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	io_prometheus_client "github.com/prometheus/client_model/go"
)

// tests ability for the controller to stop
func Test_controller_run(t *testing.T) {
	net := testNetworkHarness(t)

	done := make(chan bool)
	go func() {
		done <- false
		net.controller.run()
		done <- true
	}()

	<-done // wait for loop to start
	net.Stop()

	select {
	case <-time.After(time.Millisecond * 500):
		t.Errorf("stop signal timed out")
	case <-done:
	}
}

// test metrics hook
func Test_controller_runMetrics_hook(t *testing.T) {
	net := testNetworkHarness(t)

	done := make(chan bool, 1)
	net.SetMetricsHook(func(_ map[string]PeerMetrics) { done <- true })

	net.controller.runMetrics()
	select {
	case <-time.After(time.Millisecond * 500):
		t.Errorf("hook never called")
	case <-done:
	}
}

// test prometheus metrics
func Test_controller_runMetrics_prom(t *testing.T) {
	net := testNetworkHarness(t)
	net.prom = new(Prometheus)
	reg := prometheus.NewRegistry()
	net.prom._setup(reg)

	p := testRandomPeer(net)
	p.mpsDown = 1.123
	p.mpsUp = 1.456
	p.bpsDown = 2.123
	p.bpsUp = 2.456
	net.controller.peers.Add(p)

	net.controller.runMetrics()

	m := new(io_prometheus_client.Metric)
	if err := net.prom.MessageRateDown.Write(m); err != nil {
		t.Error(err)
	} else if *m.Gauge.Value != p.mpsDown {
		t.Errorf("MessageRateDown gauge value didn't match peer. got = %f, want = %f", *m.Gauge.Value, p.mpsDown)
	}

	if err := net.prom.MessageRateUp.Write(m); err != nil {
		t.Error(err)
	} else if *m.Gauge.Value != p.mpsUp {
		t.Errorf("MessageRateUp gauge value didn't match peer. got = %f, want = %f", *m.Gauge.Value, p.mpsUp)
	}

	if err := net.prom.ByteRateDown.Write(m); err != nil {
		t.Error(err)
	} else if *m.Gauge.Value != p.bpsDown {
		t.Errorf("ByteRateDown gauge value didn't match peer. got = %f, want = %f", *m.Gauge.Value, p.bpsDown)
	}

	if err := net.prom.ByteRateUp.Write(m); err != nil {
		t.Error(err)
	} else if *m.Gauge.Value != p.bpsUp {
		t.Errorf("ByteRateUp gauge value didn't match peer. got = %f, want = %f", *m.Gauge.Value, p.bpsUp)
	}
}

func Test_controller_runPing(t *testing.T) {
	net := testNetworkHarness(t)

	peers := []*Peer{testRandomPeer(net), testRandomPeer(net), testRandomPeer(net)}
	peers[0].lastSend = time.Now().Add(-net.conf.PingInterval).Add(time.Millisecond * 100)  // newer than cutoff
	peers[1].lastSend = time.Now().Add(-net.conf.PingInterval).Add(-time.Millisecond * 100) // older than cutoff
	peers[2].lastSend = time.Time{}                                                         // zero time

	for _, p := range peers {
		net.controller.peers.Add(p)
	}

	net.controller.runPing()

	if len(peers[0].send) > 0 {
		p := <-peers[0].send
		t.Errorf("peer 0 was sent a parcel of type %s. expected nil", p.ptype)
	}

	for i := 1; i < 2; i++ {
		if len(peers[i].send) == 0 {
			t.Error("peer i did not receive a parcel")
		} else {
			p := <-peers[i].send
			if p.ptype != TypePing {
				t.Errorf("peer %d did not receive a ping. got = %s", i, p.ptype)
			}
		}
	}
}
