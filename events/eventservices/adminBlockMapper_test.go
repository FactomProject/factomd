package eventservices_test

import (
	"github.com/FactomProject/factomd/events/eventservices"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapAdminBlock(t *testing.T) {
	adminBlockMsg := testHelper.CreateTestAdminBlock(nil)
	adminBlockEvent := eventservices.MapAdminBlock(adminBlockMsg)
	assert.NotNil(t, adminBlockEvent)
	assert.NotNil(t, adminBlockEvent.Header)
	assert.NotNil(t, adminBlockEvent.Header.PreviousBackRefHash)
	assert.Equal(t, uint32(2), adminBlockEvent.Header.MessageCount)
	assert.Equal(t, uint32(0), adminBlockEvent.Header.BlockHeight)
	assert.NotNil(t, adminBlockEvent.Entries)
	assert.Equal(t, 2, len(adminBlockEvent.Entries))
	assert.Equal(t, uint32(5), adminBlockEvent.Entries[0].AdminIdType)

	assert.NotNil(t, adminBlockEvent.Entries[1].AdminBlockEntry)
	assert.NotNil(t, adminBlockEvent.KeyMerkleRoot)
	block0 := adminBlockEvent.Entries[0].GetAddFederatedServer()
	assert.NotNil(t, block0)
	assert.NotNil(t, block0.IdentityChainID)
	assert.Equal(t, uint32(0), block0.BlockHeight)

	assert.Equal(t, uint32(8), adminBlockEvent.Entries[1].AdminIdType)
	assert.NotNil(t, adminBlockEvent.Entries[1].AdminBlockEntry)
	block11 := adminBlockEvent.Entries[1].GetAddFederatedServerSigningKey()
	assert.NotNil(t, block11)
	assert.NotNil(t, block11.IdentityChainID)
	assert.NotNil(t, block11.PublicKey)
	assert.Equal(t, uint32(1), block11.BlockHeight)
	assert.Equal(t, uint32(0), block11.KeyPriority)

}
