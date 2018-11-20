package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (s *State) CreateEOM(force bool, m interfaces.IMsg, vmIdx int) (eom *messages.EOM, ack interfaces.IMsg) {

	if m == nil || m.(*messages.EOM) == nil {
		eom = new(messages.EOM)
	} else {
		eom = m.(*messages.EOM)
	}

	eom.Timestamp = s.GetTimestamp()
	eom.ChainID = s.GetIdentityChainID()
	eom.Sign(s)
	eom.SetLocal(true)

	pl := s.ProcessLists.Get(s.LLeaderHeight)
	vm := pl.VMs[vmIdx]

	// Put the System Height and Serial Hash into the EOM
	eom.SysHeight = uint32(pl.System.Height)

	if !force && s.Syncing && vm.Synced {
		return nil, nil
	} else if !s.Syncing {
		//todo: What if we started syncing some other VM already before we get around to making our EOM?
		//s.Syncing = true
		////fmt.Println(fmt.Sprintf("EOM PROCESS: %10s LeaderExecuteEOM: !s.EOM(%v)", s.FactomNodeName, s.EOM))
		//s.EOM = true
		//s.EOMsyncing = true
		//s.EOMProcessed = 0
		//for _, vm := range pl.VMs {
		//	vm.Synced = false
		//}
		//s.EOMLimit = len(pl.FedServers)
		s.EOMMinute = int(s.CurrentMinute)
	}

	if !force && vm.EomMinuteIssued >= s.CurrentMinute+1 {
		//os.Stderr.WriteString(fmt.Sprintf("Bump detected %s minute %2d\n", s.FactomNodeName, s.CurrentMinute))
		return nil, nil
	}

	//_, vmindex := pl.GetVirtualServers(s.EOMMinute, s.IdentityChainID)

	eom.DBHeight = s.LLeaderHeight
	eom.VMIndex = vmIdx
	// EOM.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(s.CurrentMinute)
	vm.EomMinuteIssued = s.CurrentMinute + 1
	eom.Sign(s)
	eom.MsgHash = nil
	ack = s.NewAck(eom, nil).(*messages.Ack)
	eom.MsgHash = nil
	eom.RepeatHash = nil
	return eom, ack
}
