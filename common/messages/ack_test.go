// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"testing"
)

func Test(t *testing.T) {
	ack := new(Ack)
	t.Log(ack.String())
}
