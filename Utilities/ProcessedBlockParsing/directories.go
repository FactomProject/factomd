package main

import (
	"fmt"

	"path/filepath"

	"github.com/FactomProject/factomd/state"
)

func SelectDirectory() {
	var qs = []*survey.Question{
		{
			Name: "select",
			Prompt: &survey.Select{
				Message: "Choose a directory:",
				Options: append([]string{"back"}, directories...),
				Default: directories[0],
			},
		},
	}

	// the answers will be written to this struct
	answers := struct {
		Select string `survey:"select"`
	}{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if answers.Select == "back" {
		return
	}

	files, err := filepath.Glob(filepath.Join(answers.Select, "*.block"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		qs = []*survey.Question{
			{
				Name: "select",
				Prompt: &survey.Select{
					Message: "Choose a file:",
					Options: append([]string{"back"}, files...),
					Default: files[0],
				},
			},
		}

		// perform the questions
		err = survey.Ask(qs, &answers)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if answers.Select == "back" {
			break
		}

		OpenSingleFile(answers.Select)
	}

	SelectDirectory()
}

type WBSummary struct {
	DBHash map[string][]string
	AHash  map[string][]string
	FHash  map[string][]string
	ECHash map[string][]string
}

func NewWBSummary() *WBSummary {
	wb := new(WBSummary)
	wb.DBHash = make(map[string][]string)
	wb.AHash = make(map[string][]string)
	wb.FHash = make(map[string][]string)
	wb.ECHash = make(map[string][]string)
	return wb
}

func CompareDirectory() {
	var qs3 = []*survey.Question{
		{
			Name: "select",
			Prompt: &survey.Select{
				Message: "Choose a block number to compare:",
				Options: append([]string{"back"}, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}...),
				Default: directories[0],
			},
		},
	}

	// the answers will be written to this struct
	answers := struct {
		Select string `survey:"select"`
	}{}

	// perform the questions
	err := survey.Ask(qs3, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if answers.Select == "back" {
		return
	}

	summaries := NewWBSummary()

	for _, d := range directories {
		filename := filepath.Join(d, fmt.Sprintf("processed_dbstate_%s.block", answers.Select))
		wb, err := state.ReadDBStateFromDebugFile(filename)
		if err != nil {
			fmt.Printf("%s could not be opened: %s\n", filename, err.Error())
			continue
		}

		summaries.DBHash[wb.DBlock.GetKeyMR().String()] = append(summaries.DBHash[wb.DBlock.GetKeyMR().String()], filename)
		k, _ := wb.ABlock.GetKeyMR()
		summaries.AHash[k.String()] = append(summaries.AHash[k.String()], filename)
		summaries.FHash[wb.FBlock.GetKeyMR().String()] = append(summaries.FHash[wb.FBlock.GetKeyMR().String()], filename)
		k, _ = wb.ECBlock.GetFullHash()
		summaries.ECHash[k.String()] = append(summaries.ECHash[k.String()], filename)
	}

	investigated := false
	if len(summaries.DBHash) > 1 {
		fmt.Println("Found more than 1 DBlockHash:")
		investigated = PrintDiff(summaries.DBHash, investigated)
	} else {
		fmt.Println("Directory Blocks match")
	}
	if len(summaries.AHash) > 1 {
		fmt.Println("Found more than 1 ABlockHash:")
		investigated = PrintDiff(summaries.AHash, investigated)
	} else {
		fmt.Println("Admin Blocks match")
	}

	if len(summaries.FHash) > 1 {
		fmt.Println("Found more than 1 FBlockHash:")
		investigated = PrintDiff(summaries.FHash, investigated)
	} else {
		fmt.Println("Factoid Blocks match")
	}

	if len(summaries.ECHash) > 1 {
		fmt.Println("Found more than 1 ECBlockHash:")
		investigated = PrintDiff(summaries.ECHash, investigated)
	} else {
		fmt.Println("Entry Credit Blocks match")
	}

	//CompareDirectory()
}

func PrintDiff(m map[string][]string, investigated bool) bool {
	investigate := len(m) > 1 && !investigated
	comparelist := []string{}

	for k, v := range m {
		fmt.Printf("  %d match %s\n", len(v), k)
		comparelist = append(comparelist, v[0])
		for _, d := range v {
			fmt.Printf("    - %s\n", d)
		}
	}

	if investigate {
		option := false
		prompt := &survey.Confirm{
			Message: "Do you want to investigate?",
		}
		survey.AskOne(prompt, &option, nil)

		if option {
			for {
				var qs = []*survey.Question{
					{
						Name: "select",
						Prompt: &survey.Select{
							Message: "Choose a file:",
							Options: append([]string{"back"}, comparelist...),
							Default: comparelist[0],
						},
					},
				}

				answers := struct {
					Select string `survey:"select"`
				}{}

				// perform the questions
				err := survey.Ask(qs, &answers)
				if err != nil {
					fmt.Println("Error in printdiff", err.Error())
					return true
				}

				if answers.Select == "back" {
					return true
				}

				OpenSingleFile(answers.Select)
			}
		}
		return true
	}
	return investigated
}
