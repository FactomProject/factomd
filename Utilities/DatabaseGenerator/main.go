package main

import "github.com/FactomProject/factomd/Utilities/DatabaseGenerator/blockgen"

func main() {
	gen, err := blockgen.NewDBGenerator(blockgen.NewDefaultDBGeneratorConfig())
	if err != nil {
		panic(err)
	}

	err = gen.CreateBlocks(30)
	if err != nil {
		panic(err)
	}
}
