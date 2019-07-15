package testHelper

import (
	"github.com/FactomProject/factomd/wsapi"
	"strconv"
)

const defaultTestPort = 8080

func InitTestState() {
	state := CreateAndPopulateTestStateAndStartValidator()
	state.SetPort(defaultTestPort)

	if wsapi.Servers == nil {
		wsapi.Servers = make(map[string]*wsapi.Server)
	}

	port := strconv.Itoa(defaultTestPort)
	if wsapi.Servers[port] == nil {
		server := wsapi.InitServer(state)
		wsapi.Servers[port] = server
	}
}
