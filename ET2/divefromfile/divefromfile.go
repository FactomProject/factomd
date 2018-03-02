package divefromfile

import (
	"bufio"
	"fmt"
	"os"
	"io/ioutil"
	"strings"

	//	"github.com/FactomProject/electiontesting/ET2/dive"
	"github.com/FactomProject/electiontesting/controller"
	"github.com/FactomProject/electiontesting/ET2/dive"
	. "github.com/FactomProject/electiontesting/ET2/directedmessage"
)

func DiveFromFile(name string, listen string, connect string, load string,  recursions int, randomFactor int, primeIdx int, global int) {
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

	dive.MirrorMap.Init("dive")

	if load != "" {
		dive.MirrorMap.Load(load)
	}
	if connect != "" {
		dive.MirrorMap.Connect(connect)
	}
	if listen != "" {
		dive.MirrorMap.Listen(listen)
	}

//	func Dive(mList []*mymsg, leaders []*election.Election, depth int, limit int, msgPath []*mymsg) (limitHit bool, leaf bool, seeSuccess bool) {
    dive.SetGlobals(recursions, randomFactor, primeIdx, global)
	dive.Dive(con.BufferedMessages, con.Elections, 0, 10, []*DirectedMessage{})
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
