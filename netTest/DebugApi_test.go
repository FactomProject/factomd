package nettest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDebugAPI(t *testing.T) {
	n := SetupNode(SINGLE_NODE, 0, t)

	t.Run("NetworkInfo", func(t *testing.T) {
		r := n.fnodes[0].NetworkInfo()
		assert.Equal(t, "Leader", r.Role)
	})

	t.Run("RunCmd", func(t *testing.T) {
		err := n.fnodes[0].RunCmd("R0")
		assert.Nil(t, err)
	})

	t.Run("WriteConfig", func(t *testing.T) {
		err := n.fnodes[0].WriteConfig(9, "")
		assert.Nil(t, err)
	})
}
