// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
)

func GetTimeMilli() uint64 {
	return uint64(time.Now().UnixNano()) / 1000000 // 10^-9 >> 10^-3
}

func GetTime() uint64 {
	return uint64(time.Now().Unix())
}

//A structure for handling timestamps for messages
type Timestamp uint64 //in milliseconds
var _ interfaces.BinaryMarshallable = (*Timestamp)(nil)
var _ interfaces.Timestamp = (*Timestamp)(nil)

func (a *Timestamp) IsSameAs(b interfaces.Timestamp) bool {
	return a.GetTimeMilliUInt64() == b.GetTimeMilliUInt64()
}

func NewTimestampNow() *Timestamp {
	t := new(Timestamp)
	t.SetTimeNow()
	return t
}

func NewTimestampFromSeconds(s uint32) *Timestamp {
	t := new(Timestamp)
	*t = Timestamp(int64(s) * 1000)
	return t
}

func NewTimestampFromMinutes(s uint32) *Timestamp {
	t := new(Timestamp)
	*t = Timestamp(int64(s) * 60000)
	return t
}

func NewTimestampFromMilliseconds(s uint64) *Timestamp {
	t := new(Timestamp)
	*t = Timestamp(s)
	return t
}

func (t *Timestamp) SetTimestamp(b interfaces.Timestamp) {
	if b == nil {
		t.SetTimeMilli(0)
	}
	t.SetTimeMilli(b.GetTimeMilli())
}

func (t *Timestamp) SetTimeNow() {
	*t = Timestamp(GetTimeMilli())
}

func (t *Timestamp) SetTimeMilli(miliseconds int64) {
	t.SetTime(uint64(miliseconds))
}

func (t *Timestamp) SetTime(miliseconds uint64) {
	*t = Timestamp(miliseconds)
}

func (t *Timestamp) SetTimeSeconds(seconds int64) {
	t.SetTime(uint64(seconds * 1000))
}

func (t *Timestamp) GetTime() time.Time {
	return time.Unix(int64(*t/1000), int64(((*t)%1000)*1000))
}

func (t *Timestamp) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	if data == nil || len(data) < 6 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	hd, data := binary.BigEndian.Uint32(data[:]), data[4:]
	ld, data := binary.BigEndian.Uint16(data[:]), data[2:]
	*t = Timestamp((uint64(hd) << 16) + uint64(ld))
	return data, nil
}

func (t *Timestamp) UnmarshalBinary(data []byte) error {
	_, err := t.UnmarshalBinaryData(data)
	return err
}

func (t *Timestamp) GetTimeSeconds() int64 {
	return int64(*t / 1000)
}

func (t *Timestamp) GetTimeMinutesUInt32() uint32 {
	return uint32(*t / 60000)
}

func (t *Timestamp) GetTimeMilli() int64 {
	return int64(*t)
}

func (t *Timestamp) GetTimeMilliUInt64() uint64 {
	return uint64(*t)
}

func (t *Timestamp) GetTimeSecondsUInt32() uint32 {
	return uint32(*t / 1000)
}

func (t *Timestamp) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Timestamp.MarshalBinary err:%v", *pe)
		}
	}(&err)
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

func (t *Timestamp) UTCString() string {
	return t.GetTime().UTC().Format("2006-01-02 15:04:05")
}

// Clone()
// Functions that return timestamps in structures should clone said timestamps so users
// don't change the timestamp in the structures.
func (t Timestamp) Clone() interfaces.Timestamp {
	return &t
}
