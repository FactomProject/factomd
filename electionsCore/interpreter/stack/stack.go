package stack

import (
	. "github.com/PaulSnow/factom2d/electionsCore/interpreter/common"
	. "github.com/PaulSnow/factom2d/electionsCore/interpreter/names"
)

type Stack struct {
	Ptr  int
	Data []interface{}
}

func NewStack() Stack {
	s := Stack{0, make([]interface{}, 0)}
	return s
}

//predecrement stackpointer
func (s *Stack) pdec() int {
	s.Ptr--
	return s.Ptr
}

//postincrement stackpointer
func (s *Stack) pinc() int {
	p := s.Ptr
	s.Ptr++
	return p
}

func (s *Stack) Push(values ...interface{}) {
	if len(s.Data) < s.Ptr+len(values) {
		if s.Ptr+len(values) > 100 {
			panic("stack overflow")
		}
		roomToGrow := make([]interface{}, len(values)+10)
		s.Data = append(s.Data, roomToGrow...)
	}
	for _, x := range values {
		s.Data[s.pinc()] = x
	}
}

func (s *Stack) Clear()                  { s.Ptr = 0 }                        // clear datastack
func (s *Stack) Pop() interface{}        { return s.Data[s.pdec()] }          // Todo: underflow ?
func (s *Stack) Peek() interface{}       { return s.Data[s.Ptr-1] }           // Todo: underflow ?
func (s *Stack) PeekN(n int) interface{} { return s.Data[s.Ptr-1-n] }         // Todo: underflow ?
func (s *Stack) PopBool() bool           { return s.Data[s.pdec()].(bool) }   // Todo: underflow ?
func (s *Stack) PopInt() int             { return s.Data[s.pdec()].(int) }    // Todo: underflow ?
func (s *Stack) PopName() Name           { return s.Data[s.pdec()].(Name) }   // Todo: underflow ?
func (s *Stack) PopString() string       { return s.Data[s.pdec()].(string) } // Todo: underflow ?
func (s *Stack) PopArray() Array         { return s.Data[s.pdec()].(Array) }  // Todo: underflow ?

func (s *Stack) PopN(n int) { s.Ptr -= n } // Todo: underflow ?
