package elections

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
	"time"
)

var _ = fmt.Print
var _ = time.Tick

type Elections struct {
	ServerID  interfaces.IHash
	Name      string
	Sync      []bool
	Federated []interfaces.IServer
	Audit     []interfaces.IServer
	FPriority []interfaces.IHash
	APriority []interfaces.IHash
	DBHeight  int
	Minute    int
	Input     interfaces.IQueue
	Output    interfaces.IQueue
	Round     []int
	Electing  int

	LeaderElecting  int // This is the federated Server we are electing, if we are a leader
	LeaderVolunteer int // This is the volunteer that we expect
}

func (e *Elections) Print() {
	str := fmt.Sprintf("%s  dbht %d", e.Name, e.DBHeight)
	str += fmt.Sprintf("  %s\n", "Federated Servers")
	for _, s := range e.Federated {
		str += fmt.Sprintf("     %x\n", s.GetChainID().Bytes())
	}
	str += fmt.Sprintf("  %s\n", "Audit Servers")
	for _, s := range e.Audit {
		str += fmt.Sprintf("     %x\n", s.GetChainID().Bytes())
	}
	fmt.Println(str)
}

// Returns the index of the given server. -1 if it isn't a Federated Server
func (e *Elections) LeaderIndex(server interfaces.IHash) int {
	for i, b := range e.Federated {
		if server.IsSameAs(b.GetChainID()) {
			return i
		}
	}
	return -1
}

// Returns the index of the given server. -1 if it isn't a Audit Server
func (e *Elections) AuditIndex(server interfaces.IHash) int {
	for i, b := range e.Audit {
		if server.IsSameAs(b.GetChainID()) {
			return i
		}
	}
	return -1
}

func Run(s *state.State) {
	e := new(Elections)
	e.Name = s.FactomNodeName
	e.Input = s.ElectionsQueue()
	e.Output = s.InMsgQueue()
	for {
		msg := e.Input.BlockingDequeue().(interfaces.IElectionMsg)
		fmt.Println(msg.String())
		msg.(interfaces.IElectionMsg).ElectionProcess(s, e)
		fmt.Println("eee" + msg.String())
		msg.ElectionProcess(s, e)

	}
}
