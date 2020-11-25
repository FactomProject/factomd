package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
)

// V11MaxParcelSize limits the amount of ram allocated for a parcel to 128 MiB
const V11MaxParcelSize = 134217728

// V11Signature is the 4-byte sequence that indicates the remote connection wants to use V11
var V11Signature = []byte{0x70, 0x62, 0x75, 0x66} // ascii for "pbuf"

var _ Protocol = (*ProtocolV11)(nil)

type ProtocolV11 struct {
	rw io.ReadWriter
}

func newProtocolV11(rw io.ReadWriter) *ProtocolV11 {
	v11 := new(ProtocolV11)
	v11.rw = rw
	return v11
}

func (v11 *ProtocolV11) SendHandshake(hs *Handshake) error {
	if err := v11.writeCheck(V11Signature); err != nil {
		return err
	}

	v11hs := new(V11Handshake)
	v11hs.Type = uint32(hs.Type)
	v11hs.ListenPort = hs.ListenPort
	v11hs.Loopback = hs.Loopback
	v11hs.Network = uint32(hs.Network)
	v11hs.NodeID = hs.NodeID
	v11hs.Version = uint32(hs.Version)

	if len(hs.Alternatives) > 0 {
		v11hs.Alternatives = make([]*V11Endpoint, 0, len(hs.Alternatives))
		for _, alt := range hs.Alternatives {
			v11hs.Alternatives = append(v11hs.Alternatives, &V11Endpoint{Host: alt.IP, Port: alt.Port})
		}
	}

	return v11.writeMessage(v11hs)
}

func (v11 *ProtocolV11) ReadHandshake() (*Handshake, error) {
	sig := make([]byte, 4)
	if err := v11.readCheck(sig); err != nil {
		return nil, err
	}
	if !bytes.Equal(sig, V11Signature) {
		return nil, fmt.Errorf("invalid connection signature")
	}

	v11hs := new(V11Handshake)
	if err := v11.readMessage(v11hs); err != nil {
		return nil, err
	}

	hs := new(Handshake)
	hs.Type = ParcelType(v11hs.Type)
	hs.ListenPort = v11hs.ListenPort
	hs.Loopback = v11hs.Loopback
	hs.Network = NetworkID(v11hs.Network)
	hs.NodeID = v11hs.NodeID
	hs.Version = uint16(v11hs.Version)

	if len(v11hs.Alternatives) > 0 {
		hs.Alternatives = make([]Endpoint, 0, len(v11hs.Alternatives))
		for _, alt := range v11hs.Alternatives {
			hs.Alternatives = append(hs.Alternatives, Endpoint{IP: alt.Host, Port: alt.Port})
		}
	}

	return hs, nil
}

func (v11 *ProtocolV11) readCheck(data []byte) error {
	if _, err := io.ReadFull(v11.rw, data); err != nil {
		return err
	}
	return nil
}

func (v11 *ProtocolV11) readMessage(msg proto.Message) error {
	buf := make([]byte, 4)
	if err := v11.readCheck(buf); err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(buf)

	if size > V11MaxParcelSize {
		return fmt.Errorf("peer attempted to send a handshake of size %d (max %d)", size, V11MaxParcelSize)
	}

	data := make([]byte, size)
	if err := v11.readCheck(data); err != nil {
		return err
	}

	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}

	return nil
}

func (v11 *ProtocolV11) writeCheck(data []byte) error {
	if n, err := v11.rw.Write(data); err != nil {
		return err
	} else if n != len(data) {
		return fmt.Errorf("unable to write data (%d of %d)", n, len(data))
	}
	return nil
}

func (v11 *ProtocolV11) writeMessage(msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if len(data) > V11MaxParcelSize {
		return fmt.Errorf("trying to send a message that's too large %d bytes (max %d)", len(data), V11MaxParcelSize)
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(data)))
	if err := v11.writeCheck(buf); err != nil {
		return err
	}

	if err := v11.writeCheck(data); err != nil {
		return err
	}
	return nil
}

func (v11 *ProtocolV11) Send(p *Parcel) error {
	msg := new(V11Msg)
	msg.Type = uint32(p.ptype)
	msg.Payload = p.Payload

	return v11.writeMessage(msg)
}

func (v11 *ProtocolV11) Receive() (*Parcel, error) {
	msg := new(V11Msg)
	if err := v11.readMessage(msg); err != nil {
		return nil, err
	}
	// type validity is checked in parcel.Valid
	return newParcel(ParcelType(msg.Type), msg.Payload), nil
}

func (v11 *ProtocolV11) Version() uint16 {
	return 11
}

func (v11 *ProtocolV11) String() string {
	return "11"
}

func (v11 *ProtocolV11) MakePeerShare(ps []Endpoint) ([]byte, error) {
	v11share := new(V11Share)
	v11share.Share = make([]*V11Endpoint, 0, len(ps))
	for _, ep := range ps {
		v11ep := new(V11Endpoint)
		v11ep.Host = ep.IP
		v11ep.Port = ep.Port
		v11share.Share = append(v11share.Share, v11ep)
	}

	return v11share.Marshal()
}

func (v11 *ProtocolV11) ParsePeerShare(payload []byte) ([]Endpoint, error) {
	v11share := new(V11Share)
	if err := v11share.Unmarshal(payload); err != nil {
		return nil, err
	}

	eps := make([]Endpoint, 0, len(v11share.Share))
	for _, v11ep := range v11share.Share {
		eps = append(eps, Endpoint{IP: v11ep.Host, Port: v11ep.Port})
	}

	return eps, nil
}
