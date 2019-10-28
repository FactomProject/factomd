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
	goTemplates := template.Must(template.New("").Delims("ᐸ", "ᐳ").ParseFiles(goFiles...))

	// place to keep all the files
	files := make(map[string]*os.File)
	// need to parse all the instances before executing any so we can merge the imports
	instances := make(map[string][]map[string]interface{}) // map a templatename to a slice of requests

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
			details := make(map[string]interface{})
			// Add timestamp to Details, it's used in the fileheader template
			fmt.Printf("t: %s\n", m)
			s := strings.Split(m, " ")
			if len(s)%2 == 0 {
				panic("odd number of strings in key:value list")
			}
			for i := 1; i < len(s); i += 2 { // skip //FactomGenerate
				key := s[i]
				value := s[i+1]
				if strings.Contains(value, ",") {
					details[key] = strings.Split(value, ",")
				} else {
					details[key] = value
				}
			}
			// Add this request to the slice of request for this template
			instances[details["template"].(string)] = append(instances[details["template"].(string)], details)
		}
	}

	// loop thru the templates and execute the requests for that template
	for templatename, requests := range instances {
		filename := "./generated/" + templatename + ".go"

		// make a map of the required importsMap, this eliminates duplication...
		// import are either a string or a []string
		var importsMap map[string]string = make(map[string]string)
		for _, details := range requests {
			if value, ok := details["import"]; ok {
				name, ok := value.(string)
				if ok {
					importsMap[name] = "" // don't use the value just the name
				} else {
					for _, name := range value.([]string) {
						importsMap[name] = "" // don't use the value just the name
					}
				}
				delete(details, "import")
			}
		}
		//// convert the map to a slice
		//var imports []string
		//for name, _ := range importsMap {
		//	imports = append(imports, name)
		//}

		// make the file header
		details := make(map[string]interface{})
		details["timestamp"] = now
		details["imports"] = importsMap
		details["template_imports"] = "." + templatename + ".imports"
		details["test"] = strings.Contains(templatename, "_test")
		fmt.Println("Creating", filename, "with", details)
		f, err := os.Create(filename)
		die(err)
		files[filename] = f
		// Add the file header to the file
		die(templates.ExecuteTemplate(f, "fileheader", details))

		// Loop thru the requests
		for _, details := range requests {
			if len(details) > 1 { // skip empty details e.g. import only requests have just the templatename at this point
				fmt.Println("processing", templatename, "request ", filename, "with", details)
				// Expand this instance of the template
				die(goTemplates.ExecuteTemplate(f, templatename, details))
			}
		}
		// After we are done add the filetail to the file and close it.
		die(templates.ExecuteTemplate(f, "filetail", details))
		f.Close()
		catfile(filename)

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
	cmd := exec.Command("/bin/bash", "-c", commandline+" 2>&1")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out.String())
}
