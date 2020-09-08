package p2p

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"
)

// Protobuf is a fairly robust third party protocol with its own unit tests
// These tests serve to verify our own code, not whether or not protobuf can encode proto messages

type testRW struct {
	r io.Reader
	w io.Writer
}

func newTestRW(r io.Reader, w io.Writer) *testRW {
	trw := new(testRW)
	trw.r = r
	trw.w = w
	return trw
}

func (trw *testRW) Read(p []byte) (int, error)  { return trw.r.Read(p) }
func (trw *testRW) Write(p []byte) (int, error) { return trw.w.Write(p) }

func TestProtocolV11_readWriteCheck(t *testing.T) {
	A, B := net.Pipe()
	A.SetDeadline(time.Now().Add(time.Millisecond * 500))
	B.SetDeadline(time.Now().Add(time.Millisecond * 500))
	defer A.Close()
	defer B.Close()
	reader := newProtocolV11(A)
	sender := newProtocolV11(B)

	// write 100 random byte slices
	data := make([][]byte, 100)
	for i := 0; i < len(data); i++ {
		data[i] = make([]byte, rand.Intn(4094)+4) // between 4 byte and 4KiB
		rand.Read(data[i])
	}

	go func() {
		for _, d := range data {
			if err := sender.writeCheck(d); err != nil {
				t.Error(err)
			}
		}
	}()

	buf := make([]byte, 4096)
	for i := 0; i < len(data); i++ {
		ibuf := buf[:len(data[i])]
		if err := reader.readCheck(ibuf); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(data[i], ibuf) {
			t.Errorf("invalid byte sequence read. want = %x, got = %x", data[i], ibuf)
		}
	}

	// test case where it goes wrong (not sending enough data)

	dl := time.Now().Add(time.Millisecond * 50)
	A.SetDeadline(dl)
	B.SetDeadline(dl)

	go func() {
		buf := make([]byte, 32)
		rand.Read(buf)
		if err := sender.writeCheck(buf); err != nil {
			t.Error(err)
		}

	}()

	buf = make([]byte, 64)
	if err := reader.readCheck(buf); err == nil {
		t.Errorf("did not receive an error when it should have timed out")
	}

}

func TestProtocolV11_readMessage(t *testing.T) {
	msg := new(V11Msg)
	msg.Type = 1
	msg.Payload = []byte("foo")

	data, err := msg.Marshal()
	if err != nil {
		t.Fatal(data)
	}

	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))

	prot := new(ProtocolV11)

	// should produce no errors
	valid := append(size, data...)
	prot.rw = newTestRW(bytes.NewReader(valid), nil)

	validReply := new(V11Msg)
	if err := prot.readMessage(validReply); err != nil {
		t.Errorf("error valid: %v", err)
	}

	short := append(size, data[:len(data)-1]...)
	prot.rw = newTestRW(bytes.NewReader(short), nil)

	invalidReply := new(V11Msg)
	if err := prot.readMessage(invalidReply); err == nil {
		t.Errorf("short buf didn't give us an error: %v", err)
	}

	binary.BigEndian.PutUint32(size, V11MaxParcelSize+1)
	long := append(size, data[:len(data)]...)
	prot.rw = newTestRW(bytes.NewReader(long), nil)

	invalidReply = new(V11Msg)
	if err := prot.readMessage(invalidReply); err == nil {
		t.Errorf("long buf didn't give us an error: %v", err)
	}
}

func TestProtocolV11_readWriteMessage(t *testing.T) {
	messages := make([]*V11Msg, 128)
	for i := range messages {
		messages[i] = new(V11Msg)
		messages[i].Type = rand.Uint32()
		messages[i].Payload = make([]byte, 1+rand.Intn(4095))
		rand.Read(messages[i].Payload)
	}

	A, B := net.Pipe()
	dl := time.Now().Add(time.Millisecond * 500)
	A.SetDeadline(dl)
	B.SetDeadline(dl)

	sender := newProtocolV11(A)
	reader := newProtocolV11(B)

	go func() {
		for _, m := range messages {
			if err := sender.writeMessage(m); err != nil {
				t.Error(err)
			}
		}
	}()

	for _, m := range messages {
		reply := new(V11Msg)
		if err := reader.readMessage(reply); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(m, reply) {
			t.Errorf("received wrong message. sent = %+v, got = %+v", m, reply)
		}
	}
}
