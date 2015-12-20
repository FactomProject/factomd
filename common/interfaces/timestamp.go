// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import (
	"bytes"
	"encoding/binary"
	"time"
)

//A structure for handling timestamps for messages
type Timestamp uint64 //in miliseconds

func (t *Timestamp) SetTimeNow() {
	*t = Timestamp(time.Now().UnixNano() / 1000000)
}

func (t *Timestamp) SetTime(miliseconds uint64) {
	*t = Timestamp(miliseconds)
}

func (t *Timestamp) GetTime() time.Time {
	return time.Unix(int64(*t/1000), int64(((*t)%1000)*1000))
}

func (t *Timestamp) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	hd, data := binary.BigEndian.Uint32(data[:]), data[4:]
	ld, data := binary.BigEndian.Uint16(data[:]), data[2:]
	*t = Timestamp((uint64(hd) << 16) + uint64(ld))
	return data, nil
}

func (t *Timestamp) UnmarshalBinary(data []byte) error {
	_, err := t.UnmarshalBinaryData(data)
	return err
}

func (t *Timestamp) GetTimeSecond() int64 {
	return int64(*t / 1000)
}

func (t *Timestamp) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer
	hd := uint32(*t >> 16)
	ld := uint16(*t & 0xFFFF)
	binary.Write(&out, binary.BigEndian, uint32(hd))
	binary.Write(&out, binary.BigEndian, uint16(ld))
	return out.Bytes(), nil
}

func (t *Timestamp) String() string {
	return t.GetTime().Format("2006-01-02 15:04:05")
}
