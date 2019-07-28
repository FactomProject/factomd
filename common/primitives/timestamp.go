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

// GetTimeMilli returns the current time in milliseconds
func GetTimeMilli() uint64 {
	return uint64(time.Now().UnixNano()) / 1000000 // 10^-9 >> 10^-3
}

// GetTime returns the current time in seconds
func GetTime() uint64 {
	return uint64(time.Now().Unix())
}

// Timestamp is a structure for handling timestamps for messages
type Timestamp uint64 //in miliseconds
var _ interfaces.BinaryMarshallable = (*Timestamp)(nil)
var _ interfaces.Timestamp = (*Timestamp)(nil)

// IsSameAs returns true iff the timestamps match to the millisecond
func (t *Timestamp) IsSameAs(b interfaces.Timestamp) bool {
	return t.GetTimeMilliUInt64() == b.GetTimeMilliUInt64()
}

// NewTimestampNow creates a new time stamp at the current time
func NewTimestampNow() *Timestamp {
	t := new(Timestamp)
	t.SetTimeNow()
	return t
}

// NewTimestampFromSeconds creates a new timestamp in milliseconds from the input time in seconds
func NewTimestampFromSeconds(s uint32) *Timestamp {
	t := new(Timestamp)
	*t = Timestamp(int64(s) * 1000)
	return t
}

// NewTimestampFromMinutes creates a new timestamp in milliseconds from the input time in minutes
func NewTimestampFromMinutes(s uint32) *Timestamp {
	t := new(Timestamp)
	*t = Timestamp(int64(s) * 60000)
	return t
}

// NewTimestampFromMilliseconds creates a new timestamp in milliseconds from the input time in milliseconds
func NewTimestampFromMilliseconds(s uint64) *Timestamp {
	t := new(Timestamp)
	*t = Timestamp(s)
	return t
}

// SetTimestamp sets the timestamp to the timestamp value of the input. If input is nil, timestamp set to 0.
func (t *Timestamp) SetTimestamp(b interfaces.Timestamp) {
	if b == nil {
		t.SetTimeMilli(0)
	}
	t.SetTimeMilli(b.GetTimeMilli())
}

// SetTimeNow sets the timestamp in milliseconds to the current time in milliseconds
func (t *Timestamp) SetTimeNow() {
	*t = Timestamp(GetTimeMilli())
}

// SetTimeMilli sets the timestamp in milliseconds to the input time in milliseconds
func (t *Timestamp) SetTimeMilli(milliseconds int64) {
	t.SetTime(uint64(milliseconds))
}

// SetTime sets the timestamp in milliseconds to the input milliseconds
func (t *Timestamp) SetTime(milliseconds uint64) {
	*t = Timestamp(milliseconds)
}

// SetTimeSeconds sets the timestamp in milliseconds to the input time in seconds
func (t *Timestamp) SetTimeSeconds(seconds int64) {
	t.SetTime(uint64(seconds * 1000))
}

// GetTime returns the timestamp value in seconds
func (t *Timestamp) GetTime() time.Time {
	return time.Unix(int64(*t/1000), int64(((*t)%1000)*1000))
}

// UnmarshalBinaryData unmarshals the input data into the timestamp
func (t *Timestamp) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	if data == nil || len(data) < 6 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	hd, data := binary.BigEndian.Uint32(data[:]), data[4:]
	ld, data := binary.BigEndian.Uint16(data[:]), data[2:]
	*t = Timestamp((uint64(hd) << 16) + uint64(ld))
	return data, nil
}

// UnmarshalBinary unmarshals the input data into the timestamp
func (t *Timestamp) UnmarshalBinary(data []byte) error {
	_, err := t.UnmarshalBinaryData(data)
	return err
}

// GetTimeSeconds returns the timestamp value in seconds
func (t *Timestamp) GetTimeSeconds() int64 {
	return int64(*t / 1000)
}

// GetTimeMinutesUInt32 returns the timestamp value in minutes
func (t *Timestamp) GetTimeMinutesUInt32() uint32 {
	return uint32(*t / 60000)
}

// GetTimeMilli returns the timestamp value in milliseconds
func (t *Timestamp) GetTimeMilli() int64 {
	return int64(*t)
}

// GetTimeMilliUInt64 returns the timestamp value in milliseconds
func (t *Timestamp) GetTimeMilliUInt64() uint64 {
	return uint64(*t)
}

// GetTimeSecondsUInt32 returns the timestamp value in seconds
func (t *Timestamp) GetTimeSecondsUInt32() uint32 {
	return uint32(*t / 1000)
}

// MarshalBinary marshals the timestamp into a []byte array
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

// String returns the timestamp value in format: "yyyy-mm-dd hr:min:sec"
func (t *Timestamp) String() string {
	return t.GetTime().Format("2006-01-02 15:04:05")
}

// UTCString returns the timestamp value in UTC in format: "yyyy-mm-dd hr:min:sec"
func (t *Timestamp) UTCString() string {
	return t.GetTime().UTC().Format("2006-01-02 15:04:05")
}
