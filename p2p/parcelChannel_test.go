package p2p

import (
	"testing"
	"time"
)

func TestParcelChannel_Send(t *testing.T) {
	smolChannel := newParcelChannel(1)
	bigChannel := newParcelChannel(15)

	parcel := new(Parcel)
	parcel.Payload = []byte("test")

	finished := false
	go func() {
		time.Sleep(time.Millisecond * 10)
		if !finished {
			t.Error("Channels caused a deadlock while writing")
		}
	}()

	for i := 0; i < 10; i++ {
		smolChannel.Send(parcel)
		bigChannel.Send(parcel)
	}
	finished = true

	if len(smolChannel) != 1 {
		t.Errorf("Small channel has unexpected length: %d", len(smolChannel))
	}
	if len(bigChannel) != 10 {
		t.Errorf("Big channel has unexpected length: %d", len(bigChannel))
	}
}

func TestParcelChannel_Reader(t *testing.T) {
	ch := newParcelChannel(100)

	for i := 0; i < 100; i++ {
		parcel := new(Parcel)
		parcel.Payload = []byte{byte(i)}
		ch.Send(parcel)
	}

	reader := ch.Reader()

	received := 0

Outer:
	for {
		switch {
		case len(reader) > 0:
			p := <-reader
			if p.Payload[0] != byte(received) {
				t.Fatalf("Received parcel out of order, want %d got %d", received, int(p.Payload[0]))
			}
			received++
		default:
			if received != 100 {
				t.Fatalf("Did not receive all parcels, only got %d", received)
			}
			break Outer
		}
	}
}
