package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"
)

const level string = "level"
const bolt string = "bolt"

func main() {

	args := append([]string{},
		"-enablenet=false",
		"-logPort=37006",
		"-port=37007",
		"-ControlPanelPort=37008",
		"-networkPort=37009",
		"-startdelay=100")

	params := engine.ParseCmdLine(args)
	p := registry.New()
	p.Register(func(w *worker.Thread) {
		engine.Factomd(w, params, true)
	})
	go p.Run()
	p.WaitForRunning()

	CheckEntryBlocks(fnode.Get(0).State.GetDB(), true)

}

func CheckEntryBlocks(db interfaces.DBOverlaySimple, convertNames bool) error {
	head, err := db.FetchDBlockHead()
	blkCnt := 0
	if err == nil && head != nil {
		blkCnt = int(head.GetHeader().GetDBHeight())
	}

	for i := 0; i <= int(blkCnt); i++ {

		dblk, err := db.FetchDBlockByHeight(uint32(i))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Println(dblk.String())
	}
	return nil
}
