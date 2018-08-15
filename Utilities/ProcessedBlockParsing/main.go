package main

import (
	"flag"

	"path/filepath"

	"fmt"

	"os"

	"github.com/FactomProject/factomd/state"
)

func main() {
	var (
		file      = flag.String("f", "", "File to open")
		directory = flag.String("d", "", "Directory containing multiple directories")
	)

	flag.Parse()

	if *file != "" {
		OpenSingleFile(*file)
		return
	}

	if *directory != "" {
		OpenDirectory(*directory)
	}
}

func OpenSingleFile(filename string) {
	wb, err := state.ReadDBStateFromDebugFile(filename)
	if err != nil {
		panic(err)
	}
	OpenBlock(wb)
}

var directories = []string{}

func OpenDirectory(directory string) {
	// This will find all directories of interest
	filepath.Walk(directory, FindDirectory)
	fmt.Printf("Found %d directories containing DBState files\n", len(directories))
	for _, s := range directories {
		fmt.Println(" -", s)
	}

	for {
		var qs = []*survey.Question{
			{
				Name: "option",
				Prompt: &survey.Select{
					Message: "Choose an option. Compare will compare all dbstates at a given height. Select will select a single dbstate file",
					Options: []string{"compare", "select", "exit"},
					Default: "compare",
				},
			},
		}

		// the answers will be written to this struct
		answers := struct {
			Option string `survey:"option"` // or you can tag fields to match a specific name
		}{}

		// perform the questions
		err := survey.Ask(qs, &answers)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		switch answers.Option {
		case "exit":
			return
		case "compare":
			CompareDirectory()
		case "select":
			SelectDirectory()
		}
	}
}

// FindDirectory will find all directories containing *.block files at it's first depth
func FindDirectory(path string, info os.FileInfo, err error) error {
	// We don't care about files
	if !info.IsDir() {
		return nil
	}

	folders, err := filepath.Glob(filepath.Join(path, "*.block"))
	if len(folders) != 0 {
		// This directory might be useless, but it might contain a directory that has value
		directories = append(directories, path)
	}

	return nil
}
