package main

import (
	"fmt"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/engine"
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
	state := engine.Factomd(params, true)

	CheckEntryBlocks(state.GetDB(), true)

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
