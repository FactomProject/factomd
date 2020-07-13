package p2p

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestConfiguration_Sanitize(t *testing.T) {
	tmp := DefaultP2PConfiguration()
	c := &tmp
	c.MaxIncoming = 50
	c.MaxPeers = 45
	c.Sanitize()
	if c.MaxIncoming > c.MaxPeers {
		t.Errorf("allowing more incoming peers than possible peers. incoming = %d, max = %d", c.MaxIncoming, c.MaxPeers)
	}
}

func TestConfiguration_Check(t *testing.T) {
	if err := DefaultP2PConfiguration().Check(); err != nil {
		t.Errorf("DefaultConfiguration().Check() reported error = %v", err)
	}

	tests := []struct {
		name string
		v    interface{}
	}{
		{"ListenPort", ""},
		{"ListenPort", "abc"},
		{"ListenPort", ":123"},
		{"ListenPort", "123abc"},
		{"PeerShareAmount", uint(0)},
		{"RoundTime", time.Duration(0)},
		{"TargetPeers", uint(0)},
		{"Fanout", uint(0)},
		{"HandshakeTimeout", time.Duration(0)},
		{"DialTimeout", time.Duration(0)},
		{"ReadDeadline", time.Duration(0)},
		{"WriteDeadline", time.Duration(0)},
		{"ProtocolVersion", uint16(0)},
		{"ProtocolVersion", uint16(8)},
		{"ProtocolVersion", uint16(12)},
		{"ProtocolVersionMinimum", uint16(12)},
		{"ChannelCapacity", uint(0)},
		{"Special", "abc"}, // parseSpecial has its own unit tests, only check that it's checked
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tt.name), func(t *testing.T) {
			con := DefaultP2PConfiguration()
			r := reflect.ValueOf(&con)
			field := reflect.Indirect(r).FieldByName(tt.name)
			field.Set(reflect.ValueOf(tt.v))

			if err := con.Check(); err == nil {
				t.Errorf("Configuration.Check() error = %v, wantErrField = %s", err, tt.name)
			}
		})
	}
}

func TestDefaultP2PConfiguration(t *testing.T) {
	tests := []struct {
		name  string
		wantC Configuration
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotC := DefaultP2PConfiguration(); !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("DefaultP2PConfiguration() = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}
