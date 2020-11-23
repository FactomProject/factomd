package controlPanel_test

import (
	"testing"

	. "github.com/FactomProject/factomd/controlPanel"
)

func TestHeightToJsonStruct(t *testing.T) {
	j := HeightToJsonStruct(uint32(32))
	if string(j) != `{"Height":32}` {
		t.Errorf("Height Json does not match")
	}
}
