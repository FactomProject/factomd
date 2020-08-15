package files_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/controlPanel/files"
)

func TestOpen(t *testing.T) {
	_, err := Open("css/app.css")
	if err != nil {
		t.Errorf("Could not find and open a static file")
	}

	_, err = Open("templates/general/footer.html")
	if err != nil {
		t.Errorf("Could not find and open a template file")
	}
}

func TestModTime(t *testing.T) {
	mt := ModTime("css/app.css")
	if mt.IsZero() {
		t.Errorf("Could not find the static file")
	}

	mt = ModTime("templates/general/footer.html")
	if mt.IsZero() {
		t.Errorf("Could not find the template file")
	}
}

func TestHash(t *testing.T) {
	hash := Hash("css/app.css")
	if hash == "" {
		t.Errorf("Could not find the static file")
	}

	hash = Hash("templates/general/footer.html")
	if hash == "" {
		t.Errorf("Could not find the template file")
	}
}

func TestOpenGlob(t *testing.T) {
	_, err := OpenGlob("css/*")
	if err != nil {
		t.Errorf("Could not glob open stativs")
	}

	_, err = OpenGlob("templates/general/*")
	if err != nil {
		t.Errorf("Could not glob open templates")
	}
}
