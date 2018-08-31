package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"

	"fmt"
	"strconv"

	"github.com/FactomProject/factom"
)

func main() {
	var (
		file = flag.String("f", "out", "File to read from")
		host = flag.String("s", "localhost:8088", "Factomd location")
	)

	flag.Parse()

	factom.SetFactomdServer(*host)

	f, err := os.OpenFile(*file, os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(f)

	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}

		line_str := string(line)
		arr := strings.Split(line_str, ":")
		switch arr[0][:2] {
		case "FA":
			bal, err := factom.GetFactoidBalance(arr[0])
			if err != nil {
				panic(err)
			}

			arr[1] = strings.Replace(arr[1], " ", "", -1)
			value := strings.Split(arr[1], ".")

			top, _ := strconv.Atoi(value[0])
			bot, _ := strconv.Atoi(value[1])
			exp := int64(top)*1e8 + int64(bot)
			if bal != exp {
				fmt.Printf("Balance differs for %s . Exp %d, found %d\n", arr[0], exp, bal)
			}
		case "EC":
			bal, err := factom.GetECBalance(arr[0])
			if err != nil {
				panic(err)
			}

			arr[1] = strings.Replace(arr[1], " ", "", -1)
			exp, _ := strconv.Atoi(arr[1])
			if bal != int64(exp) {
				fmt.Printf("Balance differs for %s . Exp %d, found %d\n", arr[0], exp, bal)
			}
		}
	}
	fmt.Println("Complete")
}
