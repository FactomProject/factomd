package files

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PaulSnow/factom2d/controlPanel/files/statics"
	"github.com/PaulSnow/factom2d/controlPanel/files/templates"
)

// This is the general handler. The files are split into two categories to speed
// up globbing. Templates are usually parsed with a '*', and had to cycle through all
// static files as well. This handler will decide which category the user is looking for
// and only search within that category. Now can add more static files without affecting
// performance.

type staticFilesFile struct {
	data  string
	mime  string
	mtime time.Time
	size  int
}

var TemplatesServer http.Handler = templates.Server
var StaticServer http.Handler = statics.Server

func Open(name string) (io.ReadCloser, error) {
	if strings.Contains(name, "templates/") {
		return templates.Open(name[10:])
	} else {
		return statics.Open(name)
	}
}

func ModTime(file string) (t time.Time) {
	if strings.Contains(file, "templates/") {
		return templates.ModTime(file[10:])
	} else {
		return statics.ModTime(file)
	}
}

func Hash(file string) (s string) {
	if strings.Contains(file, "templates/") {
		return templates.Hash(file[10:])
	} else {
		return statics.Hash(file)
	}
}

func OpenGlob(name string) ([]io.ReadCloser, error) {
	if strings.Contains(name, "templates/") {
		return templates.OpenGlob(name[10:])
	} else {
		return statics.OpenGlob(name)
	}
}
