// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/wsapi"
	"os"
	"strconv"
	"time"
)


type SimNetwork struct {
	Recieve() interfaces.IMsg
	Broadcast(interfaces.IMsg)
	SendToPeer(interfaces.IMsg)
    Control()
}

func() (sn *SimNetwork)	Recieve() interfaces.IMsg {
    return nil
}
func() (sn *SimNetwork)	Broadcast(interfaces.IMsg){
     return 
}
func() (sn *SimNetwork)	SendToPeer(interfaces.IMsg){
    return
}   
func() (sn *SimNetwork)	Control(){
    
}
