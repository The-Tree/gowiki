//tutorial used: https://golang.org/doc/articles/wiki/
//completed all other tasks as well

package main

import (
	//"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var linkRegexp = regexp.MustCompile("\\[([a-zA-Z0-9]+)\\]")

type Page struct {
	Title string
	Body []byte
    DisplayBody template.HTML
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	p, err := loadPage("FrontPage")
	if err != nil {
		http.Redirect(w, r, "/edit/FrontPage", http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	escapedBody := []byte(template.HTMLEscapeString(string(p.Body)))

	p.DisplayBody = template.HTML(linkRegexp.ReplaceAllFunc(escapedBody, func(str []byte) []byte {
		matched := linkRegexp.FindStringSubmatch(string(str))
		out := []byte("<a href=\"/view/"+matched[1]+"\">"+matched[1]+"</a>")
		return out
    }))
	
    renderTemplate(w, "view", p)
}

/*func textToHTML(input []byte) []byte {
	bytes.Replace(input, []byte("["), []byte("<a href=\"/view/" + pageTitle + "\">"), -1)
	bytes.Replace(input, []byte("]"), []byte("</a>"), -1)
	
	return input
}*/

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")	
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler)) 
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}