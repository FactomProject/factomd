// +build all 

package p2p_test

import (
	"testing"

	. "github.com/FactomProject/factomd/p2p"
)

func TestRegisterPrometheus(t *testing.T) {
	RegisterPrometheus()
	RegisterPrometheus()
}
