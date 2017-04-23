package engine_test

import (
	"testing"

	"flag"
	. "github.com/FactomProject/factomd/engine"
	"time"
)

var _ = Factomd

func TestFactomdMain(t *testing.T) {
	{
		var svar string
		flag.StringVar(&svar, "svar", "bo", "a string var")
	}

	args := append([]string{},"-db=Map","-blktime=10","-count=5", "-logPort=37000","-port=37001","-ControlPanelPort=37002","-networkPort=37003")

	go Factomd(args)
	time.Sleep(20*time.Second)
}
