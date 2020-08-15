package testHelper

import (
	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/directoryBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

func CreateTestDirectoryBlock(prevBlock *directoryBlock.DirectoryBlock) *directoryBlock.DirectoryBlock {
	return CreateTestDirectoryBlockWithNetworkID(prevBlock, constants.LOCAL_NETWORK_ID)
}

func CreateTestDirectoryBlockWithNetworkID(prevBlock *directoryBlock.DirectoryBlock, networkID uint32) *directoryBlock.DirectoryBlock {
	dblock := new(directoryBlock.DirectoryBlock)

	dblock.SetHeader(CreateTestDirectoryBlockHeaderWithNetworkID(prevBlock, networkID))

	de := new(directoryBlock.DBEntry)
	de.ChainID = primitives.NewZeroHash()
	de.KeyMR = primitives.NewZeroHash()

	err := dblock.SetDBEntries(append(make([]interfaces.IDBEntry, 0, 5), de))
	if err != nil {
		panic(err)
	}
	//dblock.GetHeader().SetBlockCount(uint32(len(dblock.GetDBEntries())))

	return dblock
}

func CreateTestDirectoryBlockHeader(prevBlock *directoryBlock.DirectoryBlock) *directoryBlock.DBlockHeader {
	return CreateTestDirectoryBlockHeaderWithNetworkID(prevBlock, constants.LOCAL_NETWORK_ID)
}

func CreateTestDirectoryBlockHeaderWithNetworkID(prevBlock *directoryBlock.DirectoryBlock, networkID uint32) *directoryBlock.DBlockHeader {
	header := new(directoryBlock.DBlockHeader)

	header.SetBodyMR(primitives.Sha(primitives.NewZeroHash().Bytes()))
	header.SetBlockCount(0)
	header.SetNetworkID(networkID)

	if prevBlock == nil {
		header.SetDBHeight(0)
		header.SetPrevFullHash(primitives.NewZeroHash())
		header.SetPrevKeyMR(primitives.NewZeroHash())
		header.SetTimestamp(primitives.NewTimestampFromMinutes(1234))
	} else {
		header.SetDBHeight(prevBlock.Header.GetDBHeight() + 1)
		header.SetPrevFullHash(prevBlock.GetHash())
		keyMR, err := prevBlock.BuildKeyMerkleRoot()
		if err != nil {
			panic(err)
		}
		header.SetPrevKeyMR(keyMR)
		header.SetTimestamp(primitives.NewTimestampFromMinutes(prevBlock.Header.GetTimestamp().GetTimeMinutesUInt32() + 1))
	}

	header.SetVersion(1)

	return header
}
