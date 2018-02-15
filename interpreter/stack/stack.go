package stack

import . "github.com/FactomProject/electiontesting/interpreter/common"

type Stack struct {
	ptr  int
	data []interface{}
}

func NewStack() Stack {
	s := Stack{0, make([]interface{}, 0)}
	return s
}

//predecrement stackpointer
func (s *Stack) pdec() int {
	s.ptr--
	return s.ptr
}

//postincrement stackpointer
func (s *Stack) pinc() int {
	p := s.ptr
	s.ptr++
	return p
}

func (s *Stack) Push(values ...interface{}) {
	if len(s.data) < s.ptr+len(values) {
		if s.ptr+len(values) > 100 {
			panic("stack overflow")
		}
		roomToGrow := make([]interface{}, len(values)+10)
		s.data = append(s.data, roomToGrow...)
	}
	for _, x := range values {
		s.data[s.pinc()] = x
	}
}

func (s *Stack) Pop() interface{}  { return s.data[s.pdec()] }          // Todo: underflow ?
func (s *Stack) Peek() interface{} { return s.data[s.ptr-1] }           // Todo: underflow ?
func (s *Stack) PopBool() bool     { return s.data[s.pdec()].(bool) }   // Todo: underflow ?
func (s *Stack) PopInt() int       { return s.data[s.pdec()].(int) }    // Todo: underflow ?
func (s *Stack) PopString() string { return s.data[s.pdec()].(string) } // Todo: underflow ?
func (s *Stack) PopArray() Array   { return s.data[s.pdec()].(Array) }  // Todo: underflow ?
