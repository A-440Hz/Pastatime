package views

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"pastatime/internal/pastas"
)

var HomeTemplate *template.Template
var pps []*pastas.Pasta

func init() {
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		fmt.Println("views.go init err: ", err)
		os.Exit(1)
	}
	HomeTemplate = templates.Lookup("index.html")
}

func HomeFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		HomeTemplate.Execute(w, nil)
	}
}
