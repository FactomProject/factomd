package p2p

import (
	"fmt"
	"testing"
)

func TestNetworkID_String(t *testing.T) {
	tests := []struct {
		name string
		n    NetworkID
		want string
	}{
		{"zero", 0, "CustomNet ID: 0"},
		{"mainnet", MainNet, "MainNet"},
		{"community test", NetworkID(StringToUint32("fct_community_test")), "CustomNet ID: 883e093b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.String(); got != tt.want {
				t.Errorf("NetworkID.String() = %v, want %v", got, tt.want)
			}
			ref := &tt.n
			if fmt.Sprintf("%s", ref) != fmt.Sprintf("%s", tt.n) {
				t.Errorf("network and *network aren't the same. network = %v, *network = %v", fmt.Sprintf("%s", tt.n), fmt.Sprintf("%s", ref))
			}
		})
	}
}
