package p2p_test

import (
	"net"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/p2p"
)

func TestParcelString(t *testing.T) {
	p := NewParcel(0, []byte{0xFF})
	c := new(ConnectionParcel)
	c.Parcel = *p

	correct := `{"Parcel":{"Header":{"Network":0,"Version":9,"Type":6,"Length":1,"TargetPeer":"","Crc32":4278190080,"PartNo":0,"PartsTotal":0,"NodeID":0,"PeerAddress":"","PeerPort":"8108","AppHash":"NetworkMessage","AppType":"Network"},"Payload":"/w=="}}`
	data, err := c.JSONByte()
	if err != nil {
		t.Error(err)
	}

	if string(data) != correct {
		t.Error("JSON format has changed in ConnectionParcels")
	}

	str, err := c.JSONString()
	if err != nil {
		t.Error(err)
	}

	if str != correct {
		t.Error("JSON format has changed in ConnectionParcels")
	}

	str = c.String()
	if str != correct {
		t.Error("JSON format has changed in ConnectionParcels")
	}
}

func TestConnectionCommandString(t *testing.T) {
	c := new(ConnectionCommand)
	c.Command = 4
	c.Delta = 2

	correct := `{"Command":4,"Peer":{"QualityScore":0,"Address":"","Port":"","NodeID":0,"Hash":"","Location":0,"Network":0,"Type":0,"Connections":0,"LastContact":"0001-01-01T00:00:00Z","Source":null},"Delta":2,"Metrics":{"MomentConnected":"0001-01-01T00:00:00Z","BytesSent":0,"BytesReceived":0,"MessagesSent":0,"MessagesReceived":0,"PeerAddress":"","PeerQuality":0,"PeerType":"","ConnectionState":"","ConnectionNotes":""}}`

	data, err := c.JSONByte()
	if err != nil {
		t.Error(err)
	}

	if string(data) != correct {
		t.Error("JSON format has changed in ConnectionParcels")
	}

	str, err := c.JSONString()
	if err != nil {
		t.Error(err)
	}

	if str != correct {
		t.Error("JSON format has changed in ConnectionParcels")
	}

	str = c.String()
	if str != correct {
		t.Error("JSON format has changed in ConnectionParcels")
	}
}

// NodeID is global, so we will only have loopbacks and go offline
func TestConnectionLoopBack(t *testing.T) {
	peer1 := new(Peer).Init("1.1.1.1", "1111", 0, RegularPeer, 0)
	peer1.Source["Accept()"] = time.Now()

	con1, con2 := net.Pipe()

	c1 := new(Connection)
	c1.InitWithConn(con1, *peer1)

	peer2 := new(Peer).Init("2.2.2.2", "1111", 0, RegularPeer, 0)
	peer2.Source["Accept()"] = time.Now()

	c2 := new(Connection)
	c2.InitWithConn(con2, *peer2)

	c1.Start()
	c2.Start()

	time.Sleep(400 * time.Millisecond)
	if c1.IsOnline() {
		t.Error("Should not be online as we have same nodeID")
	}

	con1.Close()
	con2.Close()
}
