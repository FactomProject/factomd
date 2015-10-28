// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"time"
)

//A structure for handling timestamps for messages
type Timestamp struct {
	Time time.Time
}

func (t *Timestamp) SetTimeNow() {
	t.Time = time.Now()
}

func (t *Timestamp) SetTime(sec, nsec int64) {
	t.Time = time.Unix(sec, nsec)
}

func (t *Timestamp) SetTimeFromBytes(data []byte) (newData []byte, err error) {
	return data[15:], t.Time.UnmarshalBinary(data[:15])
}

func (t *Timestamp) GetTime() time.Time {
	return t.Time
}

func (t *Timestamp) GetTimeSecond() int64 {
	return t.Time.Unix()
}

func (t *Timestamp) GetTimeByte() ([]byte, error) {
	return t.Time.MarshalBinary()
}

func (t *Timestamp) GetTimeString() string {
	return t.Time.Format("2006-01-02 15:04:05")
}
