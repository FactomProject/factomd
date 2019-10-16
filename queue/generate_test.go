package queue_test

import (
	"io/ioutil"
	"testing"

	"github.com/FactomProject/factomd/queue"
	"github.com/stretchr/testify/assert"
)

func TestGenerateMsgQueue(t *testing.T) {

	s := queue.SourceFile{
		File:   "msg.go",
		Name:   "MsgQueue",
		Type:   "interfaces.IMsg",
		Import: "github.com/FactomProject/factomd/common/interfaces",
	}

	err := ioutil.WriteFile(s.File, s.Generate().Bytes(), 0644)
	assert.Nil(t, err)
}
