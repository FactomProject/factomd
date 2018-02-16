package common

type FlagsStruct struct {
	Immediate bool
	Executable bool
}

type HasFlags interface {
	GetFlags() FlagsStruct
}

type Func func()

type ImmediateFunc struct {
	Flags FlagsStruct
	Func
}
func (a ImmediateFunc) GetFlags() FlagsStruct { return a.Flags }

type Mark struct{} // Used to mark a spot on the stack for executable arrays

type Array struct {
	Data []interface{}
	Flags FlagsStruct
}

func NewArray() Array {
	var a Array
	a.Data = make([]interface{}, 0)
	return a
}
func(a *Array)len()int{return len(a.Data)}
func(a *Array)cap()int{return cap(a.Data)}
func (a Array) GetFlags() FlagsStruct { return a.Flags }

type Name struct {
	n string
	flags FlagsStruct
	d interface{}
}

func (a Name) GetFlags() FlagsStruct { return a.flags }
