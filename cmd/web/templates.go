package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/neogan74/snip/internal/models"
	"github.com/neogan74/snip/ui"
)

type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func HumanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"HumanDate": HumanDate,
}

// newTemplateCache creates a cache of parsed HTML templates for the application.
// It scans the embedded filesystem for all page templates matching "html/pages/*.tmpl.html",
// and for each page, it parses the base template, navigation partial, and the page itself
// into a single *template.Template instance. The resulting map uses the page's base filename
// as the key. Returns the cache map or an error if template parsing fails.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		files := []string{
			"html/base.tmpl.html",
			"html/partials/nav.tmpl.html",
			page,
		}
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, files...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
