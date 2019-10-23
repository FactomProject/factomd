package util

import (
	"fmt"
)

// Calculate the entry credits needed for the entry
func EntryCost(b []byte) (uint8, error) {
	// caulculaate the length exluding the header size 35 for Milestone 1
	l := len(b) - 35

	if l > 10240 {
		return 10, fmt.Errorf("Entry cannot be larger than 10KB")
	}

	// n is the capacity of the entry payment in KB
	r := l % 1024
	n := uint8(l / 1024)

	if r > 0 {
		n += 1
	}

	if n < 1 {
		n = 1
	}

	return n, nil
}
