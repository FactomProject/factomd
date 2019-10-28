//+build ignore

//ᐸ/*
//This looks syntatically off because it is a template used to generate go code. In order to make the template be
//gofmt able the parse delimiters are set to 'ᐸ'  and ' ᐳ' so ᐸ.typename ᐳ will be replaced by the typename
//from the //FactomGenerate command
//*/ᐳ

//ᐸif false  ᐳ
package Dummy // this is only here to make gofmt happy and is never in the generated code
//ᐸend ᐳ

//ᐸdefine "threadsafemap" ᐳ
// Start threadsafemap generated go code

import (
	"github.com/FactomProject/factomd/common"
)


type ᐸ.typename ᐳ struct {
	sync.Mutex
//	common.Name
	internalMap map[ᐸ.indextype ᐳ] ᐸ.valuetype ᐳ
}

func (q *ᐸ.typename ᐳ) Init(parent common.NamedObject, name string, size int) *ᐸ.typename ᐳ {
//	q.Name.Init(parent, name)
	q.internalMap = make(map[ᐸ.indextype ᐳ] ᐸ.valuetype ᐳ, size)
	return q
}

func (q *ᐸ.typename ᐳ) Put(index  ᐸ.indextype ᐳ, value ᐸ.valuetype ᐳ) {
	q.Lock()
	q.internalMap[index]= value
	q.Unlock()
}

func (q *ᐸ.typename ᐳ) Get(index  ᐸ.indextype ᐳ) ᐸ.valuetype ᐳ {
	q.Lock()
	defer q.Unlock()
	return q.internalMap[index]
}


// End threadsafemap generated go code

//ᐸend ᐳ
