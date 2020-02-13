package wsapi

import (
	"fmt"
	"strings"
)

type HashType int

const (
	HashTypeUnspecified HashType = iota
	HashTypeFCTTx
	HashTypeECTx
	HashTypeEntry
	HashTypeEBlock
	HashTypeECBlock
	HashTypeFBlock
	HashTypeABlock
	HashTypeDBlock
)

func (ht HashType) String() string {
	switch ht {
	case HashTypeFCTTx:
		return "Factoid Transaction Hash"
	case HashTypeECTx:
		return "Entry Credit Transaction Hash"
	case HashTypeEntry:
		return "Entry Hash"
	case HashTypeEBlock:
		return "Entry Block KeyMR"
	case HashTypeECBlock:
		return "Entry Credit Block KeyMR"
	case HashTypeFBlock:
		return "Factoid Block KeyMR"
	case HashTypeABlock:
		return "Admin Block KeyMR"
	case HashTypeDBlock:
		return "Directory Block KeyMR"
	default:
		return "unspecified"
	}
}

func (ht *HashType) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "":
		*ht = HashTypeUnspecified
	case "fcttx":
		*ht = HashTypeFCTTx
	case "ectx":
		*ht = HashTypeECTx
	case "entry":
		*ht = HashTypeEntry
	case "eblock":
		*ht = HashTypeEBlock
	case "ecblock":
		*ht = HashTypeECBlock
	case "fblock":
		*ht = HashTypeFBlock
	case "ablock":
		*ht = HashTypeABlock
	case "dblock":
		*ht = HashTypeDBlock
	default:
		return fmt.Errorf("invalid hash type")
	}
	return nil
}
