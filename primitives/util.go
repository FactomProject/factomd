package primitives

import (
	"bytes"
	"encoding/binary"
)

func Uint32ToBytes(val uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, val)

	return b
}

func BytesToUint32(data []byte) (ret uint32, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.BigEndian, &ret)
	return
}
