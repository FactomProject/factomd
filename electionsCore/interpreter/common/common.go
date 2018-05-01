package common

type FlagsStruct struct {
	Traced     bool // Object is being traced in the debugger
	Immediate  bool // Objects executed even when compiling
	Executable bool // Object is executable
}

func (f FlagsStruct) SetExecutable(b bool) FlagsStruct { f.Executable = b; return f }
func (f FlagsStruct) SetImmediate(b bool) FlagsStruct  { f.Immediate = b; return f }
func (f FlagsStruct) SetTraced(b bool) FlagsStruct     { f.Traced = b; return f }

type HasFlags interface {
	GetFlags() FlagsStruct
	SetFlags(FlagsStruct)
}

type Func func()

type Mark struct{} // Used to mark a spot on the stack for executable arrays

type Array struct {
	Data []interface{}
	FlagsStruct
}

func NewArray() Array {
	var a Array
	a.Data = make([]interface{}, 0)
	return a
}
func (a *Array) len() int              { return len(a.Data) }
func (a *Array) cap() int              { return cap(a.Data) }
func (a Array) GetFlags() FlagsStruct  { return a.FlagsStruct }
func (a Array) SetFlags(f FlagsStruct) { a.FlagsStruct = f }
