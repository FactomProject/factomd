package p2p

import (
	"reflect"
	"testing"
)

func Test_controller_parseSpecial(t *testing.T) {
	valid := []Endpoint{
		{"127.0.0.1", "80"},
		{"127.0.0.2", "8080"},
		{"1.1.1.1", "8110"},
	}

	c := new(controller)
	c.logger = controllerLogger
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		c    *controller
		args args
		want []Endpoint
	}{
		{"bunch of addresses", c, args{"127.0.0.1:80,127.0.0.1,127.0.0.2,127.0.0.2:8080,1.1.1.1:8110,1.1.1.1:8089;127.0.0.1:4000"}, valid},
		{"single address", c, args{"127.0.0.1:80"}, []Endpoint{{"127.0.0.1", "80"}}},
		{"single bad address", c, args{":50"}, nil},
		{"blank", c, args{""}, nil},
		{"just address", c, args{"127.0.0.1"}, nil},
		{"hostname w/o port", c, args{"domain.com"}, nil},
		{"hostname w port", c, args{"domain.com:80"}, []Endpoint{{"domain.com", "80"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.parseSpecial(tt.args.raw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("controller.parseSpecial() = %v, want %v", got, tt.want)
			}
		})
	}
}
