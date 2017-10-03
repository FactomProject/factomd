package state_test

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = fmt.Print

func TestProcessListString(t *testing.T) {
	// The string function is called in some unit tests, and lines that show offline nodes is sometimes hit. This
	// ensures coverage is consistent, despite it just being a String() call
	state := testHelper.CreateEmptyTestState()
	pl := NewProcessList(state, nil, 1)
	pl.VMs[0].List = append(pl.VMs[0].List, nil)
	pl.AddFedServer(primitives.NewHash([]byte("one")))
	pl.AddAuditServer(primitives.NewHash([]byte("two")))
	pl.AddFedServer(primitives.NewHash([]byte("three")))

	for _, f := range pl.FedServers {
		f.SetOnline(false)
	}

	var _ = pl.String()
}

func TestProcessListMisc(t *testing.T) {
	// The string  function is called in some unit tests, and lines that show offline nodes is sometimes hit. This
	// ensures coverage is consistent, despite it just being a String() call
	state := testHelper.CreateEmptyTestState()
	pl := NewProcessList(state, nil, 1)
	pl.VMs[0].List = append(pl.VMs[0].List, nil)
	pl.AddFedServer(primitives.NewHash([]byte("one")))
	pl.AddAuditServer(primitives.NewHash([]byte("two")))
	pl.AddFedServer(primitives.NewHash([]byte("three")))

	if pl.GetAmINegotiator() {
		t.Error("Should not be negotiator by default")
	}

	vmi := pl.VMIndexFor([]byte("one"))
	if vmi != 0 {
		t.Error("VMIndex should be 0")
	}

	fs := pl.FedServerFor(0, []byte("one"))
	if fs == nil {
		t.Error("No fed server associated with minute 0 byte slice")
	}

	wasReset := pl.Reset()
	if !wasReset {
		t.Error("Process List Reset did not work")
	}

	pl.TrimVMList(0, 0)
}

func TestProcessList_GetOldMsgs(t *testing.T) {
	type fields struct {
		DBHeight              uint32
		FactoidBalancesT      map[[32]byte]int64
		FactoidBalancesTMutex sync.Mutex
		ECBalancesT           map[[32]byte]int64
		ECBalancesTMutex      sync.Mutex
		State                 *State
		VMs                   []*VM
		ServerMap             [10][64]int
		System                VM
		SysHighest            int
		diffSigTally          int
		OldMsgs               map[[32]byte]interfaces.IMsg
		oldmsgslock           *sync.Mutex
		PendingChainHeads     *SafeMsgMap
		OldAcks               map[[32]byte]interfaces.IMsg
		oldackslock           *sync.Mutex
		NewEBlocks            map[[32]byte]interfaces.IEntryBlock
		neweblockslock        *sync.Mutex
		NewEntriesMutex       sync.RWMutex
		NewEntries            map[[32]byte]interfaces.IEntry
		AdminBlock            interfaces.IAdminBlock
		EntryCreditBlock      interfaces.IEntryCreditBlock
		DirectoryBlock        interfaces.IDirectoryBlock
		Matryoshka            []interfaces.IHash
		AuditServers          []interfaces.IServer
		FedServers            []interfaces.IServer
		AmINegotiator         bool
		DBSignatures          []DBSig
		DBSigAlreadySent      bool
		Requests              map[[32]byte]*Request
		NextHeightToProcess   [64]int
	}
	type args struct {
		key interfaces.IHash
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interfaces.IMsg
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessList{
				DBHeight:              tt.fields.DBHeight,
				FactoidBalancesT:      tt.fields.FactoidBalancesT,
				FactoidBalancesTMutex: tt.fields.FactoidBalancesTMutex,
				ECBalancesT:           tt.fields.ECBalancesT,
				ECBalancesTMutex:      tt.fields.ECBalancesTMutex,
				State:                 tt.fields.State,
				VMs:                   tt.fields.VMs,
				ServerMap:             tt.fields.ServerMap,
				System:                tt.fields.System,
				SysHighest:            tt.fields.SysHighest,
				OldMsgs:               tt.fields.OldMsgs,
				PendingChainHeads:     tt.fields.PendingChainHeads,
				OldAcks:               tt.fields.OldAcks,
				NewEBlocks:            tt.fields.NewEBlocks,
				NewEntriesMutex:       tt.fields.NewEntriesMutex,
				NewEntries:            tt.fields.NewEntries,
				AdminBlock:            tt.fields.AdminBlock,
				EntryCreditBlock:      tt.fields.EntryCreditBlock,
				DirectoryBlock:        tt.fields.DirectoryBlock,
				Matryoshka:            tt.fields.Matryoshka,
				AuditServers:          tt.fields.AuditServers,
				FedServers:            tt.fields.FedServers,
				AmINegotiator:         tt.fields.AmINegotiator,
				DBSignatures:          tt.fields.DBSignatures,
				DBSigAlreadySent:      tt.fields.DBSigAlreadySent,
				Requests:              tt.fields.Requests,
				NextHeightToProcess:   tt.fields.NextHeightToProcess,
			}
			if got := p.GetOldMsgs(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessList.GetOldMsgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
