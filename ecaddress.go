// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

type IECAddress interface {
    IBlock
}

type eCAddress struct {
    address
}