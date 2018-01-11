// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"errors"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (s *State) FixMissingKeys(id *Identity) error {
	// This identity will always have blank keys
	if id.IdentityChainID.IsSameAs(s.GetNetworkBootStrapIdentity()) {
		return nil
	}
	if !statusIsFedOrAudit(id.Status.Load()) {
		//return
	}
	// Rebuilds identity
	err := s.AddIdentityFromChainID(id.IdentityChainID)
	if err != nil {
		return err
	}
	return nil
}

// Sig is signed message, msg is raw message
func CheckSig(idKey interfaces.IHash, pub []byte, msg []byte, sig []byte) bool {
	var pubFix [32]byte
	var sigFix [64]byte

	copy(pubFix[:], pub[:32])
	copy(sigFix[:], sig[:64])

	pre := make([]byte, 0)
	pre = append(pre, []byte{0x01}...)
	pre = append(pre, pubFix[:]...)
	id := primitives.Shad(pre)

	if id.IsSameAs(idKey) {
		return ed.VerifyCanonical(&pubFix, msg, &sigFix)
	} else {
		return false
	}
}

func CheckLength(length int, item []byte) bool {
	if len(item) != length {
		return false
	} else {
		return true
	}
}

func AppendExtIDs(extIDs [][]byte, start int, end int) ([]byte, error) {
	if len(extIDs) < (end + 1) {
		return nil, errors.New("Error: Index out of bound exception in AppendExtIDs()")
	}
	appended := make([]byte, 0)
	for i := start; i <= end; i++ {
		appended = append(appended, extIDs[i][:]...)
	}
	return appended, nil
}

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

func statusIsFedOrAudit(status uint8) bool {
	if status == constants.IDENTITY_FEDERATED_SERVER ||
		status == constants.IDENTITY_AUDIT_SERVER ||
		status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
		status == constants.IDENTITY_PENDING_AUDIT_SERVER {
		return true
	}
	return false
}
