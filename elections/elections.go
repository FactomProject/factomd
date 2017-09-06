package elections

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/elections"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"time"
)

var _ = fmt.Print
var _ = time.Tick

type Elections struct {
	ServerID  interfaces.IHash
	Name      string
	sync      []bool
	Federated []interfaces.IServer
	Audit     []interfaces.IServer
	fpriority []interfaces.IHash
	apriority []interfaces.IHash
	DBHeight  int
	Minute    int
	Input     interfaces.IQueue
	Output    interfaces.IQueue
	round     []int
	electing  int

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

func Fault(e *Elections, dbheight int, minute int) {
	time.Sleep(5 * time.Second)

	timeout := new(elections.TimeoutInternal)
	timeout.Minute = minute
	timeout.DBHeight = dbheight
	e.Input.Enqueue(timeout)

}

func Run(s *state.State) {

	e := new(Elections)
	e.Name = s.FactomNodeName
	e.Input = s.Elections()
	e.Output = s.InMsgQueue()
	for {
		msg := e.Input.BlockingDequeue()
		fmt.Println(msg.String())
		msg.(interfaces.IElections).ElectionProcess(s,e)

	messages:
		switch msg.(type) {
		case *elections.AddAuditInternal:
		case *elections.AddLeaderInternal:
			as := msg.(*elections.AddLeaderInternal)

			e.Print()
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
			e.Print()
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
			e.Print()
		case *elections.EomSigInternal:
			as := msg.(*elections.EomSigInternal)
			if int(as.DBHeight) > e.DBHeight || int(as.Minute) > e.Minute {

				// Set our Identity Chain (Just in case it has changed.)
				e.ServerID = s.IdentityChainID

				// Start our timer to timeout this sync
				go Fault(e, int(as.DBHeight), int(as.Minute))

				e.DBHeight = int(as.DBHeight)
				e.Minute = int(as.Minute)
				e.sync = make([]bool, len(e.Federated))
			}
			idx := e.LeaderIndex(as.ServerID)
			if idx >= 0 {
				e.sync[idx] = true
			}
			for _, b := range e.sync {
				if !b {
					break messages
				}
			}
			e.round = e.round[:0] // Get rid of any previous round counting.
		case *elections.TimeoutInternal:

			as := msg.(*elections.TimeoutInternal)
			if e.DBHeight > as.DBHeight || e.Minute > as.Minute {
				break messages
			}

			cnt := 0
			e.electing = -1
			for i, b := range e.sync {
				if !b {
					cnt++
					if e.electing < 0 {
						e.electing = i
					}
				}
			}
			// Hey, if all is well, then continue.
			if cnt == 0 {
				break messages
			}

			// If we don't have all our sync messages, we will have to come back around and see if all is well.
			go Fault(e, int(as.DBHeight), int(as.Minute))

			for len(e.round) <= e.electing {
				e.round = append(e.round, 0)
			}

			// New timeout, new round of elections.
			e.round[e.electing]++

			fmt.Printf("eee %20s Server Index: %d Round: %d %10s on #%d leaders \n",
				"Timeout",
				e.electing,
				e.round[e.electing],
				e.Name,
				cnt)

			// Can we see a majority of the federated servers?
			if cnt > len(e.Federated)/2 {
				// Reset the timeout and give up if we can't see a majority.
				break messages
			}
			fmt.Printf("eee %10s %s\n", e.Name, "Fault!")

			// Get the priority order list of audit servers in the priority order
			e.apriority = Order(e.Audit, e.DBHeight, e.Minute, e.electing, e.round[e.electing])

			idx := e.LeaderIndex(e.ServerID)
			// We are a leader
			if idx >= 0 {
				LeaderTimeout(e)
			}

			idx = e.AuditIndex(e.ServerID)
			if idx >= 0 {
				fmt.Printf("eee %10s %s\n", e.Name, "I'm an Audit Server")
				auditIdx := MaxIdx(e.apriority)
				if idx == auditIdx {
					V := new(elections.VolunteerAudit)
					V.TS = primitives.NewTimestampNow()
					V.NName = e.Name
					V.ServerIdx = uint32(e.electing)
					V.ServerID = e.ServerID
					V.Weight = e.apriority[idx]
					V.DBHeight = uint32(e.DBHeight)
					V.Minute = byte(e.Minute)
					V.Round = e.round[e.electing]
					fmt.Printf("eee %10s %s %s\n", e.Name, "I'm an Audit Server and I Volunteer", V.String())
					V.SendOut(s, V)
				}
			}
		case *elections.VolunteerAudit:
			fmt.Printf("eee %s\n", msg.String())

		}
	}
}
