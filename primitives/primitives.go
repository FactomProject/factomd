package primitives

import (
	"crypto/sha256"
)

type Identity int

type Hash [sha256.Size]byte

