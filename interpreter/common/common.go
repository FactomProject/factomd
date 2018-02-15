package common

type Flags struct {
	Immediate bool
	Executable bool
}

type Func func()

type ImmediateFunc struct {
	Flags
	Func
}

type Mark struct{} // Used to mark a spot on the stack for executable arrays

type Array struct {
	Data []interface{}
	Flags
}
func(a *Array)len()int{return len(a.Data)}
func(a *Array)cap()int{return cap(a.Data)}

type Name struct {
	n string
	Flags
	d interface{}
}

