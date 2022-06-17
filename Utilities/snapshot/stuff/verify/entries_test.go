package verify

import (
	"encoding/hex"
	"testing"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/stretchr/testify/assert"

	"github.com/FactomProject/FactomCode/common"
)

func TestMarshal(t *testing.T) {
	ent := common.NewEntry()
	ent2 := entryBlock.NewEntry()

	data, _ := hex.DecodeString("0000511c298668bc5032a64b76f8ede6f119add1a64482c8602966152c0b936c7700340019416461204c65766572736f6e")

	err := ent.UnmarshalBinary(data)
	assert.NoError(t, err, "common")

	err = ent2.UnmarshalBinary(data)
	assert.NoError(t, err, "eblock")
}
