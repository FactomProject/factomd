// The following directive is necessary to make the package coherent:

// //+build ignore

// This program generates generated/*.  It can be invoked by running go generate
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// The FactomGenerate templates use Greek Capitol syllabary characters "Ͼ" U+03FE, "Ͽ" U+03FF as the delimiters. This
// is done so the template can be valid go code and goimports and gofmt will work correctly on the code and it can be
// tested in unmodified form. There are a few accommodated to facilitate this.
// U+03FE	GREEK CAPITAL DOTTED LUNATE SIGMA SYMBOL			Ͼ
// U+03FF	GREEK CAPITAL REVERSED DOTTED LUNATE SIGMA SYMBOL	Ͽ
// 1 - the "Ͼ", "Ͽ" are replaced with the traditional {{ and }} delimiters prior to loading the templates. This avoid
// an issue with parsing the templates caused by "Ͼ" and "Ͽ" are valid character in a template variable name.
// 2 - go templates define a template names <templatname>-imports which list the packages use in the template body.
// this is merges with the imports required by the instances of the template use to build the imports statement for
// the generate file.
// 3 - the manipulated text for the template is written to a temporary file instead of loading it from the string to
// facilitate the template error reporting producing usable error messages
//

func main() {
	// handle //FactomGenerate ... requests
	templates := LoadTemplates()
	factomgeneraterequests := CollectFactomGenerateRequests()
	RunTemplates(templates, factomgeneraterequests)

	// handle the pub/sub requests
	pubsubrequests := append(CollectPublishRequests(), CollectSubscribeRequests()...)
	RunTemplates(templates, pubsubrequests)
	fmt.Println("done")
}

func LoadTemplates() *template.Template {
	// load the templates for files wrappers
	templates := template.New("FactomGenerate")
	cwd, _ := os.Getwd()
	fmt.Println(cwd)
	template.Must(templates.ParseGlob("./factomgenerate/templates/*.tmpl"))
	// load the templates for go code
	// these templates use "Ͼ", "Ͽ" as the delimiter to make the template gofmt compatible
	goTemplateFiles, err := filepath.Glob("./factomgenerate/templates/*/*_template*.go")
	die(err)
	for _, filename := range goTemplateFiles {
		reformattedFilename := ReformatTemplateFile(filename)
		template.Must(templates.ParseFiles(reformattedFilename))
		// os.Remove(reformattedFilename) // clean up
	}
	return templates
}

// Find all requests in the form: //FactomGenerate [<key:value>]... in all the go files
func CollectFactomGenerateRequests() []string {
	var out bytes.Buffer
	cmdline := []string{"/bin/bash", "-c", "find .. -name \\*.go | xargs grep -Eh \"^//FactomGenerate\" || true"}
	// the odd || true at the end avoid grep returning a bogus error code.
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = &out
	err := cmd.Run()
	die(err)
	factomgeneraterequests := strings.Split(out.String(), "\n")
	return factomgeneraterequests
}

// Find all requests in the form: Publish_<PublisherType>_<ValueType> in all the go files
func CollectPublishRequests() []string {
	var out bytes.Buffer
	regexString := "Publish_[^ (]+"
	cmdline := []string{"/bin/bash", "-c", "find .. -name \\*.go | xargs grep -Eh \"= *generated." + regexString + "\" || true"}
	// the odd || true at the end avoid grep returning a bogus error code.
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = &out
	err := cmd.Run()
	die(err)
	pubsubrequests := make(map[string]interface{})
	re := regexp.MustCompile(regexString)
	// For each template request, split out the key value pairs and deduplicate using a map...
	for _, m := range strings.Split(out.String(), "\n") {
		matches := re.FindStringSubmatch(m)
		for _, x := range matches {
			// Split the Publish_<type> or Subscribe_<type> into template and type
			parts := strings.SplitN(x, "_", 3)
			PublisherType, ValueType := parts[1], parts[2]
			// reformat these into the FactomGenerate form
			fg := fmt.Sprintf("//FactomGenerate template publish publishertype %s valuetype %s", PublisherType, ValueType)
			pubsubrequests[fg] = nil // Only the key matters
		}
	}

	// convert map to a slice of strings
	pubsubrequestsstrings := []string{}
	for k, _ := range pubsubrequests {
		fmt.Println(k)
		pubsubrequestsstrings = append(pubsubrequestsstrings, k)
	}
	return pubsubrequestsstrings
}

// Find all requests in the form: Subscribe_<SubscriberType>_<type> in all the go files
func CollectSubscribeRequests() []string {
	var out bytes.Buffer
	regex_string := "Subscribe_[^_]+_[^ (]+"
	cmdline := []string{"/bin/bash", "-c", "find .. -name \\*.go | xargs grep -Eoh \"= *generated." + regex_string + "\\(\" || true"}
	// the odd || true at the end avoid grep returning a bogus error code.
	fmt.Println(cmdline)
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = &out
	err := cmd.Run()
	die(err)
	requests := make(map[string]interface{})
	re := regexp.MustCompile(regex_string)
	// For each template request, split out the key value pairs and deduplicate using a map...
	for _, m := range strings.Split(out.String(), "\n") {
		matches := re.FindStringSubmatch(m)
		for _, x := range matches {
			// Split the Publish_<type> or Subscribe_<type> into template and type
			parts := strings.SplitN(x, "_", 3)
			subscriptionType, valueType := parts[0]+"_"+parts[1], parts[2]
			// reformat these into the FactomGenerate form
			fg := fmt.Sprintf("//FactomGenerate template %s valuetype %s", subscriptionType, valueType)
			requests[fg] = nil // Only the key matters
		}
	}

	// convert map to a slice of strings
	subscriptionRequests := []string{}
	for k, _ := range requests {
		fmt.Println(k)
		subscriptionRequests = append(subscriptionRequests, k)
	}
	return subscriptionRequests
}

// Change the delimiters in the templates containing go code so the file works as a template and as go code
func ReformatTemplateFile(filename string) string {
	filerc, err := os.Open(filename)
	die(err)
	defer filerc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(filerc)
	contents := buf.String()

	updatedContents := strings.Replace(contents, "Ͼ_", "{{.", -1)
	updatedContents = strings.Replace(updatedContents, "Ͼ", "{{", -1)
	updatedContents = strings.Replace(updatedContents, "Ͽ", "}}", -1)
	dir := "/tmp/FactomGenerate"
	die(os.MkdirAll(dir, 0755)) // create a temp directory
	tmpfile, err := ioutil.TempFile(dir, filepath.Base(filename))
	die(err)
	_, err = tmpfile.Write([]byte(updatedContents))
	die(err)
	die(tmpfile.Close())
	// fmt.Println(tmpfile.Name())
	// fmt.Print(updatedContents)
	return tmpfile.Name()
}

// Create a file and run all the request for a template outputting the template results into the file
func RunTemplates(templates *template.Template, requests []string) {

	fmt.Print("RunTemplates()", requests)

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
		ExpandRequest(templates, templatename, requests)
	}
}

func ExpandRequest(templates *template.Template, templateName string, requests []map[string]interface{}) {
	templateName = strings.ToLower(templateName)
	filename := "./generated/" + templateName + ".go"
	// place to keep all the files

	fmt.Printf("ExpandRequests(%s,...) in %s\n", templateName, filename)
	for _, r := range requests {
		fmt.Println(r)
	}
	fmt.Println()

	// make a map of the required importsMap, this eliminates duplication...
	var importsMap map[string]string = make(map[string]string)

	// collect the imports required by the template
	{
		var out bytes.Buffer
		var templateimports string
		templateImportsName := templateName + "-imports"
		t := templates.Lookup(templateImportsName)
		if t != nil {
			die(t.Execute(&out, []interface{}{}))
			templateimports = out.String()

			re := regexp.MustCompile(".*\".*?\"") // regex to extract the quoted strings from the imports statement
			// Add the quoted strings from template imports to the imports list
			imports := re.FindAllStringSubmatch(templateimports, -1)
			for _, l := range imports {
				if len(l) != 1 {
					panic(errors.New("nested quoted string in " + templateimports))
				}
				for _, name := range l {
					importsMap[name] = "" // Only the key is used not the value
				}
			}
		} else {
			fmt.Println("Missing template", templateImportsName)
		}
	}
	// collect the imports from the requests for this template
	// import are either a string or a []string
	for _, details := range requests {
		fmt.Println("request:", details)
		value, ok := details["import"]
		if ok && details["templateName"] == templateName {
			name, ok := value.(string)
			if ok {
				importsMap[name] = "" // Only the key is used not the value
			} else {
				for _, name := range value.([]string) {
					if name != "" {
						importsMap[name] = "" // Only the key is used not the value
					}
				}
			}
			delete(details, "import") // don't need this anymore
		}
	}

	// make the file header

	fileHeaderDetails := make(map[string]interface{})
	fileHeaderDetails["templateName"] = templateName
	fileHeaderDetails["imports"] = importsMap
	fileHeaderDetails["template_imports"] = "." + templateName + ".imports"
	fileHeaderDetails["test"] = strings.Contains(templateName, "_test")
	fmt.Println("Creating", filename, "with", fileHeaderDetails)
	f, err := os.Create(filename)
	die(err)
	// Add the file header to the file
	die(templates.ExecuteTemplate(f, "fileheader", fileHeaderDetails))

	// Loop thru the requests
	for _, details := range requests {
		if len(details) > 1 { // skip empty fileHeaderDetails e.g. import only requests have just the templateName at this point
			fmt.Println("processing", templateName, "request ", filename, "with", details)
			// Expand this instance of the template
			die(templates.ExecuteTemplate(f, templateName, details))
		}
	}
	// After we are done add the filetail to the file and close it.
	die(templates.ExecuteTemplate(f, "filetail", fileHeaderDetails))
	f.Close()

	runcmd("gofmt -s -w " + filename)
	runcmd("goimports -w " + filename)
	//catfile(filename)
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
