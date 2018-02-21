package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	//"github.com/FactomProject/electiontesting/interpreter/interpreter"
	"github.com/FactomProject/electiontesting/controller"
)

import (
	"flag"
	"io/ioutil"
)

type TopLevelInterpreter struct {
	*controller.Controller
}

func main() {
	c := controller.NewController(3, 5)
	//t := new(TopLevelInterpreter)
	//t.Controller = c

	var (
		filename = flag.String("d", "in.txt", "File to read from")
		usefile  = flag.Bool("f", false, "Use in file as buffer to pull from")
	)
	flag.Parse()

	if *usefile {
		file, err := os.Open(*filename)
		if err != nil {
			panic(err)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}

		buffer := strings.Split(string(data), "\n")
		fmt.Printf("Found %d lines\n", len(buffer))
		c.LinesBuffered = buffer
	}

	Shell(c)
}

func grabInput(in *bufio.Reader) string {
	input, err := in.ReadString('\n')
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}
	return strings.TrimRight(input, "\n")
}

func Shell(i *controller.Controller) {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input := grabInput(in)
		i.Interpret(strings.NewReader(input))

	}
}
