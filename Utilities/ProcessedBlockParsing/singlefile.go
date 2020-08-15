package main

import (
	"fmt"

	"bytes"
	"encoding/json"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/state"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

func OpenBlock(block *state.WholeBlock) {
	// the questions to ask
	var qs = []*survey.Question{
		{
			Name: "block",
			Prompt: &survey.Select{
				Message: "Choose a block:",
				Options: []string{"back", "Directory Block", "Admin Block", "Factoid Block", "Entry Credit Block"},
				Default: "Directory Block",
			},
		},
	}

	var qs2 = []*survey.Question{
		{
			Name: "printtype",
			Prompt: &survey.Select{
				Message: "Choose a printing format:",
				Options: []string{"json", "string", "binary"},
				Default: "json",
			},
		},
	}

	// the answers will be written to this struct
	answers := struct {
		Block     string `survey:"block"` // or you can tag fields to match a specific name
		PrintType string `survery:"printtype"`
	}{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if answers.Block == "back" {
		return
	}

	// perform the questions
	err = survey.Ask(qs2, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	switch answers.Block {
	case "Directory Block":
		Print(block.DBlock, answers.PrintType)
	case "Admin Block":
		Print(block.ABlock, answers.PrintType)
	case "Factoid Block":
		Print(block.FBlock, answers.PrintType)
	case "Entry Credit Block":
		Print(block.ECBlock, answers.PrintType)
	case "back":
		return
	}

	OpenBlock(block)
}

func Print(obj interfaces.BinaryMarshallable, format string) {
	switch format {
	case "json":
		PrintJson(obj)
	case "string":
		PrintString(obj)
	case "binary":
		PrintBinary(obj)
	}
}

func PrintBinary(obj interfaces.BinaryMarshallable) {
	data, err := obj.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
}

func PrintString(obj interface{}) {
	fmt.Println(obj)
}

func PrintJson(obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, data, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(buf.Bytes()))
}
