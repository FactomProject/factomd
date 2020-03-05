package p2p

import (
	"fmt"
	math "math"
	"math/rand"
	"net"
	"regexp"
	"strconv"
)

func testRandomEndpoint() Endpoint {
	ip := fmt.Sprintf("%d.%d.%d.%d", 1+rand.Intn(255), rand.Intn(256), rand.Intn(256), 1+rand.Intn(255))
	return Endpoint{IP: ip, Port: testRandomPort()}
}

func testRandomPort() string {
	return fmt.Sprintf("%d", (1+rand.Uint32())%math.MaxUint16)
}

var hostnameRegex *regexp.Regexp

func init() {
	// built after https://en.wikipedia.org/wiki/Hostname#Restrictions_on_valid_hostnames
	//                                   (      optional labels           ) (    required label          )
	hostnameRegex = regexp.MustCompile(`^([0-9A-Za-z][0-9A-Za-z\-]{0,62}\.)*[0-9A-Za-z][0-9A-Za-z\-]{0,62}$`)
}

type Endpoint struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// NewEndpoint creates an Endpoint struct from a given ip and port, throws error if ip could not be resolved
func NewEndpoint(ip, port string) (Endpoint, error) {
	ep := Endpoint{ip, port}
	if !ep.Valid() {
		return Endpoint{}, fmt.Errorf("(%s:%s) is not a valid endpoint", ip, port)
	}
	return ep, nil
}

// ParseEndpoint takes input in the form of "ip:port" and returns its IP
func ParseEndpoint(s string) (Endpoint, error) {
	ip, port, err := net.SplitHostPort(s)
	if err != nil {
		return Endpoint{}, err
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return Endpoint{}, err
	}

	if p < 1 || p > 65535 {
		return Endpoint{}, fmt.Errorf("port out of range")
	}

	return NewEndpoint(ip, port)
}

func (ep Endpoint) String() string {
	return fmt.Sprintf("%s:%s", ep.IP, ep.Port)
}

// Verify checks if the data is usable. Does not check if the remote address works
func (ep Endpoint) Valid() bool {
	if ep.IP == "" || ep.Port == "" || len(ep.IP) > 253 { // 253 is max according to rfc 952
		return false
	}

	if p, err := strconv.Atoi(ep.Port); err != nil || p == 0 {
		return false
	}

	if parse := net.ParseIP(ep.IP); parse == nil {
		// check hostname
		return hostnameRegex.MatchString(ep.IP)
	}

	return true
}

// Equals returns true if both endpoints are the same
func (ep Endpoint) Equal(o Endpoint) bool {
	return ep.IP == o.IP && ep.Port == o.Port
}
