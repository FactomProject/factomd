// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

func (ebm *EBlockMaker) BuildABlock() (interfaces.IAdminBlock, error) {
	return nil, nil
}

func (ebm *EBlockMaker) ProcessABEntry(e interfaces.IABEntry) error {
	return nil
}
