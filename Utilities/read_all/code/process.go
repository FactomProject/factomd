package code

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	ecode "github.com/FactomProject/factomd/Utilities/extract_all/code"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
)

const FileIncrement = 2000

var Buff [1000000000]byte // Buffer to read a file
var buff []byte           // Slice to process the buffer
var fileNumber int
var fileCnt int

// Open
// Open a FactomObjects file.  Returns false if all files have been
// processed.
func Open() bool {
	u, err := user.Current()
	if err != nil {
		panic("no user")
	}
	filename := filepath.Join(u.HomeDir, "tmp", "FactomObjects", fmt.Sprintf("objects-%d.dat", fileNumber))
	f, err := os.OpenFile(filename, os.O_RDONLY, 07666)
	if err != nil {
		fmt.Println("Done. ", fileCnt, " files processed")
		return false
	}
	fileNumber += FileIncrement
	fileCnt++
	n, err := f.Read(Buff[:])
	buff = Buff[:n]
	f.Close()
	fmt.Println("Processing ", filename, " Reading ", n, " bytes.")
	return true
}

func Process() {

	header := new(ecode.Header)
	dBlock := directoryBlock.NewDirectoryBlock(nil)
	aBlock := adminBlock.NewAdminBlock(nil)
	fBlock := new(factoid.FBlock)
	ecBlock := entryCreditBlock.NewECBlock()
	eBlock := entryBlock.NewEBlock()
	entry := new(entryBlock.Entry)

	for Open() {

		tx := new(factoid.Transaction)
		_, _, _, _, _, _, _ = dBlock, aBlock, fBlock, ecBlock, eBlock, entry, tx
		for len(buff) > 0 {
			buff = header.UnmarshalBinary(buff)
			switch header.Tag {
			case ecode.TagDBlock:
				if err := dBlock.UnmarshalBinary(buff[:header.Size]); err != nil {
					panic("Bad Directory block")
				}
			case ecode.TagABlock:
				if err := aBlock.UnmarshalBinary(buff[:header.Size]); err != nil {
					fmt.Printf("Ht %d Admin size %d %v \n",
						dBlock.GetHeader().GetDBHeight(), header.Size, err)
				}
			case ecode.TagFBlock:
				if err := fBlock.UnmarshalBinary(buff[:header.Size]); err != nil {
					panic("Bad Factoid block")
				}
			case ecode.TagECBlock:
				if err := ecBlock.UnmarshalBinary(buff[:header.Size]); err != nil {
					panic("Bad Entry Credit block")
				}
			case ecode.TagEBlock:
				if _, err := eBlock.UnmarshalBinaryData(buff[:header.Size]); err != nil {
					panic("Bad Entry Block block")
				}
			case ecode.TagEntry:
				if err := entry.UnmarshalBinary(buff[:header.Size]); err != nil {
					panic("Bad Entry")
				}
			case ecode.TagTX:
				if err := tx.UnmarshalBinary(buff[:header.Size]); err != nil {
					panic("Bad Transaction")
				}
			default:
				panic("Unknown TX encountered")
			}
			buff = buff[header.Size:]
		}
	}
}
