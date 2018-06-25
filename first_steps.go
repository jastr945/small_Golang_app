package main

import (
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "regexp"
)

type Page struct {
  Name string
  Body  []byte
}

func (p *Page) save() error {
  filename := p.Name + ".txt"
  return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(name string) (*Page, error) {
  filename := name + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  return &Page{Name: name, Body: body}, nil
}

// Template caching
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl+".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func viewHandler(w http.ResponseWriter, r *http.Request, name string) {
  p, err := loadPage(name)
  if err != nil {
    http.Redirect(w, r, "/edit/"+name, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, name string) {
  p, err := loadPage(name)
  if err != nil {
    p = &Page{Name: name}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, name string) {
  body := r.FormValue("body")
  p := &Page{Name: name, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/"+name, http.StatusFound)
}

// Validation with RegEx
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
	}
}

func main() {
  http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
