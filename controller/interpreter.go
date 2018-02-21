package controller

import (
	"fmt"

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
	c.AddPrim(conprimitives, ".c", c.PrimPrintState, executable)
	c.AddPrim(conprimitives, ".ca", c.PrimPrintStateAll, executable)
	c.AddPrim(conprimitives, ".r", c.PrimPrintRoutingInfo, executable)
	c.AddPrim(conprimitives, ".rp", c.PrimToggleRouterPrinting, executable)

	//	Message routing
	c.AddPrim(conprimitives, "<-v", c.PrimRouteVolunteerMessage, executable)
	c.AddPrim(conprimitives, "<-o", c.PrimRouteVoteMessage, executable)
	c.AddPrim(conprimitives, "<-l", c.PrimRouteLevelMessage, executable)
	c.AddPrim(conprimitives, "<r>", c.PrimToggleRouting, executable)
	c.AddPrim(conprimitives, "s1", c.PrimRouteStep, executable)
	c.AddPrim(conprimitives, "s", c.PrimRouteStepN, executable)

	//  Pull Scenarios
	c.AddPrim(conprimitives, "runscene", c.RunScenario, executable)

	//return p
}

func (c *Controller) RunScenario() {
	scen := c.PopString()
	scene, ok := Scenarios[scen]
	if !ok {
		fmt.Printf("No scenario %s\n", scen)
		return
	}

	c.InterpretLine(scene)
}

func (c *Controller) PrimRouteStep() {
	c.Router.Step()
}

func (c *Controller) PrimRouteStepN() {
	c.Router.StepN(c.PopInt())
}

func (c *Controller) PrimPrintStateAll() {
	fmt.Println(c.ElectionStatus(-1))
	for i := 0; i < len(c.feds); i++ {
		fmt.Println(c.ElectionStatus(i))
	}
}

func (c *Controller) PrimPrintRoutingInfo() {
	fmt.Println(c.Router.Status())
}

func (c *Controller) PrimPrintState() {
	fmt.Println(c.ElectionStatus(c.PopInt()))
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
