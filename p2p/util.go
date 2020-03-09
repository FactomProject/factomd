package p2p

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
)

// IP2Location converts an ip address to a uint32
//
// If the address is a hostmask, it attempts to resolve the address first
func IP2Location(addr string) (uint32, error) {
	// Split the IPv4 octets
	ip := net.ParseIP(addr)
	if ip == nil {
		ipAddress, err := net.LookupHost(addr)
		if err != nil {
			return 0, err // We use location on 0 to say invalid
		}
		addr = ipAddress[0]
		ip = net.ParseIP(addr)
	}
	if len(ip) == 16 { // If we got back an IP6 (16 byte) address, use the last 4 byte
		ip = ip[12:]
	}

	return binary.BigEndian.Uint32(ip), nil
}

// IP2LocationQuick converts an ip address to a uint32 without a hostmask lookup
func IP2LocationQuick(addr string) uint32 {
	// Split the IPv4 octets
	ip := net.ParseIP(addr)
	if ip == nil {
		return 0
	}

	if len(ip) == 16 { // If we got back an IP6 (16 byte) address, use the last 4 byte
		ip = ip[12:]
	}

	return binary.BigEndian.Uint32(ip)
}

// StringToUint32 hashes the input to generate a deterministic number representation
func StringToUint32(input string) uint32 {
	hash := sha256.Sum256([]byte(input))
	return binary.BigEndian.Uint32(hash[:4])
}

// WebScanner is a wrapper that applies the closure f to the response body
func WebScanner(url string, f func(line string)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid http status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		f(scanner.Text())
	}

	return nil
}
