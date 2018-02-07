package primitives

import (
	"crypto/sha256"
)

type AuthSet struct {
	IdentityList []Identity
	StatusArray  []int
	IdentityMap  map[Identity]int
}

type Identity int

type Hash [sha256.Size]byte
