// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	//"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/util"
)

var cfg *util.FactomdConfig
var dbo interfaces.DBOverlay

type Printout struct {
	FetchedBlock  uint32
	FetchingUntil uint32

	FilledBlock  uint32
	FillingUntil uint32

	SavedBlock  uint32
	SavingUntil uint32
}

var printout Printout
var doPrint bool

func PrintoutLoop() {
	for {
		time.Sleep(time.Second)
		if doPrint {
			fmt.Printf("Fetching\t%5d/%v\t\tFilling\t%5d/%v\t\tSaving\t%5d/%v\n",
				printout.FetchedBlock, printout.FetchingUntil, printout.FilledBlock, printout.FillingUntil, printout.SavedBlock, printout.SavingUntil)
		}
	}
}

func main() {
	var (
		fast           = flag.Bool("fast", false, "Do a random sampling for all blocks below the designated 'checkedUpTo' block")
		completedBlock = flag.Int("completed", 0, "Will only do a random sampling of entries below 'completed' block if 'fast' enabled")

		sampleRate = flag.Int("sampleRate", 10000, "Will sample 1/sampleRate entries below completedblock if fast enabled")
	)
	flag.Parse()

	if !(*fast) {
		fmt.Println("DatabasePorter")
	} else {
		fmt.Printf("DatabasePorter running in 'fast' mode. Will only check 1/%d entries below block %d for faster performance.\n", *sampleRate, *completedBlock)
	}

	cfg = util.ReadConfig("")

	if dbo != nil {
		dbo.Close()
	}
	switch cfg.App.DBType {
	case "Bolt":
		dbo = InitBolt(cfg)
		break
	case "LDB":
		dbo = InitLevelDB(cfg)
		break
	default:
		dbo = InitMapDB(cfg)
		break
	}

	dbHead, err := dbo.FetchDirectoryBlockHead()
	if err != nil {
		panic(err)
	}

	c := make(chan []interfaces.IDirectoryBlock, 5)
	done := make(chan int, 100)

	go SaveBlocksLoop(c, done)
	go PrintoutLoop()
	doPrint = true

	savedBatches := 0
	for _, keymr := range GetDBlockList() {
		endKeyMR := "0000000000000000000000000000000000000000000000000000000000000000"
		startIndex := 0
		if dbHead != nil {
			endKeyMR = dbHead.GetHeader().GetPrevKeyMR().String()
			//fmt.Printf("Local DB Head - %v - %v\n", dbHead.GetDatabaseHeight(), endKeyMR)
			startIndex = int(dbHead.GetDatabaseHeight())

			printout.FetchingUntil = dbHead.GetDatabaseHeight()
		}

		if keymr == endKeyMR {
			continue
		}

		dBlock, err := GetDBlock(keymr)
		if err != nil {
			panic(err)
		}
		if dbHead != nil {
			if dbHead.GetDatabaseHeight() > dBlock.GetDatabaseHeight() {
				continue
			}
		}

		if dBlock == nil {
			panic("dblock head not found")
		}
		nextHead := dBlock

		dBlockList := make([]interfaces.IDirectoryBlock, int(dBlock.GetDatabaseHeight())+1)
		dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock

		//fmt.Printf("\t\tFetching DBlocks\n")
		for {
			keymr = dBlock.GetHeader().GetPrevKeyMR().String()
			if keymr == endKeyMR {
				break
			}
			dBlock, err = GetDBlock(keymr)
			if err != nil {
				panic(err)
			}
			if dBlock == nil {
				panic("dblock " + keymr + " not found")
			}

			if dbHead != nil {
				if dbHead.GetDatabaseHeight() > dBlock.GetDatabaseHeight() {
					continue
				}
			}
			dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock
			printout.FetchedBlock = dBlock.GetDatabaseHeight()
			//fmt.Printf("Fetched dblock %v\n", dBlock.GetDatabaseHeight())
		}

		dBlockList = dBlockList[startIndex:]
		c <- dBlockList
		savedBatches++

		dbHead = nextHead
	}

	for i := 0; i < savedBatches; i++ {
		<-done
	}
	doPrint = false
	time.Sleep(time.Second)

	if dbo != nil {
		dbo.Close()
	}
	switch cfg.App.DBType {
	case "Bolt":
		dbo = InitBolt(cfg)
		break
	case "LDB":
		dbo = InitLevelDB(cfg)
		break
	default:
		dbo = InitMapDB(cfg)
		break
	}

	CheckDatabaseForMissingEntries(dbo, *fast, *completedBlock, *sampleRate)

	if dbo != nil {
		dbo.Close()
	}
	switch cfg.App.DBType {
	case "Bolt":
		dbo = InitBolt(cfg)
		break
	case "LDB":
		dbo = InitLevelDB(cfg)
		break
	default:
		dbo = InitMapDB(cfg)
		break
	}

	fmt.Printf("\t\tRebulding DirBlockInfo\n")
	err = dbo.RebuildDirBlockInfo()
	if err != nil {
		panic(err)
	}

	dbo.Close()
}

func SaveBlocksLoop(input chan []interfaces.IDirectoryBlock, done chan int) {
	//fmt.Printf("\t\tSaving blocks\n")

	dbChan := make(chan []BlockSet, 2)
	go SaveToDBLoop(dbChan, done)

	for {
		dBlockList := <-input
		blockSets := []BlockSet{}

		printout.FillingUntil = dBlockList[len(dBlockList)-1].GetDatabaseHeight()

		for _, v := range dBlockList {
			blockSet := BlockSet{}
			blockSet.DBlock = v

			entries := v.GetDBEntries()
			c := make(chan int, len(entries))
			for _, ent := range entries {
				go func(e interfaces.IDBEntry) {
					defer func() {
						c <- 1
					}()
					switch e.GetChainID().String() {
					case "000000000000000000000000000000000000000000000000000000000000000a":
						ablock, err := GetABlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						blockSet.ABlock = ablock
						break
					case "000000000000000000000000000000000000000000000000000000000000000f":
						fblock, err := GetFBlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						blockSet.FBlock = fblock
						break
					case "000000000000000000000000000000000000000000000000000000000000000c":
						ecblock, err := GetECBlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						blockSet.ECBlock = ecblock
						break
					default:
						eblock, err := GetEBlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						blockSet.Mutex.Lock()
						blockSet.EBlocks = append(blockSet.EBlocks, eblock)
						blockSet.Mutex.Unlock()
						eBlockEntries := eblock.GetEntryHashes()
						c2 := make(chan int, len(eBlockEntries))
						for _, eHash := range eBlockEntries {
							go func(ehash interfaces.IHash) {
								defer func() {
									c2 <- 1
								}()
								if ehash.IsMinuteMarker() == true {
									return
								}
								entry, err := GetEntry(ehash.String())
								if err != nil {
									fmt.Printf("Problem getting entry `%v` from block %v\n", ehash.String(), e.GetKeyMR().String())
									panic(err)
								}
								if entry != nil {
									blockSet.Mutex.Lock()
									blockSet.Entries = append(blockSet.Entries, entry)
									blockSet.Mutex.Unlock()
								}
							}(eHash)
						}
						for range eBlockEntries {
							<-c2
						}
						break
					}
				}(ent)
			}
			for range entries {
				<-c
			}

			blockSets = append(blockSets, blockSet)

			printout.FilledBlock = v.GetDatabaseHeight()

		}
		dbChan <- blockSets
	}
}

type BlockSet struct {
	DBlock  interfaces.IDirectoryBlock
	ABlock  interfaces.IAdminBlock
	ECBlock interfaces.IEntryCreditBlock
	FBlock  interfaces.IFBlock
	EBlocks []interfaces.IEntryBlock
	Entries []interfaces.IEBEntry

	Mutex sync.Mutex
}

func SaveToDBLoop(input chan []BlockSet, done chan int) {
	for {
		if dbo != nil {
			dbo.Close()
		}
		switch cfg.App.DBType {
		case "Bolt":
			dbo = InitBolt(cfg)
			break
		case "LDB":
			dbo = InitLevelDB(cfg)
			break
		default:
			dbo = InitMapDB(cfg)
			break
		}

		blockSet := <-input
		printout.SavingUntil = blockSet[len(blockSet)-1].DBlock.GetDatabaseHeight()

		for _, set := range blockSet {
			dbo.StartMultiBatch()

			err := dbo.ProcessDBlockMultiBatch(set.DBlock)
			if err != nil {
				panic(err)
			}

			err = dbo.ProcessABlockMultiBatch(set.ABlock)
			if err != nil {
				panic(err)
			}

			err = dbo.ProcessFBlockMultiBatch(set.FBlock)
			if err != nil {
				panic(err)
			}

			err = dbo.ProcessECBlockMultiBatch(set.ECBlock, true)
			if err != nil {
				panic(err)
			}

			for _, v := range set.EBlocks {
				err = dbo.ProcessEBlockMultiBatch(v, true)
				if err != nil {
					panic(err)
				}
			}

			for _, v := range set.Entries {
				err = dbo.InsertEntryMultiBatch(v)
				if err != nil {
					panic(err)
				}
			}

			if err := dbo.ExecuteMultiBatch(); err != nil {
				panic(err)
			}
			printout.SavedBlock = set.DBlock.GetDatabaseHeight()
		}

		done <- int(blockSet[len(blockSet)-1].DBlock.GetDatabaseHeight())
	}
}

var HashMap map[string]string

func CheckDatabaseForMissingEntries(dbo interfaces.DBOverlay, fast bool, completed int, sampleRate int) {
	fmt.Printf("\t\tIterating over DBlocks\n")

	prevD, err := dbo.FetchDBlockHead()
	if err != nil {
		panic(err)
	}

	HashMap = map[string]string{}

	for {
		if fast && prevD.GetDatabaseHeight() < uint32(completed) {
			// Do random sampling of entries
			CheckDBlockEntries(prevD, dbo, fast, sampleRate)
		} else {
			// Do full check (All entries)
			CheckDBlockEntries(prevD, dbo, false, sampleRate)
		}
		if prevD.GetDatabaseHeight()%1000 == 0 {
			fmt.Printf(" Checking block %v\n", prevD.GetDatabaseHeight())
		}
		HashMap[prevD.DatabasePrimaryIndex().String()] = "OK"

		if prevD.GetHeader().GetPrevKeyMR().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		dBlock, err := dbo.FetchDBlock(prevD.GetHeader().GetPrevKeyMR())
		if err != nil {
			panic(err)
		}
		if dBlock == nil {
			fmt.Printf("Found a missing dblock - %v\n", prevD.GetHeader().GetPrevKeyMR().String())
			dblock, err := GetDBlock(prevD.GetHeader().GetPrevKeyMR().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessDBlockBatchWithoutHead(dblock)
			if err != nil {
				panic(err)
			}
		} else {
			//only iterate to the next block if it was properly fetched from the database
			prevD = dBlock
		}
	}

	fmt.Printf("\t\tIterating over ECBlocks\n")
	prevEC, err := dbo.FetchECBlockHead()
	if err != nil {
		panic(err)
	}
	for {
		if prevEC.GetDatabaseHeight()%1000 == 0 {
			fmt.Printf(" Checking block %v\n", prevEC.GetDatabaseHeight())
		}
		HashMap[prevEC.DatabasePrimaryIndex().String()] = "OK"
		if prevEC.GetHeader().GetPrevHeaderHash().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		ecBlock, err := dbo.FetchECBlock(prevEC.GetHeader().GetPrevHeaderHash())
		if err != nil {
			panic(err)
		}
		if ecBlock == nil {
			fmt.Printf("Found a missing ecblock - %v\n", prevEC.GetHeader().GetPrevHeaderHash().String())
			ecblock, err := GetECBlock(prevEC.GetHeader().GetPrevHeaderHash().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessECBlockBatch(ecblock, true)
			if err != nil {
				panic(err)
			}
		} else {
			//only iterate to the next block if it was properly fetched from the database
			prevEC = ecBlock
		}
	}

	fmt.Printf("\t\tIterating over FBlocks\n")
	prevF, err := dbo.FetchFBlockHead()
	if err != nil {
		panic(err)
	}
	for {
		if prevF.GetDatabaseHeight()%1000 == 0 {
			fmt.Printf(" Checking block %v\n", prevF.GetDatabaseHeight())
		}
		HashMap[prevF.DatabasePrimaryIndex().String()] = "OK"
		if prevF.GetPrevKeyMR().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		fBlock, err := dbo.FetchFBlock(prevF.GetPrevKeyMR())
		if err != nil {
			panic(err)
		}
		if fBlock == nil {
			fmt.Printf("Found a missing fblock - %v\n", prevF.GetPrevKeyMR().String())
			fBlock, err := GetFBlock(prevF.GetPrevKeyMR().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessFBlockBatch(fBlock)
			if err != nil {
				panic(err)
			}
		} else {
			//only iterate to the next block if it was properly fetched from the database
			prevF = fBlock
		}
	}

	fmt.Printf("\t\tIterating over ABlocks\n")
	prevA, err := dbo.FetchABlockHead()
	if err != nil {
		panic(err)
	}
	for {
		if prevA.GetDatabaseHeight()%1000 == 0 {
			fmt.Printf(" Checking block %v\n", prevA.GetDatabaseHeight())
		}
		HashMap[prevA.DatabasePrimaryIndex().String()] = "OK"
		if prevA.GetHeader().GetPrevBackRefHash().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		aBlock, err := dbo.FetchABlock(prevA.GetHeader().GetPrevBackRefHash())
		if err != nil {
			panic(err)
		}
		if aBlock == nil {
			fmt.Printf("Found a missing ablock - %v\n", prevA.GetHeader().GetPrevBackRefHash().String())
			aBlock, err := GetABlock(prevA.GetHeader().GetPrevBackRefHash().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessABlockBatch(aBlock)
			if err != nil {
				panic(err)
			}
		} else {
			//only iterate to the next block if it was properly fetched from the database
			prevA = aBlock
		}
	}

	fmt.Printf("\t\tFinding unused blocks\n")

	hashes, err := dbo.FetchAllDBlockKeys()
	if err != nil {
		panic(err)
	}
	for _, h := range hashes {
		if HashMap[h.String()] == "" {
			fmt.Printf("Superfluous DBlock - %v\n", h)
			dbo.Delete(databaseOverlay.DIRECTORYBLOCK, h.Bytes())
		}
	}

	hashes, err = dbo.FetchAllABlockKeys()
	if err != nil {
		panic(err)
	}
	for _, h := range hashes {
		if HashMap[h.String()] == "" {
			fmt.Printf("Superfluous ABlock - %v\n", h)
			dbo.Delete(databaseOverlay.ADMINBLOCK, h.Bytes())
		}
	}

	hashes, err = dbo.FetchAllECBlockKeys()
	if err != nil {
		panic(err)
	}
	for _, h := range hashes {
		if HashMap[h.String()] == "" {
			fmt.Printf("Superfluous ECBlock - %v\n", h)
			dbo.Delete(databaseOverlay.ENTRYCREDITBLOCK, h.Bytes())
		}
	}

	hashes, err = dbo.FetchAllFBlockKeys()
	if err != nil {
		panic(err)
	}
	for _, h := range hashes {
		if HashMap[h.String()] == "" {
			fmt.Printf("Superfluous FBlock - %v\n", h)
			dbo.Delete(databaseOverlay.FACTOIDBLOCK, h.Bytes())
		}
	}
}

func CheckDBlockEntries(dBlock interfaces.IDirectoryBlock, dbo interfaces.DBOverlay, fast bool, sampleRate int) {
	entries := dBlock.GetDBEntries()
	for {
		missing := 0
		for _, e := range entries {
			HashMap[e.GetKeyMR().String()] = "OK"
			switch e.GetChainID().String() {
			case "000000000000000000000000000000000000000000000000000000000000000a":
				aBlock, err := dbo.FetchABlock(e.GetKeyMR())
				if err != nil {
					panic(err)
				}
				if aBlock != nil {
					break
				}
				fmt.Printf("Found missing aBlock in #%v\n", dBlock.GetDatabaseHeight())
				missing++
				aBlock, err = GetABlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessABlockBatch(aBlock)
				if err != nil {
					panic(err)
				}
				break
			case "000000000000000000000000000000000000000000000000000000000000000f":
				fBlock, err := dbo.FetchFBlock(e.GetKeyMR())
				if err != nil {
					panic(err)
				}
				if fBlock != nil {
					break
				}
				fmt.Printf("Found missing fBlock in #%v\n", dBlock.GetDatabaseHeight())
				missing++
				fBlock, err = GetFBlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessFBlockBatch(fBlock)
				if err != nil {
					panic(err)
				}
				break
			case "000000000000000000000000000000000000000000000000000000000000000c":
				ecBlock, err := dbo.FetchECBlock(e.GetKeyMR())
				if err != nil {
					panic(err)
				}
				if ecBlock != nil {
					break
				}
				fmt.Printf("Found missing ecBlock in #%v\n", dBlock.GetDatabaseHeight())
				missing++
				ecBlock, err = GetECBlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessECBlockBatch(ecBlock, true)
				if err != nil {
					panic(err)
				}
				break
			default:
				eBlock, err := dbo.FetchEBlock(e.GetKeyMR())
				if err != nil {
					if err.Error() != "EOF" {
						panic(err)
					}
				}
				if eBlock == nil {
					fmt.Printf("Found missing eBlock in #%v\n", dBlock.GetDatabaseHeight())
					missing++
					eBlock, err = GetEBlock(e.GetKeyMR().String())
					if err != nil {
						panic(err)
					}
					err = dbo.ProcessEBlockBatch(eBlock, true)
					if err != nil {
						panic(err)
					}
				}

				eBlockEntries := eBlock.GetEntryHashes()
				// Used to determine random sampling if fast==true
				entryCount := 0
				// Not going to use the loop counter, as it also includes minute marks. We
				// Do not want to include those in our random sampling
				for _, eHash := range eBlockEntries {
					if eHash.IsMinuteMarker() == true {
						continue
					}
					entryCount++ // Not a minute marker, AKA an entry
					if fast && entryCount%sampleRate != 0 {
						continue
					}
					entry, err := dbo.FetchEntry(eHash)
					if err != nil {
						panic(err)
					}
					if entry == nil {
						fmt.Printf("Found missing entry in #%v\n", dBlock.GetDatabaseHeight())
						missing++
						entry, err := GetEntry(eHash.String())
						if err != nil {
							fmt.Printf("Problem getting entry `%v` from block %v\n", eHash.String(), e.GetKeyMR().String())
							panic(err)
						}
						err = dbo.InsertEntry(entry)
						if err != nil {
							panic(err)
						}
					}
				}
				break
			}
		}
		if missing == 0 {
			break
		}
	}
}

func GetDBlockList() []string {
	/*return []string{
		"3a5ec711a1dc1c6e463b0c0344560f830eb0b56e42def141cb423b0d8487a1dc", //10
		"cde346e7ed87957edfd68c432c984f35596f29c7d23de6f279351cddecd5dc66", //100
		"d13472838f0156a8773d78af137ca507c91caf7bf3b73124d6b09ebb0a98e4d9", //200
		"2978233e69cf207a92bac162598a0398c408caecec7092151db5d044587af5d6", //500
	}*/

	keymr, err := GetDBlockHead()
	if err != nil {
		panic(err)
	}
	return []string{
		"3a5ec711a1dc1c6e463b0c0344560f830eb0b56e42def141cb423b0d8487a1dc", //10
		"cde346e7ed87957edfd68c432c984f35596f29c7d23de6f279351cddecd5dc66", //100
		"d13472838f0156a8773d78af137ca507c91caf7bf3b73124d6b09ebb0a98e4d9", //200
		"2978233e69cf207a92bac162598a0398c408caecec7092151db5d044587af5d6", //500

		"cd45e38f53c090a03513f0c67afb93c774a064a5614a772cd079f31b3db4d011", //1000
		"0fae4e8749045bcec480a47019ab2423ac8339d33447cc1f7978395a841b6f55", //2000
		"599c1e4527cf5880210d21f7aa1063aea68dd1a985d65ba037c57acc433e867e", //3000
		"3163946232e9e8ec22b21a9db1373c172ebf7a7993dc54c1a0f41f4251e8d7f5", //4000
		"ca02b78949b80427ddecbbf266d0c18b5dbbbfd840e5d505d64147fa109bf29e", //5000
		"d1975eb7bd7e0002f7f4d77469a95b466340556ae0461135d7f9469b9eec173e", //6000
		"91a523f521e910a870c64c155076ceb203b210d009d34e50cc441194ee621de8", //7000
		"4e2d73d19959240c491df3edc03d02f6a0a2a05f7b75e1b4fe7f299636f073c7", //8000
		"19da7a9a36dc146c740ab4fc4bbf53b25441fdd8925eda29a8b870cda81a1bb5", //9000

		"3670a63eb8051b925213a4a350e8d37d87e43da8a577a609d7fd30629b73a3aa", //10000
		"53b0884fc5bb9de48db83b5e66d4ca1cb1d29dcb15e865903c37f9137dfe2cc8", //11000
		"c4045aaf92e71ae3c25135022fb2d777164a6530fdd279ea6d5c0d383e87d60d", //12000
		"cff9894749008b42874e6bea9da33d6ba0f7ff85405a10f3b980a900d9618104", //13000
		"491809dc5a07ae895a9aca8634113267a4e38be11bb980012a244cd6559e8308", //14000
		"4358041d6773351dd0a42a8d16778c6544b1196a03c6c41645340cd076a29b6b", //15000
		"1c5a3ea4233871c564b0cae1c01201bafa3f84d3477532bb5318b72fdbc51804", //16000
		"7af7d416d96bbdd0acbeada4b3a70d2e400ec11706db742b6c7c8f60adb35b49", //17000
		"b5837c846cc314c9cedde7a0f0633b9d4f278a867a5808de1bf9da1a6c06e795", //18000
		"8bf15f172bd03f13db2937c28fd333e867faa39a55c09f095f84be6658ff8cea", //19000

		"623f18fb113dca850b78389fab662033f86d65a0efc2f1760f11939f3a8df98a", //20000
		"1bb09820f0c4650d53b3be362242eca5284d43ae74ea88abafc82b5761b1bfce", //21000
		"560312328e848d9e7680370c6e54480389e6ae692ea6ae23a4253bdaf8bace15", //22000
		"91d1fa6b470235cdd334436173620472feecda86cbb122df693066372e411008", //23000
		"23a16b202dc818001457483fcbfab1417aed727aabbb156a4f6ffa82f2736ef2", //24000
		"be8f161e0ffa2e3d50cdbded924ee47e9419bf52b900a4150ef74d9016dfd1c0", //25000
		"7395266728a716c9f4d6f994da64cb9f91a4b4e804cf601dcd371a679d38400a", //26000
		"9b765658a2e5ff36d28f335eacd474d9bfdae8112ae348d2674cb3710fdb9b3b", //27000
		"60f250826550092034003cc1c06c9c33a33bf59733800c6a5e4bac094b62b2d9", //28000
		"0b1e4bdeb1098d590813ce44b9eb256181e5513669cdf93e56f135e2c80cb880", //29000

		"50d9cf6c596a09d0fab37467601a4caa5c7d1b5fd2ee007af25646f6152a392b", //30000
		"b350499dfbb973455385e5b826de77c3e5efcbea0e6388cecf64134416f47c1b", //31000
		"c49cc7d2de2b10feb1c6dafd598485dda0a67fc7584756dc465bacb8bf05090e", //32000
		"eb8e40327d6a60b00e4d23255b29c3b12f8b4093e7f0b266cb9dd25e32ab297d", //33000
		"0879a4866628e0eeab98e479f59ce36d776c17d56849e4d9678182847b5e6689", //34000
		"e630fbd538efcdece0b134ba93b719072d871297e233abda7e79dcc5f3bba9f5", //35000
		"f421b36795edf9b60a74c8f2342e836f5a1693a6cff3e332a886a4783415cacd", //36000
		"9660a52b7e130862d1c043562310f6c98eb5ac9999c827615e73a4693cd75f99", //37000
		"faf8ace8a60a68ff7d2e9b02b145b2958a7ba9c13bf05d55bf12dc8f94bc7c6c", //38000
		"a08aad2aea05be4a4fd4583f068af28601ce9d905c08d07434f2aca1865e6a3e", //39000

		"df11b01490dd5f7e8a849205aa72b56158f3022c33b0075677c747ae4c2cac65", //40000
		"01a9fb685df848f887281decbc2446fd22490305bffbdeb065d937deb34147c3", //41000
		"03497a662826e06645d60c97a8af6e4044654d5513a4c3b8593940973116fb9f", //42000
		"ee21501e47c5307d3bcee7f730fe878f5b331126df688d7c20410b9d75fb6739", //43000
		"6f43842101f04bf6cd62aaa1ba3e98591ec89f31b41c043c3b5ed333e4be7918", //44000
		"87e17c6740c84088d3e9d6d49c51c5999d7a29ae66b73b270661d2ae18b36d11", //45000
		"09d3e1fce45d296f4bc299471ab5edcb2cd3a366c71402230afcf3f69dfe7a9d", //46000
		"acbd841298e84c8edb72db09433d3419964631da21824cdd94c1a1d9bff5ccf3", //47000
		"b5885b4780dd63950a9d69236d57b9d505e1059e0038d26167daaa30e1137120", //48000
		"036fc982d9d534cf31d4d16419964c4cc00c6ad13bf39dc90e0ea5d59dc57d01", //49000

		"5abd0dd2b470c40afd864796a9408fe9a9ed46c360672387cb6a8a09057d3ef3", //50000
		"5b8dfb559c03ab0f73e045c512c92a06d927363cb40b40e42cb746a7884cc8af", //51000
		"6296a645d03e22cfa769ce48ab735e69789f026ca65d20f26c2a8adc2e9bf630", //52000
		"792bce3c65bab4321db09b3f5b017b5496dd217352858215ed058faaa00aefeb", //53000
		"11f44cadaf19eea29dc366a828531e36e8137ea2a5687041cf51e5ad66a7233d", //54000
		"1101b4c1003a393bc17bae9305b103d07ef7b21bc5f170cf41c7e07e8862e6ad", //55000
		"1e893bf343234de2a31192a0e5fff02f785989afd99823c1d37de5af76c5be45", //56000
		"28de098d3c84249e070736b923d026c6f6b432b60efb89678249baf28c6ae9ac", //57000
		"c1345922478345fb4124591583bbf369dcd2b5e7e3aadb0941233285bb89f993", //58000
		"b8d0837c26aacb6355a13aac8d80073f855fdf8c0cf2f528f7daf3258044c8b3", //59000

		keymr}
}
