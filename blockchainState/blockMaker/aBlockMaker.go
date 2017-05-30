// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

func (bm *BlockMaker) BuildABlock() (interfaces.IAdminBlock, error) {
	return nil, nil
}

func (bm *BlockMaker) ProcessABEntry(e interfaces.IABEntry) error {
	return nil
}
