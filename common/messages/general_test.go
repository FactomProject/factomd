// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/messages"
)

func TestUnmarshalNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	_, _, err := UnmarshalMessageData(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	_, _, err = UnmarshalMessageData([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}
