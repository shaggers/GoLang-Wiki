package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	//"fmt"
)

//test file retrieval

type Page struct {
    Title string
    Body  []byte
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) { // here 6
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

// landing

func landingHandler(w http.ResponseWriter, r *http.Request, title string) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	var index []string
	for _, file := range files {
		index = append(index, file.Name())
	}
	var returnedFiles []string
	for _, item := range index {
		splitFile := strings.Split(item, "")
		length := len(splitFile)
		thirdLast := length - 3
		lastThreeChars := splitFile[thirdLast:length]
		joinedLastChars := strings.Join(lastThreeChars, "")
		if joinedLastChars == "txt" {
			fourLast := length - 4
			firstChars := splitFile[0:fourLast]
			joinedFirstChars := strings.Join(firstChars, "")
			returnedFiles = append(returnedFiles, joinedFirstChars)
		}
	}
	t, err := template.ParseFiles("landing.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w,returnedFiles)
}

//view

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

//new

func newHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	inputValue := r.Form["title"]
	title := strings.Join(inputValue, "")
	http.Redirect(w, r, "/edit/" + title, 301)
}

//edit

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

//save 

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

//Template Caching

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//validation

var validPath = regexp.MustCompile("^/(edit|save|view|landing)/([a-zA-Z0-9]+)$")

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

//saving pages to redirect on err

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/landing/", makeHandler(landingHandler))
	http.HandleFunc("/new", newHandler)

    log.Fatal(http.ListenAndServe(":8080", nil))
}

