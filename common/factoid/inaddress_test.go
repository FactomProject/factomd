package factoid_test

import (
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestInAddress(t *testing.T) {
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := h.(interfaces.IAddress)
	inAdd := NewInAddress(add, 12345678)
	str := inAdd.String()

	t.Logf("InAdd str - %v", str)

	if strings.Contains(str, "input") == false {
		t.Error("'input' not found")
	}
	if strings.Contains(str, "0.12345678") == false {
		t.Error("'0.12345678' not found")
	}
	if strings.Contains(str, "FA3mHjgsVvQJjVbvJpy67deDKzEsqc8FsLU122i8Tj76rmakpqRL") == false {
		t.Error("'FA3mHjgsVvQJjVbvJpy67deDKzEsqc8FsLU122i8Tj76rmakpqRL' not found")
	}
	if strings.Contains(str, "0000000000bc614e") == false {
		t.Error("'0000000000bc614e' not found")
	}
	if strings.Contains(str, "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973") == false {
		t.Error("'ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973' not found")
	}
}
