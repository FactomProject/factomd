package engine_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/engine"
)

func TestRegisterPrometheusTwice(t *testing.T) {
	// prometheus will panic if this fails
	RegisterPrometheus()
	RegisterPrometheus()
}
