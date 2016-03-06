package state_test

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"testing"
)

var _ = fmt.Print

func TestFedServer(t *testing.T) {
	pl := new(ProcessList)
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("1"))
		fmt.Println("Adding:", server.ChainID.String())
		pl.AddFedServer(server)
	}
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("1"))
		fmt.Println("Adding:", server.ChainID.String())
		pl.AddFedServer(server)
	}
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("2"))
		fmt.Println("Adding:", server.ChainID.String())
		pl.AddFedServer(server)
	}
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("2"))
		fmt.Println("Adding:", server.ChainID.String())
		pl.AddFedServer(server)
	}
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("2"))
		fmt.Println("Adding:", server.ChainID.String())
		pl.AddFedServer(server)
	}
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("3"))
		fmt.Println("Adding:", server.ChainID.String())
		pl.AddFedServer(server)
	}
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("4"))
		fmt.Println("Adding:", server.ChainID.String())
		pl.AddFedServer(server)
	}

	for _, fed := range pl.FedServers {
		fmt.Println("  ", fed.GetChainID().String())
	}

	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("2"))
		fmt.Println("Removing:", server.ChainID.String())
		pl.RemoveFedServer(server)
	}

	for _, fed := range pl.FedServers {
		fmt.Println("  ", fed.GetChainID().String())
	}

	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("2"))
		fmt.Println("Looking for:", server.ChainID.String())
		fmt.Println(pl.GetFedServerIndex(server.ChainID))
	}
	{
		server := new(interfaces.Server)
		server.ChainID = primitives.Sha([]byte("1"))
		fmt.Println("Looking for:", server.ChainID.String())
		fmt.Println(pl.GetFedServerIndex(server.ChainID))
	}
}
