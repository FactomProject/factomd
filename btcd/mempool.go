// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package btcd

import (
	"errors"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"sync"
	"time"
)

// ftmMemPool is used as a source of factom transactions
// (CommitChain, RevealChain, CommitEntry, RevealEntry)
type ftmMemPool struct {
	sync.RWMutex
	pool        map[interfaces.IHash]messages.Message
	orphans     map[interfaces.IHash]messages.Message
	blockpool   map[string]messages.Message // to hold the blocks or entries downloaded from peers
	lastUpdated time.Time               // last time pool was updated
}

// Add a factom message to the orphan pool
func (mp *ftmMemPool) initFtmMemPool() error {

	mp.pool = make(map[interfaces.IHash]messages.Message)
	mp.orphans = make(map[interfaces.IHash]messages.Message)
	mp.blockpool = make(map[string]messages.Message)

	return nil
}

// Add a factom message to the  Mem pool
func (mp *ftmMemPool) addMsg(msg messages.Message, hash interfaces.IHash) error {

	if len(mp.pool) > constants.MAX_TX_POOL_SIZE {
		return errors.New("Transaction mem pool exceeds the limit.")
	}

	mp.pool[hash] = msg

	return nil
}

// Add a factom message to the orphan pool
func (mp *ftmMemPool) addOrphanMsg(msg messages.Message, hash interfaces.IHash) error {

	if len(mp.orphans) > constants.MAX_ORPHAN_SIZE {
		return errors.New("Ophan mem pool exceeds the limit.")
	}

	mp.orphans[hash] = msg

	return nil
}

// Add a factom block message to the  Mem pool
func (mp *ftmMemPool) addBlockMsg(msg messages.Message, hash string) error {

	if len(mp.blockpool) > constants.MAX_BLK_POOL_SIZE {
		return errors.New("Block mem pool exceeds the limit. Please restart.")
	}
	mp.Lock()
	mp.blockpool[hash] = msg
	mp.Unlock()

	return nil
}

// Add a factom block message to the  Mem pool
func (mp *ftmMemPool) getBlockMsg(hash string) messages.Message {
	return mp.blockpool[hash]
}

// Delete a factom block message from the  Mem pool
func (mp *ftmMemPool) deleteBlockMsg(hash string) error {

	if mp.blockpool[hash] != nil {
		mp.Lock()
		delete(mp.blockpool, hash)
		mp.Unlock()
	}

	return nil
}
