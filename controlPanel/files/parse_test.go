// +build all

package files_test

import (
	"testing"
	"text/template"

	. "github.com/FactomProject/factomd/controlPanel/files"
)

func TestCustomParseGlob(t *testing.T) {
	temp := template.New("TestTemplate")
	temp = CustomParseGlob(temp, "templates/index/*")
	for _, temps := range temp.Templates() {
		if temps.Name() == "indexPage" {
			return // We pass, as it was parsed
		}
	}

	t.Errorf("Template was not parsed")
}

func TestCustomParseFiles(t *testing.T) {
	var err error
	temp := template.New("TestTemplate")
	temp, err = CustomParseFile(temp, "templates/index/index.html")
	if err != nil {
		t.Fail()
	}
	for _, temps := range temp.Templates() {
		if temps.Name() == "indexPage" {
			return // We pass, as it was parsed
		}
	}

	t.Errorf("Template was not parsed")
}
