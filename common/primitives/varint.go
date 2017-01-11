// Decode a variable integer from the given data buffer.
// Returns the uint64 bit value and a data slice positioned
// after the variable integer

package primitives

import (
	"github.com/FactomProject/factomd/common/primitives/random"
)

func RandomVarInt() uint64 {
	return random.RandUInt64()
}

func VarIntLength(v uint64) uint64 {
	buf := new(Buffer)
	EncodeVarInt(buf, v)
	return uint64(buf.Len())
}

func DecodeVarInt(data []byte) (uint64, []byte) {
	return DecodeVarIntGo(data)
}

func EncodeVarInt(out *Buffer, v uint64) error {
	return EncodeVarIntGo(out, v)
}

// Decode a varaible integer from the given data buffer.
// We use the algorithm used by Go, only BigEndian.
func DecodeVarIntGo(data []byte) (uint64, []byte) {
	var v uint64
	var cnt int
	var b byte
	for cnt, b = range data {
		v = v << 7
		v += uint64(b) & 0x7F
		if b < 0x80 {
			break
		}
	}
	return v, data[cnt+1:]
}

// Encode an integer as a variable int into the given data buffer.
func EncodeVarIntGo(out *Buffer, v uint64) error {
	if v == 0 {
		out.WriteByte(0)
	}
	h := v
	start := false

	if 0x8000000000000000&h != 0 { // Deal with the high bit set; Zero
		out.WriteByte(0x81) // doesn't need this, only when set.
		start = true        // Going the whole 10 byte path!
	}

	for i := 0; i < 9; i++ {
		b := byte(h >> 56) // Get the top 7 bits
		if b != 0 || start {
			start = true
			if i != 8 {
				b = b | 0x80
			} else {
				b = b & 0x7F
			}
			out.WriteByte(b)
		}
		h = h << 7
	}

	return nil
}
