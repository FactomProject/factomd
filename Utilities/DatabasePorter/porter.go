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

	cfg := util.ReadConfig("", "")

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
			endKeyMR = dbHead.GetKeyMR().String()
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

		dBlockList = dBlockList[startIndex+1:]

		fmt.Printf("Saving blocks")

		for _, v := range dBlockList {
			dbo.StartMultiBatch()

			err = dbo.ProcessDBlockMultiBatch(v)
			if err != nil {
				panic(err)
			}

			entries := v.GetDBEntries()
			c := make(chan int, len(entries))
			for _, e := range entries {
				go func() {
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
						err = dbo.ProcessECBlockMultiBatch(ecblock)
						if err != nil {
							panic(err)
						}
						break
					default:
						eblock, err := GetEBlock(e.GetKeyMR().String())
						if err != nil {
							panic(err)
						}
						err = dbo.ProcessEBlockMultiBatch(eblock)
						if err != nil {
							panic(err)
						}
						eBlockEntries := eblock.GetEntryHashes()
						c2 := make(chan int, len(eBlockEntries))
						for _, eHash := range eBlockEntries {
							go func() {
								defer func() {
									c2 <- 1
								}()
								if eHash.IsMinuteMarker() == true {
									return
								}
								entry, err := GetEntry(eHash.String())
								if err != nil {
									fmt.Printf("Problem getting entry %v from block %v\n", eHash.String(), e.GetKeyMR().String())
									panic(err)
								}
								err = dbo.InsertEntry(entry)
								if err != nil {
									panic(err)
								}
							}()
						}
						for range eBlockEntries {
							<-c2
						}
						break
					}
				}()
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
	err := dbo.RebuildDirBlockInfo()
	if err != nil {
		panic(err)
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
