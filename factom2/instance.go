package factom2

import (
	"fmt"
	"github.com/PaulSnow/factom2d/common/interfaces"
)

type Instance struct {
	InMsg         chan interfaces.IMsg
	NetOutMsg     chan interfaces.IMsg
	NetOutInvalid chan interfaces.IMsg
	APIQueue      chan interfaces.IMsg
}

func (ins *Instance) Run() {
	for {
		msg := <-ins.InMsg
		fmt.Println(msg.String())
	}
}
