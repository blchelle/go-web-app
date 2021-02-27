package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// Page holds the title and body of a web page
type Page struct {
	Title string
	Body []byte
}

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
func viewHandler(w http.ResponseWriter, r *http.Request) {
	// Extracts the page title from the path and trims the '/view/' prefix
	title := r.URL.Path[len("/view/"):]

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
func editHandler(w http.ResponseWriter, r *http.Request) {
	// Extracts the page title from the path and trims the '/view/' prefix
	title := r.URL.Path[len("/edit/"):]

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
func saveHandler(w http.ResponseWriter, r *http.Request) {
	// Gets the body from the path and the body from the form submission
	title := r.URL.Path[len("/save/"):]
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
	t, err := template.ParseFiles(pageName + ".html")

	// Catches any potential errors that occurred while parsing the template
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, p)

	// Catches any potential errors that occurred executing the
	// page into the template
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Sets up handlers for the view, edit and save routes
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	// Spins up the server and listens on port 8000
	log.Fatal(http.ListenAndServe(":8000", nil))
}
