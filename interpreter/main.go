package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	//"github.com/FactomProject/electiontesting/interpreter/interpreter"
	"github.com/FactomProject/electiontesting/interpreter/primitives"
)

func main() {
	i := primitives.NewPrimitives()
	Shell(i)
}

func grabInput(in *bufio.Reader) string {
	input, err := in.ReadString('\n')
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}
	return strings.TrimRight(input, "\n")
}

func Shell(i *primitives.Primitives) {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input := grabInput(in)
		i.Interpret(strings.NewReader(input))

	}
}
