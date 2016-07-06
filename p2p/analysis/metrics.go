// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strings"
	"log"
)

type MessageEvents struct {
    hash string
    minute int // Minute this message belongs to for time series
    time int64  // time message was originated
    reciepts map[string]time.Time  // map of recievers and time of reciept
    sends int // number of times this message has been sent
	target string // the intended node for the message
}

// Structure:
// main -
// 	- parse parameters get files to import
//	- importFile() 
//		- for each line, find message in hasmap
//		- if message is there, update stats
//		- if not, create it
//		- insert the new or updated message into hashmap
		- map[hash]MessageEvents - global map of all message metrics unordered
		- Keep track of the highest minute found
// - Analyze()
		- Create an array as large as the highest minute.
			- Array of map[string]MessageEvents - 
//		- For each record in the hashmap,
//			- Update global stats: 
				- Total messages sent
				- Total messages recieved
				- Total directed messages sent
				- Total directed messages sent
			- Add 
func main() {
	r := csv.NewReader(strings.NewReader(string(b)))
	s, _ := r.ReadAll()
	for i := 0; i < len(s); i++ {
		fmt.Println(s[i][0])
	}

	in := `first_name;last_name;username
"Rob";"Pike";rob
# lines beginning with a # character are ignored
Ken;Thompson;ken
"Robert";"Griesemer";"gri"
`
	r := csv.NewReader(strings.NewReader(in))
	r.Comma = ';'
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(records)
}

func importFile(file string) {
    	file, err := os.Open(file)
	    if nil != err {
		fmt.Printf("Metrics.importFile() File read error on file: %s, Error: %+v", file, err)
        panic(err)
		return
	}
	dec := json.NewDecoder(bufio.NewReader(file))

	b, _ := ioutil.ReadFile("1.csv")

}



i
	
}