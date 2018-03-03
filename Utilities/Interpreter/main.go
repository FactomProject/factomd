package main

import "fmt"

const (
	Invalid = iota
	Integer
	Function
)

type Element interface {
	Type() int
	Execute(i *Interpret)
	SetExecutable(bool)
}

//===============================================Default stuff
type ElementDefault struct {
}

func (e *ElementDefault) Type() int { return Invalid }
func (e *ElementDefault) Execute(i *Interpret) {
	i.DataStack[i.Dstkptr] = e
	i.Dstkptr++
}
func (e *ElementDefault) SetExecutable(bool) {}

//=================================================Default function
type ElementFunction struct {
	ElementDefault
	Executable bool
}

func (e *ElementFunction) Type() int { return Function }
func (e *ElementFunction) Execute(i *Interpret) {
	i.DataStack[i.Dstkptr] = e
	i.Dstkptr++
}
func (e *ElementFunction) SetExecutable(bool) {}

//================================================Integer

type El_Int int64

func (e El_Int) Type() int            { return Integer }
func (e El_Int) Execute(i *Interpret) { i.DataStack[i.Dstkptr] = e }
func (e El_Int) SetExecutable(bool)   {}

//=================================================Plus

type Interpret struct {
	Dstkptr   int
	DataStack [1000]Element
}

func main() {
	v := El_Int(5)
	fmt.Print("Try ", v, v.Type())

	recurse(3, 3, 40)

}

var breath = 0

func dive(msgs []int, leaders int, depth int, limit int) {
	depth++
	if depth > limit {
		fmt.Println("Breath ", breath)
		breath++
		return
	}

	for d, v := range msgs {
		msgs2 := append(msgs[0:d], msgs[d+1:]...)
		ml2 := len(msgs2)
		for i := 0; i < leaders-1; i++ {
			msgs2 = append(msgs2, v)
			dive(msgs2, leaders, depth, limit)
			msgs2 = msgs2[:ml2]
		}
	}
}

func recurse(msg int, leaders int, limit int) {
	var msgs []int
	for i := 0; i < msg; i++ {
		for j := 0; j < leaders-1; j++ {
			msgs = append(msgs, i*1024+j)
		}
	}
	dive(msgs, leaders, 0, limit)
}
