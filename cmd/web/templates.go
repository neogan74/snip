package main

import (
	"github.com/neogan74/snip/internal/models"
	"github.com/neogan74/snip/ui"
	"html/template"
	"io/fs"
	"path/filepath"
	"time"
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
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"HumanDate": HumanDate,
}

func newTempalteCache() (map[string]*template.Template, error) {
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
