package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/factoid/binary2fct/code"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			fmt.Println("Usage:   binary2fct 0xD40D49BA60ECA0B66AC73699749DDF4911F0E05F948159AE64989FCDE51EE1AE")
			fmt.Println("Returns: FA3aU8zXyDVv72no45Rf1itwqLb5Z3WWVbWVGhCEWjyGshtNBzRM")
			fmt.Println("\nNote that \"0x\" can be omitted")
		}
	}()
	if len(os.Args) != 2 {
		panic("Must have only 1 argument, a 32 byte hash as a string.")
	}
	s := os.Args[1]
	if len(s) == 66 {
		s = s[2:]
	}
	if len(s) != 64 {
		panic("Not 32 bytes")
	}

	fmt.Println(code.Doit(s))
}
