package exhaustiveTest

import (
	. "github.com/FactomProject/electiontesting/errorhandling"
	. "github.com/FactomProject/electiontesting/primitives"
	"testing"
	"fmt"
	"reflect"
	"github.com/davecgh/go-spew/spew"
)

func TestPermute(t *testing.T) {
	T = t // Set test for error handling
	var messages []*DummyMessage = []*DummyMessage{{"a"}, {"b"}, {"c"}}
	var results chan []*DummyMessage
	results = make(chan ([]*DummyMessage))

	go permute(messages, 0, len(messages)-1, results)

	answers := make(map[string]bool)
	limit := factorial(len(messages))
	for i := 0; i < limit; i++ {
		r := <-results
		f := fmt.Sprintf("%v\n", r)
		_, exists := answers[f]
		fmt.Println(f)
		if (exists) {
			T.Errorf("Duplicate permutation")
		}
	}
	if (len(results) != 0) {
		T.Errorf("Duplicate permutation 2")
	}
}

func TestClone2(t *testing.T) {
	var foo, foo2, foo3 *DummyElection
	foo = new(DummyElection)
	foo.Best = new(DummyMessage)
	foo.Best.S = "ACB"
	foo.Seen = make(map[*DummyMessage]bool)
    foo.Seen[foo.Best] = true

	foo2 = clone2(foo)
	foo3 = clone2(foo2)

	spew.Printf("foo  %#+v\n",foo)
	spew.Printf("foo2 %#+v\n",foo2)
	spew.Printf("foo3 %#+v\n",foo3)
	fmt.Printf("%+v %+v %+v %v %v\n", *foo, *foo2, *foo3, reflect.DeepEqual(foo,foo2), reflect.DeepEqual(foo,foo3))
}

func TestClone(t *testing.T) {
	var foo, foo2 []*DummyElection
	foo = make([]*DummyElection,2)
//	foo2 = make([]*DummyElection,2)
	foo[0] = new(DummyElection)
	foo[0].Best = new(DummyMessage)
	foo[0].Best.S = "ABC"
	foo[0].Seen = make(map[*DummyMessage]bool)
	foo[0].Seen[foo[0].Best] = true

	foo[1] = new(DummyElection)
	foo[1].Best = new(DummyMessage)
	foo[1].Best.S = "CBA"
	foo[1].Seen = make(map[*DummyMessage]bool)
	foo[1].Seen[foo[1].Best] = true



	foo2 = clone(foo)
	foo3 := clone(foo2)
	spew.Printf("%v\n",foo)
	spew.Printf("%v\n",foo2)
	spew.Printf("%v\n",foo3)
	fmt.Printf("%+v %+v %+v %v\n", foo, foo2, foo3, reflect.DeepEqual(foo,foo2), reflect.DeepEqual(foo,foo2))
}


func TestRunExhaustiveTest(t *testing.T) {

	var volunteers []Identity = []Identity{1, 2}
	var nodes [3] *DummyElection

	var messages []*DummyMessage;

	// make a message for every volunteer
	for i := 0; i < len(volunteers); i++ {
		var m *DummyMessage = &DummyMessage{string(65 + i)}
		messages = append(messages, m)
	}
	// make a message for every volunteer
	for i := 0; i < len(nodes); i++ {
		nodes[i] = new(DummyElection)
		nodes[i].Seen = make(map[*DummyMessage]bool, 0)
		nodes[i].Best = new(DummyMessage)
	}
	exhaustiveTest1(messages, nodes[:])
}
