package testHelper_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/adminBlock"
	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
	. "github.com/PaulSnow/factom2d/testHelper"
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
