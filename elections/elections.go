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
	Electing  int // This is the federated Server index that we are looking to replace
	State     interfaces.IState
	feedback  []string
}

func (e *Elections) NewFeedback() {
	e.feedback = make([]string, len(e.Federated)+len(e.Audit))
}

func (e *Elections) FeedBackStr(v string, index int) string {

	// If I have no feedback, then get some.
	if e.feedback == nil || len(e.feedback) == 0 {
		e.NewFeedback()
	}

	// Add the status if it is in my known range.
	if index >= 0 && index < len(e.feedback) {
		e.feedback[index] = v
	}

	// Make a string of the status.
	r := ""
	for _, v := range e.feedback {
		r = r + fmt.Sprintf("%3s ", v)
	}
	return r
}

func (e *Elections) String() string {
	str := fmt.Sprintf("eee %10s %s  dbht %d\n", e.State.GetFactomNodeName(), e.Name, e.DBHeight)
	str += fmt.Sprintf("eee %10s  %s\n", e.State.GetFactomNodeName(), "Federated Servers")
	for _, s := range e.Federated {
		str += fmt.Sprintf("eee %10s     %x\n", e.State.GetFactomNodeName(), s.GetChainID().Bytes())
	}
	str += fmt.Sprintf("eee %10s  %s\n", e.State.GetFactomNodeName(), "Audit Servers")
	for _, s := range e.Audit {
		str += fmt.Sprintf("eee %10s     %x\n", e.State.GetFactomNodeName(), s.GetChainID().Bytes())
	}
	return str
}

func (e *Elections) Print() {
	fmt.Println(e.String())
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

// Runs the main loop for elections for this instance of factomd
func Run(s *state.State) {
	e := new(Elections)
	s.Elections = e
	e.State = s
	e.Name = s.FactomNodeName
	e.Input = s.ElectionsQueue()
	e.Output = s.InMsgQueue()
	for {
		msg := e.Input.BlockingDequeue().(interfaces.IElectionMsg)
		msg.ElectionProcess(s, e)
	}
}
