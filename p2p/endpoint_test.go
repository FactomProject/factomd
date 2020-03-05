package p2p

import (
	"reflect"
	"testing"
)

func TestNewEndpoint(t *testing.T) {
	type args struct {
		addr string
		port string
	}
	tests := []struct {
		name    string
		args    args
		want    Endpoint
		wantErr bool
	}{
		{"ok localhost", args{"127.0.0.1", "8088"}, Endpoint{"127.0.0.1", "8088"}, false},
		{"ok other ip", args{"1.2.3.4", "8088"}, Endpoint{"1.2.3.4", "8088"}, false},
		{"empty", args{"", ""}, Endpoint{}, true},
		{"no port", args{"127.0.0.1", ""}, Endpoint{}, true},
		{"no ip", args{"", "8088"}, Endpoint{}, true},
		//{"invalid ip", args{"127.0.0.256", "8088"}, Endpoint{}, true}, // technically a valid hostname
		{"hostname", args{"localhost", "8088"}, Endpoint{"localhost", "8088"}, false}, // likely uses ::1 ipv6 address
		{"punycode", args{"xn--qei9019maa.xn--z38hpa", "8088"}, Endpoint{"xn--qei9019maa.xn--z38hpa", "8088"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEndpoint(tt.args.addr, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEndpoint(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Endpoint
		wantErr bool
	}{
		// valid formats use NewIP() which is tested above, so these test cases don't need to cover them again
		// only checking ones that would fail the parsing
		{"ok localhost", args{"127.0.0.1:80"}, Endpoint{"127.0.0.1", "80"}, false},
		{"port out of range", args{"127.0.0.1:70000"}, Endpoint{}, true},
		{"no port", args{"127.0.0.1"}, Endpoint{}, true},
		{"empty", args{""}, Endpoint{}, true},
		{"no ip", args{":80"}, Endpoint{}, true},
		// {"bad ip", args{"127.0:80"}, Endpoint{}, true}, // this is theoretically a valid hostname
		{"wrong format 1", args{"127.0.0.1,80"}, Endpoint{}, true},
		{"wrong format 2", args{"127.0.0.1:80 test"}, Endpoint{}, true},
		{"wrong format 3", args{"ip:127.0.0.1 port:80"}, Endpoint{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEndpoint(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndpoint_String(t *testing.T) {
	tests := []struct {
		name string
		ps   Endpoint
		want string
	}{
		{"normal", Endpoint{IP: "127.0.0.1", Port: "8088"}, "127.0.0.1:8088"},
		{"no addr", Endpoint{IP: "", Port: "8088"}, ":8088"},
		{"no port", Endpoint{IP: "127.0.0.1", Port: ""}, "127.0.0.1:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.String(); got != tt.want {
				t.Errorf("Endpoint.ConnectAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndpoint_Valid(t *testing.T) {
	tests := []struct {
		name string
		ps   Endpoint
		want bool
	}{
		{"valid address", Endpoint{IP: "127.0.0.1", Port: "80"}, true},
		{"no addr", Endpoint{IP: "", Port: "8088"}, false},
		{"no port", Endpoint{IP: "127.0.0.1", Port: ""}, false},
		{"zero port", Endpoint{IP: "127.0.0.1", Port: "0"}, false},
		{"nonnumeric port", Endpoint{IP: "127.0.0.1", Port: "eighty"}, false},
		{"nonnumeric port", Endpoint{IP: "127.0.0.1", Port: "80th"}, false},
		{"localhost", Endpoint{IP: "localhost", Port: "8088"}, true},
		{"domain name", Endpoint{IP: "factom.fct", Port: "8088"}, true},
		{"invalid characters", Endpoint{IP: "localho$t", Port: "8088"}, false},
		{"invalid start", Endpoint{IP: "-test.com", Port: "8088"}, false},
		{"invalid start middle", Endpoint{IP: "test.-com", Port: "8088"}, false},
		{"longest hostname", Endpoint{IP: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcde.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.com", Port: "8088"}, true},
		{"real domain", Endpoint{IP: "www.google.com", Port: "8088"}, true},
		{"invalid url", Endpoint{IP: "https://www.google.com", Port: "8088"}, false},
		{"punycode", Endpoint{IP: "xn--qei9019maa.xn--z38hpa", Port: "8088"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.Valid(); got != tt.want {
				t.Errorf("Endpoint.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndpoint_Equal(t *testing.T) {
	ep := func(ip, port string) Endpoint {
		return Endpoint{IP: ip, Port: port}
	}

	type args struct {
		o Endpoint
	}
	tests := []struct {
		name string
		ep   Endpoint
		args args
		want bool
	}{
		{"both empty", Endpoint{}, args{Endpoint{}}, true},
		{"both empty strings", ep("", ""), args{ep("", "")}, true},
		{"localhost, no port", ep("localhost", ""), args{ep("localhost", "")}, true},
		{"no ip, same port", ep("", "80"), args{ep("", "80")}, true},
		{"both set", ep("127.0.0.1", "8108"), args{ep("127.0.0.1", "8108")}, true},
		{"port wrong", ep("127.0.0.1", "51"), args{ep("127.0.0.1", "50")}, false},
		{"ip wrong", ep("127.0.0.1", "80"), args{ep("127.0.0.2", "80")}, false},
		{"both wrong", ep("127.0.0.1", "80"), args{ep("127.0.0.2", "81")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ep.Equal(tt.args.o); got != tt.want {
				t.Errorf("Endpoint.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
