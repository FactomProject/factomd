package controlPanel_test

import (
	"fmt"
	"testing"

	. "github.com/PaulSnow/factom2d/controlPanel"
)

var _ = fmt.Sprintf("")

func TestHeightToJsonStruct(t *testing.T) {
	j := HeightToJsonStruct(uint32(32))
	if string(j) != `{"Height":32}` {
		t.Errorf("Height Json does not match")
	}
}
