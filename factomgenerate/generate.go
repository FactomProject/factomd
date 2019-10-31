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
	"regexp"
	"strings"
	"text/template"
)

func main() {
	var out bytes.Buffer

	// load the templates for files wrappers
	templates := template.Must(template.ParseGlob("./factomgenerate/*.tmpl"))

	// load the templates for go code
	// these templates use 'r["' and '"]' as the delimiter to make the template gofmt compatible
	goFiles, err := filepath.Glob("./factomgenerate/*_template.go")
	die(err)
	templates = template.Must(templates.Delims("ᐸ", "ᐳ").ParseFiles(goFiles...))

	// Find all requests in the form: //FactomGenerate [<key:value>]...
	cmd := exec.Command("/bin/bash", "-c", "find .. -name \\*.go | xargs grep -Eh \"^//FactomGenerate\" || true")
	//cmd := exec.Command("pwd")
	cmd.Stdout = &out
	err = cmd.Run()

	fmt.Printf("CMD: %v", out.String())
	if err != nil {
		fmt.Printf("ERR: %v", err)
		log.Fatal(err)
	}

	factomgeneraterequests := strings.Split(out.String(), "\n")

	RunTemplates(err, templates, factomgeneraterequests)

	// Find all requests in the form: Publish_<type> or Subscribe_<type>
	// handle the pub/sun requests
	cmdline := []string{"/bin/bash", "-c", "find .. -name \\*.go | xargs grep -Eh \"= *Publish_|= *Subscribe_\" || true"}
	fmt.Println(cmdline)
	cmd = exec.Command(cmdline[0], cmdline[1:]...)
	//cmd := exec.Command("pwd")

	cmd.Stdout = &out
	err = cmd.Run()
	fmt.Printf("CMD: %v", out.String())
	if err != nil {
		fmt.Printf("ERR: %v", err)
		log.Fatal(err)
	}

	pubsubrequests := []string{}
	re := regexp.MustCompile("Publish_[^ (]+|Subscribe_[^ (]+")
	// For each template request, split out the key value pairs ...
	for _, m := range strings.Split(out.String(), "\n") {
		matches := re.FindStringSubmatch(m)

		for _, x := range matches {
			parts := strings.SplitN(x, "_", 2)
			PorS, Type := parts[0], parts[1]
			fmt.Println(PorS, Type)
			pubsubrequests = append(pubsubrequests, fmt.Sprintf("//FactomGenerate template %s type %s", PorS, Type))
		}
	}
	RunTemplates(err, templates, pubsubrequests)
	fmt.Println("done")
}

// Create a file and run all the request for a template outputing the template results into the file
func RunTemplates(err error, templates *template.Template, requests []string) {

	fmt.Print("RunTemplates()", requests)

	// place to keep all the files
	files := make(map[string]*os.File)
	// need to parse all the instances before executing any so we can merge the imports
	instances := make(map[string][]map[string]interface{}) // map a templatename to a slice of requests

	// For each template request, split out the key value pairs ...
	for _, m := range requests {
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
		templatename = strings.ToLower(templatename)
		filename := "./generated/" + templatename + ".go"

		// make a map of the required importsMap, this eliminates duplication...
		var importsMap map[string]string = make(map[string]string)

		// collect the imports required by the template
		var out bytes.Buffer
		s := templatename + "-imports"
		templateimports := ""
		t := templates.Lookup(s)
		if t != nil {
			die(t.Execute(&out, []interface{}{}))
			templateimports = out.String()
		} else {
			fmt.Println("No template", s)
		}

		re := regexp.MustCompile("\".*?\"") // regex to extract the quoted strings from the imports statment
		// Add the quoted strings from template imports to the imports list
		for _, name := range re.FindStringSubmatch(templateimports) {
			importsMap[name] = "" // Only the key is used not the value
		}

		// collect the imports from the requests
		// import are either a string or a []string
		for _, details := range requests {
			if value, ok := details["import"]; ok {
				name, ok := value.(string)
				if ok {
					importsMap[name] = "" // don't use the value just the name
				} else {
					for _, name := range value.([]string) {
						if name != "" {
							importsMap[name] = "" // Only the key is used not the value
						}
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
		details["templatename"] = templatename
		details["imports"] = importsMap
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
				die(templates.ExecuteTemplate(f, templatename, details))
			}
		}
		// After we are done add the filetail to the file and close it.
		die(templates.ExecuteTemplate(f, "filetail", details))
		f.Close()
		runcmd("gofmt -w " + filename)
		runcmd("goimports -w " + filename)
		catfile(filename)
	}
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
