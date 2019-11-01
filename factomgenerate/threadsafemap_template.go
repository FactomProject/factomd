//+build ignore

//ᐸ/*
//This looks syntatically off because it is a template used to generate go code. In order to make the template be
//gofmt able the parse delimiters are set to 'ᐸ'  and 'ᐳ' so ᐸ_typenameᐳ will be replaced by the typename
//from the //FactomGenerate command
//*/ᐳ

package Dummy // this is only here to make gofmt happy and is never in the generated code
//ᐸdefine "accountedqueue-imports"ᐳ

import (
	"sync"

	"github.com/FactomProject/factomd/common"
)

//ᐸendᐳ

//ᐸdefine "threadsafemap"ᐳ
// Start threadsafemap generated go code

type ᐸ_typenameᐳ struct {
	sync.Mutex
	common.Name
	internalMap map[ᐸ_indextypeᐳ]ᐸ_valuetypeᐳ
}

func (q *ᐸ_typenameᐳ) Init(parent common.NamedObject, name string, size int) *ᐸ_typenameᐳ {
	q.Name.Init(parent, name)
	q.internalMap = make(map[ᐸ_indextypeᐳ]ᐸ_valuetypeᐳ, size)
	return q
}

func (q *ᐸ_typenameᐳ) Put(index ᐸ_indextypeᐳ, value ᐸ_valuetypeᐳ) {
	q.Lock()
	q.internalMap[index] = value
	q.Unlock()
}

func (q *ᐸ_typenameᐳ) Get(index ᐸ_indextypeᐳ) ᐸ_valuetypeᐳ {
	q.Lock()
	defer q.Unlock()
	return q.internalMap[index]
}

// End threadsafemap generated go code

//ᐸendᐳ
