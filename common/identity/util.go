package identity

import (
	"encoding/binary"

	"github.com/PaulSnow/factom2d/common/constants"
)

const (
	TWELVE_HOURS_S uint64 = 12 * 60 * 60
)

// Makes sure the timestamp is within the designated window to be valid : 12 hours
// TimeEntered is in seconds
func CheckTimestamp(time []byte, timeEntered int64) bool {
	if len(time) < 8 {
		zero := []byte{00}
		add := make([]byte, 0)
		for i := len(time); i <= 8; i++ {
			add = append(add, zero...)
		}
		time = append(add, time...)
	}

	// In Seconds
	ts := binary.BigEndian.Uint64(time)
	var res uint64
	timeEnteredUint := uint64(timeEntered)
	if timeEnteredUint > ts {
		res = timeEnteredUint - ts
	} else {
		res = ts - timeEnteredUint
	}
	if res <= TWELVE_HOURS_S {
		return true
	} else {
		return false
	}
}

func BubbleSortUint32(arr []uint32) []uint32 {
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
	return arr
}

func statusIsFedOrAudit(status uint8) bool {
	if status == constants.IDENTITY_FEDERATED_SERVER ||
		status == constants.IDENTITY_AUDIT_SERVER ||
		status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
		status == constants.IDENTITY_PENDING_AUDIT_SERVER {
		return true
	}
	return false
}
