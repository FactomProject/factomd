package p2p

import (
	"testing"
)

func TestIP2Location(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		{"localhost", args{"localhost"}, 1, false},
		{"localhost ipv4", args{"127.0.0.1"}, 2130706433, false},
		{"min ip", args{"0.0.0.0"}, 0, false},
		{"max ip", args{"255.255.255.255"}, 4294967295, false},
		{"invalid hostname", args{"#"}, 0, true},
		{"invalid ip", args{"256.0.0.0"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IP2Location(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("IP2Location() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IP2Location() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToUint32(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{"empty", args{""}, 0xE3B0C442},
		{"testnet", args{"fct_community_test"}, 0x883e093b},
		{"default name", args{"FNode0"}, 0x38BAB145},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToUint32(tt.args.input); got != tt.want {
				t.Errorf("StringToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}
