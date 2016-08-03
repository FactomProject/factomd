package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	file, err := os.Open("exampleNet.txt")
	if err != nil {
		panic(fmt.Sprintf("File network.txt failed to open: %s", err.Error()))
	} else if file == nil {
		panic(fmt.Sprintf("File network.txt failed to open, and we got a file of <nil>"))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		fmt.Println(scanner.Text())
		var a, b int
		fmt.Sscanf(scanner.Text(), "%d %d", &a, &b)
		fmt.Println(a, b)
	}
}
