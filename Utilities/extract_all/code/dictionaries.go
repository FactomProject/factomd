package code

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
)

const (
	DBlockFreq = uint32(2000)
	OutputDir  = "/tmp/FactomObjects"
)

var FactomdState interfaces.IState
var DB interfaces.DBOverlaySimple
var OutputFile *os.File
var EntryCnt uint64
var TXCnt uint64
var Start time.Time
var FullDir string
var FCTAccountCnt uint64
var ECAccountCnt uint64

func ProcessDictionaries() {

	_, _ = EntryCnt, TXCnt
	Start = time.Now()

	DB = FactomdState.GetDB()
	dbHeight := FactomdState.GetDBHeightComplete()
	fmt.Println("Height Complete ", dbHeight)

	fmt.Println("Factoid Addresses:      ", FCTAccountCnt)
	fmt.Println("Entry Credit Addresses: ", ECAccountCnt)

	for i := uint32(0); i < dbHeight; i++ {
		// Do any needed output file updating, and
		// console feedback
		if i%DBlockFreq == 0 {

			// Split up output files by Directory Block Height
			filename := fmt.Sprintf("objects-%d.dat", i)
			if f, err := os.Create(path.Join(FullDir, filename)); err != nil {
				panic(fmt.Sprintf("Could not open %s: %v", path.Join(FullDir, filename), err))
			} else {
				OutputFile = f
			}
			if i > 0 {
				runtime := time.Now().Sub(Start).Seconds()
				seconds := int(runtime) % 60
				minutes := int(runtime) / 60
				bps := float64(i) / float64(runtime) // blocks per second
				stg := float64(dbHeight-i) / bps     // seconds to go
				ToGoseconds := int(stg) % 60
				ToGominutes := int(stg) / 60
				fmt.Printf("%5d:%02d ETA: %5d:%02d blocks/sec: %5d Height %7d/%d Entries %9d TXs %7d\n",
					minutes, seconds,
					ToGominutes,ToGoseconds, 
					int(bps),
					i, dbHeight,
					EntryCnt, TXCnt,
				)
			}
		}
		DBlock, err := DB.FetchDBlockByHeight(i)
		if err != nil {
			panic("Bad DBHeight")
		}

		ProcessDirectory(DBlock)
		ProcessFactoids(i)
		ProcessEntries(DBlock)
	}
}

func ProcessDirectory(DBlock interfaces.IDirectoryBlock) {
	// Process a directory block

	DBlockBytes, err := DBlock.MarshalBinary()
	if err != nil {
		print("Bad DBheight")
	}
	header := Header{Tag: TagDBlock, Size: uint64(len(DBlockBytes))}
	OutputFile.Write(header.MarshalBinary())
	OutputFile.Write(DBlockBytes)
}
