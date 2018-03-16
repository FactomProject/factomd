package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	//"github.com/FactomProject/factomd/electionsCore/interpreter/interpreter"
	"github.com/FactomProject/factomd/electionsCore/controller"
)

import "flag"

type TopLevelInterpreter struct {
	*controller.Controller
}

func main() {
	c := controller.NewControllerInterpreter(3, 5)
	//t := new(TopLevelInterpreter)
	//t.Controller = c

	var ()
	flag.Parse()

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

func Shell(i *controller.ControllerInterpreter) {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input := grabInput(in)
		i.Interpret(strings.NewReader(input))

	}
}
