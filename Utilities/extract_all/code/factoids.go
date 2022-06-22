package code

import (
	"fmt"

	"github.com/FactomProject/factomd/common/factoid"
)

func ProcessFactoids(dbheight uint32) {
	ifblock, err := DB.FetchFBlockByHeight(dbheight)
	if err != nil {
		panic(fmt.Sprintf("Bad FCT block %v", err))
	}
	fblock := ifblock.(*factoid.FBlock)

	fblockBytes, err := fblock.MarshalBinary()
	if err != nil {
		print("Bad DBheight")
	}
	header := Header{Tag: TagFBlock, Size: uint64(len(fblockBytes))}
	OutputFile.Write(header.MarshalBinary())
	OutputFile.Write(fblockBytes)

	for _, t := range fblock.Transactions {
		txData, err := t.MarshalBinary()
		if err != nil {
			panic("Bad tx")
		}
		h := Header{Tag: TagTX, Size: uint64(len(txData))}
		OutputFile.Write(h.MarshalBinary())
		OutputFile.Write(txData)
		TXCnt++
	}
}
