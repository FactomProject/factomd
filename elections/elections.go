package elections

import (
	"bytes"
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
}

// Creates an order for all servers by using a certain hash function.  The list of unordered hashes (in the same order
// as the slice of servers) is returned.
func Order(servers []interfaces.IServer, dbheight int, minute int, serverIdx int, round int) (priority []interfaces.IHash) {
	for _, s := range servers {
		var data []byte
		data = append(data, byte(round>>24), byte(round>>16), byte(round>>8), byte(round))
		data = append(data, byte(dbheight>>24), byte(dbheight>>16), byte(dbheight>>8), byte(dbheight))
		data = append(data, byte(minute))
		data = append(data, byte(serverIdx>>8), byte(serverIdx))
		data = append(data, s.GetChainID().Bytes()...)
		hash := primitives.Sha(data)
		priority = append(priority, hash)
	}
	return
}

// Returns the index of the maximum priority entry
func MaxIdx(priority []interfaces.IHash) (idx int) {
	for i, v := range priority {
		if bytes.Compare(v.Bytes(), priority[idx].Bytes()) > 0 {
			idx = i
		}
	}
	return
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

func Fault(e *Elections, dbheight int, minute int) {
	time.Sleep(5 * time.Second)

	timeout := new(elections.TimeoutInternal)
	timeout.Minute = minute
	timeout.DBHeight = dbheight
	fmt.Println("\neee Timeout triggered ", e.Name)
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
	messages:
		switch msg.(type) {
		case *elections.AddAuditInternal:
			as := msg.(*elections.AddAuditInternal)
			e.Audit = append(e.Audit, &state.Server{ChainID: as.GetServerID(), Online: true})
			sort(e.Audit)
			e.Print()
		case *elections.AddLeaderInternal:
			as := msg.(*elections.AddLeaderInternal)
			e.Federated = append(e.Federated, &state.Server{ChainID: as.GetServerID(), Online: true})
			sort(e.Federated)
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

				fmt.Printf("eee %20s %10s at DBHeight %d Minute %d\n", "Sync Starting", e.Name, as.DBHeight, as.Minute)
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
			fmt.Printf("eee %20s %10s across %d leaders \n", "Sync Complete", e.Name, len(e.sync))
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

			for len(e.round) < e.electing {
				e.round = append(e.round, 0)
			}

			// New timeout, new round of elections.
			e.round[e.electing]++

			fmt.Printf("eee %20s %10s on #%d leaders \n", "Timeout", e.Name, cnt)
			fmt.Println("eee", e.Name)

			// Can we see a majority of the federated servers?
			if cnt > len(e.Federated)/2 {
				// Reset the timeout and give up if we can't see a majority.
				go Fault(e, int(as.DBHeight), int(as.Minute))
				break messages
			}

			// Get the priority order list of audit servers in the priority order
			e.apriority = Order(e.Audit, e.DBHeight, e.Minute, e.electing, e.round[e.electing])

			idx := e.LeaderIndex(e.ServerID)
			// We are a leader
			if idx >= 0 {
				LeaderTimeout(e)
			}

			idx = e.AuditIndex(e.ServerID)
			if idx >= 0 {
				auditIdx := MaxIdx(e.apriority)
				if idx == auditIdx {
					V := new(elections.VolunteerAudit)
					V.NName = e.Name
					V.ServerIdx = e.electing
					V.ServerID = e.ServerID
					V.DBHeight = uint32(e.DBHeight)
					V.Minute = byte(e.Minute)
					V.SendOut(s, V)
				}
			}
		case *elections.VolunteerAudit:
			fmt.Printf("eee %s\n", msg.String())
		}
	}
}
