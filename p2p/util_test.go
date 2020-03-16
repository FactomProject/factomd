package p2p

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestWebScanner(t *testing.T) {
	lines := []string{"foo", "bar", "moo"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "foo")
		fmt.Fprintln(w, "bar")
		fmt.Fprint(w, "moo")
	}))
	defer ts.Close()

	testf1c := 0
	testf1 := func(line string) {
		if testf1c > len(lines) {
			t.Error("got more lines than test cases")
			return
		}
		if line != lines[testf1c] {
			t.Errorf("mismatch on line %d. got = %s, want = %s", testf1c, line, lines[testf1c])
		}
		testf1c++ // note: webscanner isn't multithreaded
	}

	type args struct {
		url string
		f   func(line string)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"first test case", args{ts.URL, testf1}, false},
		{"bad url", args{ts.URL + "@", nil}, true},
		{"404", args{"https://httpstat.us/404", nil}, true}, // if it goes down or is unreachable, test case will still pass
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WebScanner(tt.args.url, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("WebScanner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseSpecial(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name    string
		args    args
		want    []Endpoint
		wantErr bool
	}{
		{"bunch of addresses", args{"127.0.0.1:80,127.0.0.2:8080,1.1.1.1:8110"}, []Endpoint{{"127.0.0.1", "80"}, {"127.0.0.2", "8080"}, {"1.1.1.1", "8110"}}, false},
		{"spaces 1", args{" a:1,a:2,a:3,a:4"}, []Endpoint{{"a", "1"}, {"a", "2"}, {"a", "3"}, {"a", "4"}}, false},
		{"spaces 2", args{"a:1 ,a:2,a:3,a:4"}, []Endpoint{{"a", "1"}, {"a", "2"}, {"a", "3"}, {"a", "4"}}, false},
		{"spaces 3", args{"a:1 , a:2,a:3,a:4"}, []Endpoint{{"a", "1"}, {"a", "2"}, {"a", "3"}, {"a", "4"}}, false},
		{"spaces 4", args{"a:1,a:2,a:3,a:4"}, []Endpoint{{"a", "1"}, {"a", "2"}, {"a", "3"}, {"a", "4"}}, false},
		{"spaces 5", args{"a:1         ,a:2,a:3,a:4"}, []Endpoint{{"a", "1"}, {"a", "2"}, {"a", "3"}, {"a", "4"}}, false},
		{"spaces 6", args{"a:1,a:2,a:3,a :4"}, nil, true},
		{"semicolon", args{"a:1;a:2"}, nil, true},
		{"single address", args{"127.0.0.1:80"}, []Endpoint{{"127.0.0.1", "80"}}, false},
		{"single bad address", args{":50"}, nil, true},
		{"blank", args{""}, nil, true},
		{"just address", args{"127.0.0.1"}, nil, true},
		{"hostname w/o port", args{"domain.com"}, nil, true},
		{"hostname w port", args{"domain.com:80"}, []Endpoint{{"domain.com", "80"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSpecial(tt.args.raw)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSpecial() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("controller.parseSpecial() = %v, want %v", got, tt.want)
			}
		})
	}
}
