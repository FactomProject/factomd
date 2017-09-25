package electionMsgs

import (
	"bytes"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
)

var _ = fmt.Print

func Sort(serv []interfaces.IServer) {
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

func Fault(e *elections.Elections, dbheight int, minute int, round int) {
	/*
		fmt.Printf("eee %10s %20s %20s dbheight %d minute %d round %d\n",
			e.Name,
			"Fault Start",
			time.Now().String(),
			dbheight,
			minute,
			round)
	*/

	time.Sleep(10 * time.Second)

	timeout := new(TimeoutInternal)
	timeout.Minute = minute
	timeout.DBHeight = dbheight
	timeout.Round = round
	e.Input.Enqueue(timeout)

}
