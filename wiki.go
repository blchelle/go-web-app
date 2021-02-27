package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// Page holds the title and body of a web page
type Page struct {
	Title string
	Body  []byte
}

// Parses the html files ahead of time
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

// Sets up a regular expression to compile path names later
var validPath = regexp.MustCompile("^/(edit|save|view)/([\\w]+)$")

// save gets a title and a body and creates a text file from that
func (p *Page) save() error {
	filename := p.Title + ".txt"

	// 0600 indicates that the file should be created with read-write
	// permissions for the current user only
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// loadPage searches for a specific file and returns the title and body of
// that page if it exists, otherwise it returns an error
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)

	// Checks if the read failed
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// viewHandler attempts to find a file with a name matching the path on the
// request. If it can find it, then it will return the info in html form.
// Otherwise it will redirect the user to the edit page for the same topic
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Attempts to load a page with the given title
	p, err := loadPage(title)

	// If no page exists, then the user will be redirected to the edit page
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	// Renders the html for the given page
	renderTemplate(w, "view", *p)
}

// editHandler displays a page for a user to edit the information for a given
// topic. Pressing save will create a '/send/' request, which is handled
// by sendHandler
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Attempts to load a page with the given title
	p, err := loadPage(title)

	// If the page doesn't exist then we render a page with the given title
	// and a blank body
	if err != nil {
		p = &Page{Title: title}
	}

	// Renders the html for the given page
	renderTemplate(w, "edit", *p)
}

// saveHandler attempts to create a page from a title specified in the path
// and a body from a form submission
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")

	// Creates a Page, converting the body to a byte array in the process
	p := &Page{Title: title, Body: []byte(body)}

	// Saves the page to a .txt file
	err := p.save()

	// Catches any errors that occurred while saving the new page
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirects the user the view route, which will display the newly
	// created page
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// renderTemplate is a helper function to render an html template from a
// specified file (pageName) and a specified page (p)
func renderTemplate(w http.ResponseWriter, pageName string, p Page) {
	// Executes on one of the cached templates
	err := templates.ExecuteTemplate(w, pageName+".html", p)

	// Catches any potential errors that occurred executing the
	// page into the template
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getTitle gets the title from the request URL path, it also throws an error
// if the path does not match the regular expression above
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)

	// The path does not match the pattern so the request is invalid
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}

	return m[2], nil // The title is the second subexpression
}

// makeHandler is a
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Attempts to match the path with the pattern
		m := validPath.FindStringSubmatch(r.URL.Path)

		// Invalid path, 404
		if m == nil {
			http.NotFound(w, r)
			return
		}

		// Execute the call back, passing the title in
		fn(w, r, m[2])
	}
}

func main() {
	// Sets up handlers for the view, edit and save routes
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	// Spins up the server and listens on port 8000
	log.Fatal(http.ListenAndServe(":8000", nil))
}
