package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

var CheckFloating bool
var UsingAPI bool
var FixIt bool

const level string = "level"
const bolt string = "bolt"

func main() {
	var (
		useApi        = flag.Bool("api", false, "Use API instead")
		checkFloating = flag.Bool("floating", false, "Check Floating")
		fix           = flag.Bool("fix", false, "Actually fix head")
	)

	flag.Parse()
	UsingAPI = *useApi
	CheckFloating = *checkFloating
	FixIt = *fix

	fmt.Println("Usage:")
	fmt.Println("CorrectChainHeads level/bolt/api DBFileLocation")
	fmt.Println("Program will fix chainheads")

	if len(flag.Args()) < 2 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(flag.Args()) > 2 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	if flag.Args()[0] == "api" {
		UsingAPI = true
	}

	var reader Fetcher

	if UsingAPI {
		reader = NewAPIReader(flag.Args()[1])
	} else {
		levelBolt := flag.Args()[0]

		if levelBolt != level && levelBolt != bolt {
			fmt.Println("\nFirst argument should be `level` or `bolt`")
			os.Exit(1)
		}
		path := flag.Args()[1]
		reader = NewDBReader(levelBolt, path)
	}

	// dblock, err := reader.FetchDBlockHead()

	FindHeads(reader)
}

func FindHeads(f Fetcher) {
	chainHeads := make(map[string]interfaces.IHash)

	var allEblockLock sync.Mutex
	allEblks := make(map[string]interfaces.IHash)

	var err error
	var dblock interfaces.IDirectoryBlock

	head, err := f.FetchDBlockHead()
	if err != nil {
		panic(fmt.Sprintf("Error fetching head"))
	}

	height := head.GetDatabaseHeight()
	dblock = head
	top := height
	fmt.Println("Starting at", height)
	errCount := 0
	waiting := new(int32)
	done := new(int32)
	total := 0

	var wg sync.WaitGroup
	allowedSimulataneous := 1000
	permission := make(chan bool, allowedSimulataneous)
	for i := 0; i < allowedSimulataneous; i++ {
		permission <- true
	}
	start := time.Now()

	doPrint := CheckFloating
	go func() {
		for {
			if !doPrint {
				return
			}
			time.Sleep(10 * time.Second)
			v := atomic.LoadInt32(waiting)
			d := atomic.LoadInt32(done)
			fmt.Printf("%d are still waiting. %d Done. Permission: %d\n", v, d, len(permission))
		}
	}()

	for ; height > 0; height-- {
		v := atomic.LoadInt32(waiting)
		for v > 50000 {
			time.Sleep(1 * time.Second)
		}

		dblock, err = f.FetchDBlockByHeight(height)
		if err != nil {
			fmt.Printf("Error fetching height %d: %s\n", height, err.Error())
			continue
		}

		eblockEnts := dblock.GetEBlockDBEntries()

		total += len(eblockEnts)
		if CheckFloating {
			for i := 0; i < len(eblockEnts); i++ {
				wg.Add(1)
				atomic.AddInt32(waiting, 1)
				func(eb interfaces.IDBEntry) {
					defer wg.Done()
					defer atomic.AddInt32(waiting, -1)
					<-permission
					defer func() {
						permission <- true
						atomic.AddInt32(done, 1)
					}()
					eblkF, err := f.FetchEBlock(eb.GetKeyMR())
					if err != nil {
						fmt.Printf("Error getting eblock %s for %s\n", eb.GetKeyMR().String(), eb.GetChainID().String())
						return
					}
					kmr, err := eblkF.KeyMR()
					if err != nil {
						fmt.Printf("Error getting eblock keymr %s for %s\n", eb.GetKeyMR().String(), eb.GetChainID().String())
						return
					}

					allEblockLock.Lock()
					allEblks[kmr.String()] = eblkF.GetHeader().GetPrevKeyMR()
					allEblockLock.Unlock()
				}(eblockEnts[i])
			}
		}

		for _, eblk := range eblockEnts {
			if _, ok := chainHeads[eblk.GetChainID().String()]; ok {
				// Chainhead already exists
				continue
			}
			chainHeads[eblk.GetChainID().String()] = eblk.GetKeyMR()
			ch, err := f.FetchHeadIndexByChainID(eblk.GetChainID())
			if err != nil {
				fmt.Printf("Error getting chainhead for %s\n", eblk.GetChainID().String())
			} else {
				if !ch.IsSameAs(eblk.GetKeyMR()) {
					fmt.Printf("ERROR: Chainhead found: %s, Expected %s :: For Chain: %s at height %d\n",
						ch.String(), eblk.GetKeyMR().String(), eblk.GetChainID().String(), height)
					errCount++
					if FixIt {
						f.SetChainHeads([]interfaces.IHash{eblk.GetKeyMR()}, []interfaces.IHash{eblk.GetChainID()})
					}
				}
			}
		}
		if height%500 == 0 {
			d := atomic.LoadInt32(done)
			ps := float64(top-height) / time.Since(start).Seconds()
			fmt.Printf("Currently on %d out of %d at %.3fp/s. %d EblocksPerHeight, %d done. %d ChainHeads so far. %d Are bad\n", height, top, ps, total, d, len(chainHeads), errCount)
		}

		var _ = dblock
	}

	if CheckFloating {
		wg.Wait()
	}
	doPrint = false

	fmt.Printf("%d Chains found in %f seconds", len(chainHeads), time.Since(start).Seconds())
	errCount = 0
	if CheckFloating {
		fmt.Println("Checking all EBLK links")
		for k, h := range chainHeads {
			var prev interfaces.IHash
			prev = h
			for {
				if prev.IsZero() {
					break
				}
				p, ok := allEblks[prev.String()]
				if !ok {
					errCount++
					fmt.Printf("Error finding Eblock %s for chain %s\n", h.String(), k)
				}
				prev = p
			}
		}
	}
	fmt.Printf("%d Errors found checking for bad links\n", errCount)

}

type Fetcher interface {
	FetchDBlockHead() (interfaces.IDirectoryBlock, error)
	FetchDBlockByHeight(dBlockHeight uint32) (interfaces.IDirectoryBlock, error)
	//FetchDBlock(hash interfaces.IHash) (interfaces.IDirectoryBlock, error)
	FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error)
	FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error)
	SetChainHeads(primaryIndexes, chainIDs []interfaces.IHash) error
}

func NewDBReader(levelBolt string, path string) *databaseOverlay.Overlay {
	var dbase *hybridDB.HybridDB
	var err error
	if levelBolt == bolt {
		dbase = hybridDB.NewBoltMapHybridDB(nil, path)
	} else {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			panic(err)
		}
	}

	dbo := databaseOverlay.NewOverlay(dbase)
	return dbo
}

type APIReader struct {
	location string
}

func NewAPIReader(loc string) *APIReader {
	a := new(APIReader)
	a.location = loc
	factom.SetFactomdServer(loc)

	return a
}

func (a *APIReader) SetChainHeads(primaryIndexes, chainIDs []interfaces.IHash) error {
	return nil
}

func (a *APIReader) FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	return nil, fmt.Errorf("Not implemented for api")
}

func (a *APIReader) FetchDBlockHead() (interfaces.IDirectoryBlock, error) {
	head, err := factom.GetDBlockHead()
	if err != nil {
		return nil, err
	}
	raw, err := factom.GetRaw(head)
	if err != nil {
		return nil, err
	}
	return rawBytesToblock(raw)
}

func (a *APIReader) FetchDBlockByHeight(dBlockHeight uint32) (interfaces.IDirectoryBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("d", int64(dBlockHeight))
	if err != nil {
		return nil, err
	}

	return rawRespToBlock(raw.RawData)
}

func (a *APIReader) FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error) {
	resp, err := factom.GetChainHead(chainID.String())
	if err != nil {
		return nil, err
	}
	return primitives.HexToHash(resp)
}

func rawBytesToblock(raw []byte) (interfaces.IDirectoryBlock, error) {
	dblock := directoryBlock.NewDirectoryBlock(nil)
	err := dblock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return dblock, nil
}

func rawRespToBlock(raw string) (interfaces.IDirectoryBlock, error) {
	by, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	return rawBytesToblock(by)
}
