package controller

import (
	"fmt"

	"strings"

	. "github.com/FactomProject/electiontesting/interpreter/common"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	priminterpreter "github.com/FactomProject/electiontesting/interpreter/primitives"
)

var executable FlagsStruct = FlagsStruct{Traced: false, Immediate: false, Executable: true}
var immediate FlagsStruct = FlagsStruct{Traced: false, Immediate: true, Executable: true}

func (c *Controller) InitInterpreter() {
	p := priminterpreter.NewPrimitives()
	c.Primitives = p

	// Controller dict
	conprimitives := NewDictionary()
	c.DictionaryPush(conprimitives)

	// Add custom functions
	//	Debug/Printing
	c.AddPrim(conprimitives, ".m", c.PrimPrintMessages, executable)
	c.AddPrim(conprimitives, ".ma", c.PrimPrintMessagesAll, executable)
	c.AddPrim(conprimitives, ".n", c.PrimPrintNodeStack, executable)
	c.AddPrim(conprimitives, ".na", c.PrimPrintNodeStackAll, executable)
	c.AddPrim(conprimitives, ".c", c.PrimPrintState, executable)
	c.AddPrim(conprimitives, ".ca", c.PrimPrintStateAll, executable)
	c.AddPrim(conprimitives, ".v", c.PrimPrintVoteState, executable)
	c.AddPrim(conprimitives, ".va", c.PrimPrintVoteStateAll, executable)
	c.AddPrim(conprimitives, ".r", c.PrimPrintRoutingInfo, executable)
	c.AddPrim(conprimitives, ".rp", c.PrimToggleRouterPrinting, executable)
	c.AddPrim(conprimitives, ".cp", c.PrimToggleControllerPrinting, executable)

	//	Message routing
	c.AddPrim(conprimitives, "<-v", c.PrimRouteVolunteerMessage, executable)
	c.AddPrim(conprimitives, "<-o", c.PrimRouteVoteMessage, executable)
	c.AddPrim(conprimitives, "<-l", c.PrimRouteLevelMessage, executable)
	c.AddPrim(conprimitives, "<r>", c.PrimToggleRouting, executable)
	c.AddPrim(conprimitives, "s1", c.PrimRouteStep, executable)
	c.AddPrim(conprimitives, "s", c.PrimRouteStepN, executable)

	//  Pull Scenarios
	c.AddPrim(conprimitives, "runscene", c.RunScenario, executable)

	// Crazy
	c.AddPrim(conprimitives, "reset", c.Reset, executable)
	c.AddPrim(conprimitives, "setcon", c.SetController, executable)

	//return p
}

func (c *Controller) SetController() {
	a := c.PopInt()
	f := c.PopInt()
	newc := NewController(f, a)

	*c = *newc
}

func (c *Controller) Reset() {
	newc := NewController(len(c.AuthSet.GetFeds()), len(c.AuthSet.GetAuds()))
	*c = *newc
}

func (c *Controller) RunScenario() {
	scen := c.PopString()
	scene, ok := Scenarios[scen]
	if !ok {
		fmt.Printf("No scenario %s\n", scen)
		return
	}

	c.Interpret(strings.NewReader(scene))
}

func (c *Controller) PrimRouteStep() {
	c.Router.Step()
}

func (c *Controller) PrimRouteStepN() {
	c.Router.StepN(c.PopInt())
}

func (c *Controller) PrimPrintRoutingInfo() {
	fmt.Println(c.Router.Status())
}

func (c *Controller) PrimPrintMessages() {
	f := c.PopInt()
	fmt.Println("Node", f)
	fmt.Println(c.Elections[f].PrintMessages())
}

func (c *Controller) PrimPrintMessagesAll() {
	for i := 0; i < len(c.feds); i++ {
		fmt.Println("Node", i)
		fmt.Println(c.Elections[i].PrintMessages())
	}
}

func (c *Controller) PrimPrintNodeStack() {
	fmt.Println(c.Router.NodeStack(c.PopInt()))
}

func (c *Controller) PrimPrintNodeStackAll() {
	for i := 0; i < len(c.feds); i++ {
		fmt.Println(c.Router.NodeStack(i))
	}
}

func (c *Controller) PrimPrintState() {
	fmt.Println(c.ElectionStatus(c.PopInt()))
}

func (c *Controller) PrimPrintStateAll() {
	fmt.Println(c.ElectionStatus(-1))
	for i := 0; i < len(c.feds); i++ {
		fmt.Println(c.ElectionStatus(i))
	}
}

func (c *Controller) PrimPrintVoteState() {
	fmt.Println(string(c.Elections[c.PopInt()].StateString()))
}

func (c *Controller) PrimPrintVoteStateAll() {
	for i := 0; i < len(c.feds); i++ {
		fmt.Printf("Node %d\n", i)
		fmt.Println(string(c.Elections[i].StateString()))
	}
}

// Vol  To
//  1 { 1 2 }<-v
//		Route vol 1 to 1, and 2
func (c *Controller) PrimRouteVolunteerMessage() {
	leaders := c.PrimSelectLeaders()
	vol := c.PopInt()

	c.RouteVolunteerMessage(vol, leaders)
}

//  From    Vote    To
// { 1 2 }   1    { 0 2 } <-o
//		Route vote 1 from (0, 2) to (1, 2)
func (c *Controller) PrimRouteVoteMessage() {
	to := c.PrimSelectLeaders()
	vote := c.PopInt()
	from := c.PrimSelectLeaders()
	c.RouteLeaderSetVoteMessage(from, vote, to)
}

//  From   Level    To
// { 1 2 }   1    { 0 2 } <-o
//		Route level 1 from (0, 2) to (1, 2)
func (c *Controller) PrimRouteLevelMessage() {
	to := c.PrimSelectLeaders()
	vote := c.PopInt()
	from := c.PrimSelectLeaders()
	c.RouteLeaderSetLevelMessage(from, vote, to)
}

func (c *Controller) PrimToggleControllerPrinting() {
	c.PrintingTrace = !c.PrintingTrace
	fmt.Printf("Printing: %t", c.PrintingTrace)
}
func (c *Controller) PrimToggleRouterPrinting() {
	c.Router.PrintMode(!c.Router.Printing)
	fmt.Printf("Printing: %t", c.Router.Printing)
}

func (c *Controller) PrimToggleRouting() {
	c.SendOutputsToRouter(!c.OutputsToRouter)
	fmt.Printf("Routing: %t", c.OutputsToRouter)
}

//
//func (c *Controller) PrimRouteMessage() {
//	c.RouteVolunteerMessage()
//}

func (c *Controller) PrimSelectLeaders() []int {
	// Select leaders groups leaders into array
	arr := c.PopArray()
	iarr := make([]int, len(arr.Data))
	for i, v := range arr.Data {
		iarr[i] = v.(int)
	}

	return iarr
}
