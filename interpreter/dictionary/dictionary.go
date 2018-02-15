package dictionary

type Dictionary map[string]interface{}

func NewDictionary() Dictionary {
	return make(map[string]interface{}, 0)
}

func (d Dictionary) Add(s string, e interface{}) { d[s] = e }
