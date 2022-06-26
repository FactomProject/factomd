package code

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/FactomProject/FactomCode/common"
	ecode "github.com/FactomProject/factomd/Utilities/extract_all/code"
)

const FileIncrement = 2000

var Buff [1000000000]byte // Buffer to read a file
var buff []byte           // Slice to process the buffer
var fileNumber int

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
		fmt.Println("Done. ", fileNumber, " files processed")
		return false
	}
	fileNumber += FileIncrement
	n, err := f.Read(Buff[:])
	buff = Buff[:n]
	f.Close()
	fmt.Println("Processing ", filename, " Reading ", n, " bytes.")
	return true
}

func Process() {
	for Open() {
		header := new(ecode.Header)
		dBlock := common.NewDBlock()
		eBlock := common.NewEBlock()
		entry := common.NewEntry()
		_, _, _ = dBlock, eBlock, entry
		for len(buff) > 0 {
			buff = header.UnmarshalBinary(buff)
			switch header.Tag {
			case ecode.TagDBlock:
				if err := dBlock.UnmarshalBinary(buff); err != nil {
					panic("Bad Directory block")
				}
				buff = buff[header.Size:]
			case ecode.TagEBlock:
				if _, err := eBlock.UnmarshalBinaryData(buff[:header.Size]); err != nil {
					panic("Bad Entry Block block")
				}
				buff = buff[header.Size:]
			case ecode.TagEntry:
				if err := entry.UnmarshalBinary(buff[:header.Size]); err != nil {
					panic("Bad Entry")
				}
				buff = buff[header.Size:]
			default:
				buff = buff[header.Size:]
			}
		}
	}
}
