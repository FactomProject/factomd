package code

import "encoding/binary"

type Header struct {
	Tag  byte
	Size uint64
}

const (
	TagDBlock = iota // Blockchain structs
	TagABlock
	TagFBlock
	TagECBlock
	TagEBlock
	TagEntry
	TagTX

	TagFCT // Balances
	TagEC
)

func (h *Header) MarshalBinary() []byte {
	var data [9]byte
	data[0] = h.Tag
	binary.BigEndian.PutUint64(data[1:], h.Size)
	return data[:]
}

func (h *Header) UnmarshalBinary(data []byte) []byte {
	h.Tag = data[0]
	h.Size = binary.BigEndian.Uint64(data[1:])
	return data[9:]
}
