package identity_test

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/common/identity"
)

// TestCheckTimestamp checks that the CheckTimestamp function works properly. CheckTimestamp returns valid if the input is <= 12 hours
func TestCheckTimestamp(t *testing.T) {
	var out bytes.Buffer
	now := time.Now()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix()))
	hex := out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}

	// Check 1 minute less than 12 hours before
	var delta uint64 = (11*60 + 59) * 60
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}

	// Check 1 minute less than 12 hours after
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())+delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == false {
		t.Error("Timestamp check failed")
	}

	// Check 1 minute more than 12 hours before
	delta = (11*60 + 61) * 60
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}

	// Check 1 minute more than 12 hours after
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())+delta)
	hex = out.Bytes()

	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}

	// Check 10 seconds more than 12 hours before
	delta = TWELVE_HOURS_S + 10
	out.Reset()
	binary.Write(&out, binary.BigEndian, uint64(now.Unix())-delta)
	hex = out.Bytes()
	if CheckTimestamp(hex, now.Unix()) == true {
		t.Error("Timestamp check failed")
	}
}
