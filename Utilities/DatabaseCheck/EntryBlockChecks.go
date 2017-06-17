package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/hybridDB"
	"time"
)

const level string = "level"
const bolt string = "bolt"

func main() {

	err = CheckEntryBlocks(dbase, true)
	if err != nil {
		panic(err)
	}
}

func CheckEntryBlocks(db interfaces.ISCDatabaseOverlay, convertNames bool) error {
	head, err := db.FetchDBlockHead()
	if err == nil && head != nil {
		blkCnt = head.GetHeader().GetDBHeight()
	}

	last := time.Now()

	//msg, err := s.LoadDBState(blkCnt)
	start := s.GetDBHeightComplete()
	if start > 10 {
		start = start - 10
	}

	for i := int(start); i <= int(blkCnt); i++ {
		if i > 0 && i%1000 == 0 {
			bps := float64(1000) / time.Since(last).Seconds()
			os.Stderr.WriteString(fmt.Sprintf("%20s Loading Block %7d / %v. Blocks per second %8.2f\n", s.FactomNodeName, i, blkCnt, bps))
			last = time.Now()
		}

		msg, err := s.LoadDBState(uint32(i))
		if err != nil {
			s.Println(err.Error())
			os.Stderr.WriteString(fmt.Sprintf("%20s Error reading database at block %d: %s\n", s.FactomNodeName, i, err.Error()))
			break
		} else {
			if msg != nil {
				s.InMsgQueue().Enqueue(msg)
				msg.SetLocal(true)
				if s.InMsgQueue().Length() > 500 {
					for s.InMsgQueue().Length() > 100 {
						time.Sleep(10 * time.Millisecond)
					}
				}
			} else {
				// os.Stderr.WriteString(fmt.Sprintf("%20s Last Block in database: %d\n", s.FactomNodeName, i))
				break
			}
		}

		s.Print("\r", "\\|/-"[i%4:i%4+1])
	}

}

