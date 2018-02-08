package primitives

import (
	"crypto/sha256"
)

type ProcessListLocation struct {
	Vm     int
	Minute int
	Height int
}

type AuthSet struct {
	IdentityList []Identity
	StatusArray  []int
	IdentityMap  map[Identity]int
}

type Identity int

type Hash [sha256.Size]byte
