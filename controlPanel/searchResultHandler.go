package controlPanel

import (
	"fmt"
	"net/http"
)

type Content struct {
	Hash string
}

func handleSearchResult(content *SearchedStruct, w http.ResponseWriter) {
	templates.ParseGlob(TEMPLATE_PATH + "searchresults/*.html")
	templates.ParseFiles(TEMPLATE_PATH + "searchresults/type/" + content.Type + ".html")
	fmt.Println(content)
	err := templates.ExecuteTemplate(w, "searchResult", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
