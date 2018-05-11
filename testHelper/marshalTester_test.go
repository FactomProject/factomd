package testHelper_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestMarshalTestingAssist(t *testing.T) {
	a := new(messages.Bounce)
	a.Timestamp = primitives.NewTimestampNow()
	b := new(messages.Bounce)
	TestMarshaling(a, b, 0, t)
	TestMarshaling(a, b, 10, t)
}

func TestABlockEntryFields(t *testing.T) {
	a := new(adminBlock.ForwardCompatibleEntry)
	a.Size = 0
	a.Data = []byte{}
	a.AdminIDType = 0x0D
	TestABlockEntryFunctions(a, a, t)
}
