package divefromfile

import (
	"bufio"
	"fmt"
	"os"
	"io/ioutil"
	"strings"

	"github.com/FactomProject/electiontesting/ET2/dive"
	"github.com/FactomProject/electiontesting/controller"
)

func DiveFromFile(name string) {
	con := controller.NewControllerInterpreter(1, 1)
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	con.InitInterpreter()
	con.Interpret(strings.NewReader(string(data)))

	dive.Dive(con.BufferedMessages, con.Elections, 0, 10, []*controller.DirectedMsg{})
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
