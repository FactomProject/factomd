// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/util"
)

func main() {
	fmt.Println("DatabasePorter")

	cfg := util.ReadConfig("")

	var dbo interfaces.DBOverlay

mainloop:
	for _, keymr := range GetDBlockList() {
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

		fmt.Printf("dbo - %v\n", dbo)

		dbHead, err := dbo.FetchDirectoryBlockHead()
		if err != nil {
			panic(err)
		}
		endKeyMR := "0000000000000000000000000000000000000000000000000000000000000000"
		startIndex := 0
		if dbHead != nil {
			endKeyMR = dbHead.GetHeader().GetPrevKeyMR().String()
			fmt.Printf("Local DB Head - %v - %v\n", dbHead.GetDatabaseHeight(), endKeyMR)
			startIndex = int(dbHead.GetDatabaseHeight())
		}

		if keymr == endKeyMR {
			continue
		}

		dBlock, err := GetDBlock(keymr)
		if err != nil {
			panic(err)
		}
		if dBlock == nil {
			panic("dblock head not found")
		}
		dBlockList := make([]interfaces.IDirectoryBlock, int(dBlock.GetDatabaseHeight())+1)
		dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock

		fmt.Printf("\t\tFetching DBlocks\n")
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
					continue mainloop
				}
			}
			dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock
			fmt.Printf("Fetched dblock %v\n", dBlock.GetDatabaseHeight())
		}

		dBlockList = dBlockList[startIndex:]

		fmt.Printf("\t\tSaving blocks\n")

		for _, v := range dBlockList {
			dbo.StartMultiBatch()

			err = dbo.ProcessDBlockMultiBatch(v)
			if err != nil {
				panic(err)
			}

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
						err = dbo.ProcessABlockMultiBatch(ablock)
						if err != nil {
							panic(err)
						}
						break
					case "000000000000000000000000000000000000000000000000000000000000000f":
						fblock, err := GetFBlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						err = dbo.ProcessFBlockMultiBatch(fblock)
						if err != nil {
							panic(err)
						}
						break
					case "000000000000000000000000000000000000000000000000000000000000000c":
						ecblock, err := GetECBlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						err = dbo.ProcessECBlockMultiBatch(ecblock, true)
						if err != nil {
							panic(err)
						}
						break
					default:
						eblock, err := GetEBlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						err = dbo.ProcessEBlockMultiBatch(eblock, true)
						if err != nil {
							panic(err)
						}
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
								err = dbo.InsertEntry(entry)
								if err != nil {
									panic(err)
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

			if err := dbo.ExecuteMultiBatch(); err != nil {
				panic(err)
			}
			fmt.Printf("Saved block height %v\n", v.GetDatabaseHeight())
		}
	}
	fmt.Printf("\t\tIterating over ECBlocks\n")
	prevEC, err := dbo.FetchECBlockHead()
	if err != nil {
		panic(err)
	}
	for {
		if prevEC.GetHeader().GetPrevHeaderHash().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		ecBlock, err := dbo.FetchECBlock(prevEC.GetHeader().GetPrevHeaderHash())
		if err != nil {
			panic(err)
		}
		if ecBlock == nil {
			fmt.Printf("Found a missing block - %v\n", prevEC.GetHeader().GetPrevHeaderHash().String())
			ecblock, err := GetECBlock(prevEC.GetHeader().GetPrevHeaderHash().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessECBlockBatchWithoutHead(ecblock, true)
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
		if prevF.GetPrevKeyMR().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		fBlock, err := dbo.FetchFBlock(prevF.GetPrevKeyMR())
		if err != nil {
			panic(err)
		}
		if fBlock == nil {
			fmt.Printf("Found a missing block - %v\n", prevF.GetPrevKeyMR().String())
			fBlock, err := GetFBlock(prevF.GetPrevKeyMR().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessFBlockBatchWithoutHead(fBlock)
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
		if prevA.GetHeader().GetPrevBackRefHash().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		aBlock, err := dbo.FetchABlock(prevA.GetHeader().GetPrevBackRefHash())
		if err != nil {
			panic(err)
		}
		if aBlock == nil {
			fmt.Printf("Found a missing block - %v\n", prevA.GetHeader().GetPrevBackRefHash().String())
			aBlock, err := GetABlock(prevA.GetHeader().GetPrevBackRefHash().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessABlockBatchWithoutHead(aBlock)
			if err != nil {
				panic(err)
			}
		} else {
			//only iterate to the next block if it was properly fetched from the database
			prevA = aBlock
		}
	}

	fmt.Printf("\t\tIterating over DBlocks\n")
	prevD, err := dbo.FetchDBlockHead()
	if err != nil {
		panic(err)
	}
	for {
		CheckDBlockEntries(prevD, dbo)

		if prevD.GetHeader().GetPrevKeyMR().String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		dBlock, err := dbo.FetchDBlock(prevD.GetHeader().GetPrevKeyMR())
		if err != nil {
			panic(err)
		}
		if dBlock == nil {
			fmt.Printf("Found a missing block - %v\n", prevD.GetHeader().GetPrevKeyMR().String())
			ecblock, err := GetDBlock(prevD.GetHeader().GetPrevKeyMR().String())
			if err != nil {
				panic(err)
			}
			err = dbo.ProcessDBlockBatchWithoutHead(ecblock)
			if err != nil {
				panic(err)
			}
		} else {
			//only iterate to the next block if it was properly fetched from the database
			prevD = dBlock
		}
	}

	fmt.Printf("\t\tRebulding DirBlockInfo\n")
	err = dbo.RebuildDirBlockInfo()
	if err != nil {
		panic(err)
	}
}

func CheckDBlockEntries(dBlock interfaces.IDirectoryBlock, dbo interfaces.DBOverlay) {
	entries := dBlock.GetDBEntries()
	for {
		missing := 0
		for _, e := range entries {
			switch e.GetChainID().String() {
			case "000000000000000000000000000000000000000000000000000000000000000a":
				aBlock, err := dbo.FetchABlock(e.GetKeyMR())
				if err != nil {
					panic(err)
				}
				if aBlock != nil {
					break
				}
				fmt.Printf("Found missing aBlock")
				missing++
				aBlock, err = GetABlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessABlockBatchWithoutHead(aBlock)
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
				fmt.Printf("Found missing fBlock")
				missing++
				fBlock, err = GetFBlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessFBlockBatchWithoutHead(fBlock)
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
				fmt.Printf("Found missing ecBlock")
				missing++
				ecBlock, err = GetECBlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessECBlockBatchWithoutHead(ecBlock, true)
				if err != nil {
					panic(err)
				}
				break
			default:
				eBlock, err := dbo.FetchEBlock(e.GetKeyMR())
				if err != nil {
					panic(err)
				}
				if eBlock == nil {
					fmt.Printf("Found missing eBlock")
					missing++
					eBlock, err = GetEBlock(e.GetKeyMR().String())
					if err != nil {
						panic(err)
					}
					err = dbo.ProcessEBlockBatchWithoutHead(eBlock, true)
					if err != nil {
						panic(err)
					}
				}

				eBlockEntries := eBlock.GetEntryHashes()
				for _, eHash := range eBlockEntries {
					if eHash.IsMinuteMarker() == true {
						return
					}
					entry, err := dbo.FetchEntry(eHash)
					if err != nil {
						panic(err)
					}
					if entry == nil {
						fmt.Printf("Found missing entry")
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
		"264394dffc9be2bc408c12b54d98f053b3def493804f7beafac99f7bfaa95ddb", //2500
		"ca02b78949b80427ddecbbf266d0c18b5dbbbfd840e5d505d64147fa109bf29e", //5000
		"3670a63eb8051b925213a4a350e8d37d87e43da8a577a609d7fd30629b73a3aa", //10000
		"4358041d6773351dd0a42a8d16778c6544b1196a03c6c41645340cd076a29b6b", //15000
		"623f18fb113dca850b78389fab662033f86d65a0efc2f1760f11939f3a8df98a", //20000
		"be8f161e0ffa2e3d50cdbded924ee47e9419bf52b900a4150ef74d9016dfd1c0", //25000
		"50d9cf6c596a09d0fab37467601a4caa5c7d1b5fd2ee007af25646f6152a392b", //30000
		"e630fbd538efcdece0b134ba93b719072d871297e233abda7e79dcc5f3bba9f5", //35000
		keymr}
}
