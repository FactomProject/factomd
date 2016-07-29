package state

import (
	"bytes"
	"encoding/binary"
	"errors"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

type AnchorSigningKey struct {
	BlockChain string
	KeyLevel   byte
	KeyType    byte
	SigningKey []byte //if bytes, it is hex
}

type Identity struct {
	IdentityChainID      interfaces.IHash
	IdentityRegistered   uint32
	IdentityCreated      uint32
	ManagementChainID    interfaces.IHash
	ManagementRegistered uint32
	ManagementCreated    uint32
	MatryoshkaHash       interfaces.IHash
	Key1                 interfaces.IHash
	Key2                 interfaces.IHash
	Key3                 interfaces.IHash
	Key4                 interfaces.IHash
	SigningKey           interfaces.IHash
	Status               int
	AnchorKeys           []AnchorSigningKey
}

var _ interfaces.Printable = (*Identity)(nil)

func (e *Identity) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Identity) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Identity) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *Identity) String() string {
	str, _ := e.JSONString()
	return str
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

// Checking the external ids if they match the needed lengths
func CheckExternalIDsLength(extIDs [][]byte, lengths []int) bool {
	if len(extIDs) != len(lengths) {
		return false
	}
	for i := range extIDs {
		if !CheckLength(lengths[i], extIDs[i]) {
			log.Printfln("DEBUGL: Hit 2")
			return false
		}
	}
	return true
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
func CheckTimestamp(time []byte) bool {
	if len(time) < 8 {
		zero := []byte{00}
		add := make([]byte, 0)
		for i := len(time); i <= 8; i++ {
			add = append(add, zero...)
		}
		time = append(add, time...)
	}
	//TODO: get time from State for replaying?
	now := primitives.GetTime()

	ts := binary.BigEndian.Uint64(time)
	var res uint64
	if now > ts {
		res = now - ts
	} else {
		res = ts - now
	}
	if res <= TWELVE_HOURS_S {
		return true
	} else {
		return false
	}
}

func statusIsFedOrAudit(status int) bool {
	if status == constants.IDENTITY_FEDERATED_SERVER ||
		status == constants.IDENTITY_AUDIT_SERVER ||
		status == constants.IDENTITY_PENDING_FEDERATED_SERVER ||
		status == constants.IDENTITY_PENDING_AUDIT_SERVER {
		return true
	}
	return false
}
