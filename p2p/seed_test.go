package p2p

import (
	"net/http"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

func testServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/seed.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("127.0.0.1:80\n192.168.0.1:8088\n10.12.13.14:8110"))
	})
	mux.HandleFunc("/seedBad.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("127.0.0.256:80\n192.168.0.1:8088\n10.12.13.14:8110"))
	})
	mux.HandleFunc("/git.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("52.17.183.121:8108\n52.17.153.126:8108\n52.19.117.149:8108\n52.18.72.212:8108\n52.19.44.249:8108\n52.214.189.110:8108\n34.249.228.82:8108\n34.248.202.6:8108\n52.19.181.120:8108\n34.248.6.133:8108"))
	})

	go http.ListenAndServe("127.0.0.1:8000", mux)
}

func Test_seed_retrieve(t *testing.T) {

	testServer()

	log.SetLevel(log.DebugLevel)
	s := newSeed("http://localhost:8000/seed.txt", 0)
	s2 := newSeed("http://localhost:8000/seedBad.txt", 0)
	s3 := newSeed("http://localhost:8000/git.txt", 0)

	tests := []struct {
		name string
		s    *seed
		want []Endpoint
	}{
		{"base", s, []Endpoint{Endpoint{"127.0.0.1", "80"}, Endpoint{"192.168.0.1", "8088"}, Endpoint{"10.12.13.14", "8110"}}},
		{"bad", s2, []Endpoint{Endpoint{"192.168.0.1", "8088"}, Endpoint{"10.12.13.14", "8110"}}},
		{"mainnet", s3, []Endpoint{
			Endpoint{"52.17.183.121", "8108"},
			Endpoint{"52.17.153.126", "8108"},
			Endpoint{"52.19.117.149", "8108"},
			Endpoint{"52.18.72.212", "8108"},
			Endpoint{"52.19.44.249", "8108"},
			Endpoint{"52.214.189.110", "8108"},
			Endpoint{"34.249.228.82", "8108"},
			Endpoint{"34.248.202.6", "8108"},
			Endpoint{"52.19.181.120", "8108"},
			Endpoint{"34.248.6.133", "8108"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.retrieve(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("seed.retrieve() = %v, want %v", got, tt.want)
			}
		})
	}
}
