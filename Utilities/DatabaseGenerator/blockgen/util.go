package blockgen

import (
	"bytes"
	"encoding/binary"
	"time"
)

// milliTime returns a 6 byte slice representing the unix time in milliseconds
func milliTime(unix int64) (r []byte) {
	buf := new(bytes.Buffer)
	t := time.Unix(unix, 0).UnixNano()
	m := t / 1e6
	binary.Write(buf, binary.BigEndian, m)
	return buf.Bytes()[2:]
}
