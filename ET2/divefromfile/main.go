package main

import (
	"os"

	"io/ioutil"

	"fmt"

	"github.com/FactomProject/electiontesting/ET2/dive"
	"github.com/FactomProject/electiontesting/controller"
)

func main() {
	con := controller.NewController(1, 1)
	file, err := os.OpenFile("input.txt", os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	con.InterpretString(string(data))

	dive.Dive(con.BufferedMessages, con.Elections, 0, 0, []*controller.DirectedMsg{})
}
