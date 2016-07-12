package controlPanel

import (
	"fmt"
	"net/http"
	"text/template"
)

func handleSearchResult(content *SearchedStruct, w http.ResponseWriter) {
	fmt.Print("Content: ")
	fmt.Println(content.Content)
	fmt.Println("type: " + content.Type)
	funcMap := template.FuncMap{
		// Now unicode compliant
		"truncate": func(s string) string {
			str := s
			fmt.Println(s)
			ret := ""
			if len(s) > 100 {
				for len(str) > 100 {
					ret = ret + str[:101] + "\n"
					str = str[100:]
				}
			}
			ret = ret + str[:]
			return ret
		},
	}
	templates.Funcs(funcMap)
	templates.ParseFiles(TEMPLATE_PATH + "searchresults/type/" + content.Type + ".html")
	templates.ParseGlob(TEMPLATE_PATH + "searchresults/*.html")
	err := templates.ExecuteTemplate(w, content.Type, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
