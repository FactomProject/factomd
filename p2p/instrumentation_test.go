package p2p_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/p2p"
)

func TestRegisterPrometheus(t *testing.T) {
	RegisterPrometheus()
	RegisterPrometheus()
}
