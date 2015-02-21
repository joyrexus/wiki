// We're following the golang article "Writing Web Applications".
// See http://golang.org/doc/articles/wiki.

package main // wiki

import (
    "fmt"
    "regexp"
    "errors"
    "html/template"
    "io/ioutil"
    "net/http"
)

// regexp for validating url path
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

var templates = template.Must(template.ParseFiles("static/edit.html", 
     "static/view.html"))

// wiki page data struct
type Page struct {
    Title string
    Body []byte
    DisplayBody template.HTML
}

// save with read-write permissions for the current user only
func (p *Page) save() error {
    name := "pages/" + p.Title + ".txt"
    return ioutil.WriteFile(name, p.Body, 0600)
}

// open a particular wiki page
func open(title string) (*Page, error) {
    name := "pages/" + title + ".txt"
    body, err := ioutil.ReadFile(name)
    if err != nil {
        return nil, err
    }
    return &Page{title, body, ""}, nil
}

// render the named template with `p` as context
func render(w http.ResponseWriter, name string, p *Page) {
    err := templates.ExecuteTemplate(w, name, p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello wiki!")
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := open(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    p.DisplayBody = template.HTML(Filter(p.Body)) // filter/escape for display
    render(w, "view.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := open(title)
    if err != nil {
        p = &Page{Title: title}
    }
    render(w, "edit.html", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{title, []byte(body), ""}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil    // page title is second sub-expression of match
}

// Filter transforms [foo] into a link to /view/foo.
func Filter(input []byte) []byte {
    var pattern *regexp.Regexp
    pattern = regexp.MustCompile("\\[(\\w+)\\]")
    output := pattern.ReplaceAll(input, []byte("<a href=\"/view/$1\">$1</a>"))
    return output
}

// Handler type used as arg by makeHandler to build a normal http.HandlerFunc
type Handler func(http.ResponseWriter, *http.Request, string)

// makeHandler builds a handler acceptable by the http dispatcher from a
// Handler type.  It deals with url path validation and 
func makeHandler(fn Handler) http.HandlerFunc {
    handler := func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        title := m[2]   // title of page to load
        fn(w, r, title)
    }
    return handler
}


func main() {
    http.HandleFunc("/", indexHandler)
    // http.HandleFunc("/view/", viewHandler)
    // http.HandleFunc("/edit/", editHandler)
    // http.HandleFunc("/save/", saveHandler)
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.ListenAndServe(":8080", nil)
    
    /*
    p := Page{"Hi", []byte("Hello World!")}
    err := p.save()
    if err != nil {
        fmt.Println("Could not write file %q.txt: %v", p.Title, err)
    }
    pageName := "Hi"
    q, err := open(pageName)
    if err != nil {
        fmt.Println("Could not open file %q.txt: %v", pageName, err)
    }
    fmt.Printf("Loaded %q from `pages/%v.txt`", q.Body, q.Title)
    */
}
