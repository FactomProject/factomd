package eventservices

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestDBStateMapping(t *testing.T) {
	msg := newDBStateMsg()
	msg.String()

}

func newDBStateMsg() *messages.DBStateMsg {
	msg := new(messages.DBStateMsg)
	msg.Timestamp = primitives.NewTimestampNow()

	set := testHelper.CreateTestBlockSet(nil)
	set = testHelper.CreateTestBlockSet(set)

	msg.DirectoryBlock = set.DBlock
	msg.AdminBlock = set.ABlock
	msg.FactoidBlock = set.FBlock
	msg.EntryCreditBlock = set.ECBlock
	msg.EBlocks = []interfaces.IEntryBlock{set.EBlock, set.AnchorEBlock}
	for _, e := range set.Entries {
		msg.Entries = append(msg.Entries, e)
	}

	return msg
}
