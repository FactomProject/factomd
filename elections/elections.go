package elections

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

type Elections struct {
	Federated []interfaces.IServer
	Audit     []interfaces.IServer
	DBHeight  int
	Input     interfaces.IQueue
	Output    interfaces.IQueue
}

func Run(state *state.State) {
	e := new(Elections)
	e.Input = state.Elections()
	e.Output = state.InMsgQueue()
	for {
		msg := e.Input.Dequeue()
		fmt.Println("eeee", msg.String())
	}
}
