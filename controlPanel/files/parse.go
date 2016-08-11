package files

import (
	"io"
	"io/ioutil"
	"text/template"

	"fmt"
)

// Custom parsing of static files. Makes it easier to use with templates

func CustomParseGlob(temp *template.Template, file string) *template.Template {
	var err error
	readers, err := OpenGlob(file)
	if err != nil {
		return temp
	}

	for _, reader := range readers {
		if reader != nil {
			temp, err = parseData(temp, reader)
		}
	}
	return temp
}

func CustomParseFile(temp *template.Template, file string) (*template.Template, error) {
	reader, err := Open(file)
	if err != nil {
		return temp, err
	}
	temp, err = parseData(temp, reader)
	if err != nil {
		return temp, err
	}
	return temp, nil
}

func parseData(temp *template.Template, reader io.ReadCloser) (*template.Template, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return temp, err
	}
	reader.Close()

	if len(data) > 0 {
		if temp == nil {
			x := template.New("new")
			x.Parse(string(data))
			return x, nil
		} else {
			temp.Parse(string(data))
			return temp, nil
		}
	}
	return temp, fmt.Errorf("no data")
}
