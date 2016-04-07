// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
    "time"
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"os"
)

func LoadJournal(s interfaces.IState, journal string) {

	f, err := os.Open(journal)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 4*1024)
    i := 0
	for {
        fmt.Print(i,"            \r")
		
        line, err := r.ReadBytes('\n')
		if len(line) == 0 {
			break
		}

		adv, word, err := bufio.ScanWords(line, true)
		if string(word) != "MsgHex:" {
			continue
		}
		line = line[adv:]

		adv, data, err := bufio.ScanWords(line, true)
		if err != nil {
			fmt.Println(err)
			return
		}

		binary, err := hex.DecodeString(string(data))
		if err != nil {
			fmt.Println(err)
			return
		}

		msg, err := messages.UnmarshalMessage(binary)
		if err != nil {
			fmt.Println(err)
			return
		}

		s.InMsgQueue() <- msg
        i++
        if len(s.InMsgQueue())>200 {
            for len(s.InMsgQueue())>50 {
                time.Sleep(time.Millisecond*10)
            }       
            time.Sleep(time.Millisecond*100)
        }
        time.Sleep(time.Millisecond*50)
	}

}
