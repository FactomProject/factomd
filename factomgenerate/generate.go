// The following directive is necessary to make the package coherent:

// //+build ignore

// This program generates contributors.go. It can be invoked by running
// go generate
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)


func main() {
	// load the templates for files wrappers
	templates := template.Must(template.ParseGlob("./factomgenerate/*.tmpl"))

	// load the templates for go code
	// these templates use 'r["' and '"]' as the delimiter to make the template gofmt compatible
	goFiles, err := filepath.Glob("./factomgenerate/*_template.go")
	die(err)
	goTemplates := template.Must(template.New("").Delims("r[\"", "\"]").ParseFiles(goFiles...))


	// place to keep all the files
	files := make(map[string]*os.File)
	// need to parse all the instances before executing any so we can merge the imports
	instances := make(map[string][]map[string]string) // map a templatename to a slice of requests

	// Find all comments in the form:
	//FactomGenerate [<key:value>]...
	cmd := exec.Command("/bin/bash", "-c", "find .. -name \\*.go | xargs grep -Eh \"^//FactomGenerate\"")
	//cmd := exec.Command("pwd")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	/* KLUDGE: exits w/ status 123
	fmt.Printf("CMD: %v", out.String())
	if err != nil {
		fmt.Printf("ERR: %v", err)
		log.Fatal(err)
	}
	 */

	now := time.Now().String()
	// For each template request, split out the key value pairs ...
	for _, m := range strings.Split(out.String(), "\n") {
		// ignore blank lines
		if len(m) > 0 {
			details := make(map[string]string)
			// Add timestamp to Details, it's used in the fileheader template
			fmt.Printf("t: %s\n", m)
			s := strings.Split(m, " ")
			if len(s)%2 == 0 {
				panic("odd number of strings in key:value list")
			}
			for i := 1; i < len(s); i += 2 { // skip //FactomGenerate
				details[s[i]] = s[i+1]
			}
			// Add this request to the slice of request for this template
			instances[details["template"]] = append(instances[details["template"]], details)
		}
	}

	// loop thru the templates and execute the requests for that template
	for templatename, requests := range instances {
		filename := "./generated/" + templatename + ".go"

		// make a list of the required imports... probably should sort/uniq it...
		var imports []string
		for _, details := range requests {
			if name, ok := details["import"]; ok {
				imports = append(imports, name)
			}
		}
		details := make(map[string]interface{})
		details["timestamp"] = now
		details["imports"] = imports
		// open the file if it is a new file
		var f *os.File
		var err error
		if files[filename] == nil {
			fmt.Println("Creating", filename, "with", details)
			f, err = os.Create(filename)
			die(err)
			files[filename] = f
			// After we are done add the filetail to the file and close it.
		} else {
			f = files[filename]
		}
		// Add the file header to the file
		die(templates.ExecuteTemplate(f, "fileheader", details))

		// Loop thru the requests
		for _, details := range requests {
			fmt.Println("processing", templatename, "request ", filename, "with", details)
			// Expand this instance of the template
			die(goTemplates.ExecuteTemplate(f, templatename, details))
		}

		die(templates.ExecuteTemplate(f, "filetail", details))
		f.Close()
		runcmd("gofmt -w " + filename)
		runcmd("goimports -w " + filename)
		catfile(filename)
	}

	fmt.Println("done")
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// for debug, cat the file
func catfile(filename string) {
	fmt.Println("File: ", filename)
	runcmd("cat " + filename)
}

// for debug, cat the file
func runcmd(commandline string) {
	fmt.Println("Run: ", commandline)
	cmd := exec.Command("/bin/bash", "-c", commandline)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out.String())
}
