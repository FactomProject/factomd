// +build all

package engine_test

import (
	"testing"

	. "github.com/FactomProject/factomd/engine"
)

func TestRegisterPrometheusTwice(t *testing.T) {
	// prometheus will panic if this fails
	RegisterPrometheus()
	RegisterPrometheus()
}
