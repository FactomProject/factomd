package cmd

import (
	"fmt"
	"os"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/factomd/Utilities/snapshot/stuff/snapshot"
)

const FileIncrement = 2000

type Objects struct {
	Buff [1000000000]byte // Buffer to read a file
	buff []byte // Slice to process the buffer
	fileNumber int
}

// Open
// Open a FactomObjects file.  Returns false if all files have been
// processed.
func (o *Objects) Open() bool {
	filename := fmt.Sprintf("../snapshot/FactomState/FactomObjects-%d.dat",o.fileNumber)
	f,err := os.OpenFile(filename,os.O_RDONLY,07666)
	if err != nil {
		fmt.Println("Done. ",o.fileNumber, " files processed")
		return false
	}
	o.fileNumber+=FileIncrement
	n,err := f.Read(o.Buff[:])
	o.buff = o.Buff[:n]
	f.Close()
	fmt.Println("Processing ",filename, " Reading ",n," bytes.")
	return true
}

func (o *Objects) Process() {
	for o.Open() {
		header := new(snapshot.ObjectHeader)
		dBlock := common.NewDBlock()
		eBlock := common.NewEBlock()
		entry  := common.NewEntry()
		_,_,_ = dBlock,eBlock,entry
		for len(o.buff)>0{
			o.buff = header.Unmarshal(o.buff)
			switch header.Tag {
			case snapshot.TagDirectoryBlock:
				if err := dBlock.UnmarshalBinary(o.buff);err != nil {
					panic("Bad Directory block")
				}
				o.buff = o.buff[header.Size:]
			case snapshot.TagEntryBlock:
				if _, err := eBlock.UnmarshalBinaryData(o.buff[:header.Size]);err != nil {
					panic("Bad Entry Block block")
				}
				o.buff = o.buff[header.Size:]
			case snapshot.TagEntry:
				if err := entry.UnmarshalBinary(o.buff[:header.Size]);err != nil {
					panic("Bad Entry")
				}
				o.buff = o.buff[header.Size:]
			}
		}
	}
}