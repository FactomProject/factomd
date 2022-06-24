package code

import (
	"errors"
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
	currentDBHeight := FactomdState.GetHighestSavedBlk()
	fmt.Println("Height Complete ", currentDBHeight)

	fmt.Println("Factoid Addresses:      ", FCTAccountCnt) // Add Totals
	fmt.Println("Entry Credit Addresses: ", ECAccountCnt)

	DBHeight := uint32(0)

	DBHeight = Open(DBHeight) // Finds where we were and starts there.
	StartHeight := DBHeight

	for {
		currentDBHeight = FactomdState.GetHighestSavedBlk()
		for i := uint32(0); DBHeight > currentDBHeight; i++ {
			fmt.Printf("Highest block: %d Want to process block: %d minute %d\n", currentDBHeight, DBHeight, i+1)
			time.Sleep(60 * time.Second)
			currentDBHeight = FactomdState.GetHighestSavedBlk()
		}
		// Do any needed output file updating, and
		// console feedback
		if DBHeight != StartHeight && (DBHeight%DBlockFreq == 0) {

			Open(DBHeight)

			if DBHeight > 0 {
				runtime := time.Now().Sub(Start).Seconds()
				seconds := int(runtime) % 60
				minutes := int(runtime) / 60
				bps := float64(DBHeight-StartHeight) / float64(runtime) // blocks per second
				stg := float64(currentDBHeight-DBHeight) / bps          // seconds to go
				ToGoseconds := int(stg) % 60
				ToGominutes := int(stg) / 60
				fmt.Printf("%5d:%02d ETA: %5d:%02d blocks/sec: %5d Height %7d/%d Entries %9d TXs %7d\n",
					minutes, seconds,
					ToGominutes, ToGoseconds,
					int(bps),
					DBHeight, currentDBHeight,
					EntryCnt, TXCnt,
				)
			}
		}

		DBlock, err := DB.FetchDBlockByHeight(DBHeight)
		if err != nil {
			panic("Bad DBHeight")
		}

		ProcessDirectory(DBlock)
		ProcessAdmin(DBHeight)
		ProcessECBlock(DBHeight)
		ProcessFactoids(DBHeight)
		ProcessEntries(DBlock)

		DBHeight++
	}
}

func Open(dbheight uint32) uint32 {
	inc := 0
	for {
		// Split up output files by Directory Block Height
		filename := fmt.Sprintf("objects-%d.dat", dbheight)
		filename = path.Join(FullDir, filename)
		_, err := os.Stat(filename)
		if errors.Is(err, os.ErrNotExist) {
			if dbheight > DBlockFreq && inc > 1 { // If we are catching up to a prior effort
				dbheight -= DBlockFreq
				filename = fmt.Sprintf("objects-%d.dat", dbheight)
				filename = path.Join(FullDir, filename)
				os.Remove(filename)
			}
			// Note that we create and overwrite the last file (which could be incomplete) if
			// it exists.
			if f, err := os.Create(filename); err != nil {
				panic(fmt.Sprintf("Could not open %s: %v", path.Join(FullDir, filename), err))
			} else {
				OutputFile.Close()
				OutputFile = f
			}
			return dbheight
		}
		dbheight += DBlockFreq
		inc++
	}
}

func ProcessDirectory(DBlock interfaces.IDirectoryBlock) {
	// Process a directory block

	DBlockBytes, err := DBlock.MarshalBinary()
	if err != nil {
		panic("Bad DBheight")
	}
	header := Header{Tag: TagDBlock, Size: uint64(len(DBlockBytes))}
	OutputFile.Write(header.MarshalBinary())
	OutputFile.Write(DBlockBytes)
}

func ProcessAdmin(dbheight uint32) {
	ABlock, err := DB.FetchABlockByHeight(dbheight)
	if err != nil {
		panic("Bad Admin Block")
	}
	ABlockBytes, err := ABlock.MarshalBinary()
	if err != nil {
		panic("Bad ABlock")
	}
	header := Header{Tag: TagABlock, Size: uint64(len(ABlockBytes))}
	OutputFile.Write(header.MarshalBinary())
	OutputFile.Write(ABlockBytes)
}

func ProcessECBlock(dbheight uint32) {

	switch {
	case dbheight >= 70387 && dbheight <= 70410:
		return
	}

	defer func() {
		if e := recover(); e != nil {
			fmt.Println("==================Error ECBlock", dbheight)
		}
	}()

	ECBlock, err := DB.FetchECBlockByHeight(dbheight)
	if err != nil {
		panic("Bad Entry Credit Block")
	}
	ECBlockBytes, err := ECBlock.MarshalBinary()
	if err != nil {
		print("Bad ECBlock")
	}
	header := Header{Tag: TagABlock, Size: uint64(len(ECBlockBytes))}
	OutputFile.Write(header.MarshalBinary())
	OutputFile.Write(ECBlockBytes)
}
