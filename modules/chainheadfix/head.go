package chainheadfix

import (
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
)

// head has the data necessary to figure out which eblock is the latest
type head struct {
	mtx    sync.Mutex
	height int64            // max height of eblock
	head   interfaces.IHash // keymr of max height eblock
	id     interfaces.IHash // chain id
}

// Update will save the new information if the height is greater
func (ch *head) Update(height int64, head interfaces.IHash) {
	ch.mtx.Lock()
	if height > ch.height {
		ch.head = head
		ch.height = height
	}
	ch.mtx.Unlock()
}

// create a new head for a chain
func newhead(id interfaces.IHash) *head {
	ch := new(head)
	ch.id = id
	ch.height = -1 // genesis dblock = 0
	return ch
}
