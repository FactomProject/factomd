package p2p

import (
	"strings"
	"testing"
)

func Test_controller_setSpecial(t *testing.T) {
	add1 := []Endpoint{Endpoint{"aa", "1"}, Endpoint{"aa", "2"}, Endpoint{"ab", "3"}, Endpoint{"ab", "4"}}
	add2 := []Endpoint{Endpoint{"ba", "1"}, Endpoint{"ba", "2"}, Endpoint{"bb", "3"}, Endpoint{"bb", "4"}}
	add3 := []Endpoint{Endpoint{"c", "1"}}

	c := new(controller)
	c.logger = packageLogger
	c.setSpecial("a:1,localhost:80,127.0.0.1:8110") // parseSpecial has its own unit tests

	addf := func(el []Endpoint) {
		var eps []string
		for _, e := range el {
			eps = append(eps, e.String())
		}

		c.setSpecial(strings.Join(eps, ","))
	}

	cmp := func(el []Endpoint, prev []Endpoint) {
		for _, e := range el {
			if c.isSpecial(e) {
				t.Errorf("Endpoint %s was already special before being added", e)
			} else if c.isSpecialIP(e.IP) {
				t.Errorf("IP %s was already special before being added", e.IP)
			}
		}

		addf(el)

		for _, e := range el {
			if !c.isSpecial(e) {
				t.Errorf("Endpoint %s was not set as special", e)
			} else if !c.isSpecialIP(e.IP) {
				t.Errorf("IP %s was not set as special", e.IP)
			}
		}

		for _, e := range prev {
			if c.isSpecial(e) {
				t.Errorf("Old endpoint %s was not removed", e)
			} else if c.isSpecialIP(e.IP) {
				t.Errorf("Old IP %s was not removed", e.IP)
			}
		}
	}

	cmp(add1, nil)
	cmp(add2, add1)
	cmp(add3, add2)
	cmp(nil, add3) // will call setSpecial("")
}
