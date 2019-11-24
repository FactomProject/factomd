package p2p

import (
	"testing"
	"time"
)

func TestDialer(t *testing.T) {
	ep := Endpoint{IP: "127.255.255.254", Port: "65535"}

	interval, timeout := time.Millisecond*150, time.Millisecond*25

	d, err := NewDialer("127.0.0.1", interval, timeout)
	if err != nil {
		t.Error(err)
	}

	if !d.CanDial(ep) {
		t.Error("failed to connect first time")
	}
	d.Dial(ep) // takes <=25 ms
	if d.CanDial(ep) {
		t.Error("can dial during blocking interval")
	}

	time.Sleep(interval)

	if !d.CanDial(ep) {
		t.Error("failed to connect second time")
	}
	d.Dial(ep)
	if d.CanDial(ep) {
		t.Error("can dial during second blocking interval")
	}
}
