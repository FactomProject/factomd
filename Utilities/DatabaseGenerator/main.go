package main

import "github.com/FactomProject/factomd/Utilities/DatabaseGenerator/blockgen"

func main() {
	blockgen.NewDBGenerator(blockgen.NewDefaultDBGeneratorConfig())
}
