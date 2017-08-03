package elections

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/elections"
	"github.com/FactomProject/factomd/state"
	"time"
)

var _ = fmt.Print
var _ = time.Tick

type Elections struct {
	Name      string
	Federated []interfaces.IServer
	Audit     []interfaces.IServer
	DBHeight  int
	Input     interfaces.IQueue
	Output    interfaces.IQueue
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

type AuthMsg interface {
	GetServerID() interfaces.IHash
}

func sort(serv []interfaces.IServer) {
	for i := 0; i < len(serv)-1; i++ {
		allgood := true
		for j := 0; j < len(serv)-1-i; j++ {
			if bytes.Compare(serv[j].GetChainID().Bytes(), serv[j+1].GetChainID().Bytes()) > 0 {
				s := serv[j]
				serv[j] = serv[j+1]
				serv[j+1] = s
				allgood = false
			}
		}
		if allgood {
			return
		}
	}
}

func Run(s *state.State) {

	e := new(Elections)
	e.Name = s.FactomNodeName
	e.Input = s.Elections()
	e.Output = s.InMsgQueue()
	for {
		msg := e.Input.BlockingDequeue()
		fmt.Println(msg.String())
		switch msg.(type) {
		case *elections.AddAuditInternal:
			as := msg.(*elections.AddAuditInternal)
			e.Audit = append(e.Audit, &state.Server{ChainID: as.GetServerID(), Online: true})
			sort(e.Audit)
		case *elections.AddLeaderInternal:
			as := msg.(*elections.AddLeaderInternal)
			e.Federated = append(e.Federated, &state.Server{ChainID: as.GetServerID(), Online: true})
			sort(e.Federated)
		case *elections.RemoveAuditInternal:
			as := msg.(*elections.RemoveAuditInternal)
			idx := 0
			for i, s := range e.Audit {
				idx = i
				if s.GetChainID().IsSameAs(as.GetServerID()) {
					break
				}
			}
			if idx < len(e.Audit) {
				e.Audit = append(e.Audit[:idx], e.Audit[idx+1:]...)
			}
		case *elections.RemoveLeaderInternal:
			as := msg.(*elections.RemoveLeaderInternal)
			idx := 0
			for i, s := range e.Federated {
				idx = i
				if s.GetChainID().IsSameAs(as.GetServerID()) {
					break
				}
			}
			if idx < len(e.Federated) {
				e.Federated = append(e.Federated[:idx], e.Federated[idx+1:]...)
			}
			//case *elections.EomSigInternal:
			//	as := msg.(*elections.EomSigInternal)
			//	if as.DBHeight > e.DBHeight {
			//		fmt.Printf("Starting %10s at DBHeight %d\n", e.Name, as.DBHeight)
			//		e.DBHeight = as.DBHeight
			//	}
		}

		e.Print()
	}
}
