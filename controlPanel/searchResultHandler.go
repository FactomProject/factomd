package controlPanel

import (
	"fmt"
	"net/http"
)

type Content struct {
	Hash string
}

func handleSearchResult(content *SearchedStruct, w http.ResponseWriter) {
	fmt.Print("Content: ")
	fmt.Println(content.Content)
	fmt.Println("type: " + content.Type)
	templates.ParseFiles(TEMPLATE_PATH + "searchresults/type/" + content.Type + ".html")
	templates.ParseGlob(TEMPLATE_PATH + "searchresults/*.html")
	err := templates.ExecuteTemplate(w, content.Type, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
