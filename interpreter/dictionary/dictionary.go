package dictionary

import . "github.com/FactomProject/electiontesting/interpreter/names"

type Dictionary map[Name]interface{}

func NewDictionary() Dictionary {
	return make(map[Name]interface{}, 0)
}

func (d Dictionary) Add(s Name, e interface{}) { d[s] = e }
