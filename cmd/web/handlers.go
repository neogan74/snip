package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/neogan74/snip/internal/models"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type SnippetCreateForm struct {
	Title      string
	Content    string
	Expires    int
	FieldErros map[string]string
}

func (app *App) home(w http.ResponseWriter, r *http.Request) {
	//if r.URL.Path != "/" {
	//	app.notFound(w)
	//	return
	//}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTempalteData(r)
	data.Snippets = snippets
	app.render(w, http.StatusOK, "home.tmpl.html", data)
}

func (app *App) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTempalteData(r)
	data.Snippet = snippet

	app.render(w, http.StatusOK, "view.tmpl.html", data)

}

func (app *App) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTempalteData(r)
	data.Form = &SnippetCreateForm{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

func (app *App) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil || expires < 1 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := SnippetCreateForm{Title: r.PostForm.Get("title"), Content: r.PostForm.Get("content"), Expires: expires, FieldErros: make(map[string]string)}

	if strings.TrimSpace(form.Title) == "" {
		form.FieldErros["title"] = "Title field cannot be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		form.FieldErros["title"] = "Title field cannot be longer than 100 characters"
	}
	if strings.TrimSpace(form.Content) == "" {
		form.FieldErros["content"] = "Content field cannot be blank"
	}

	if expires != 1 && expires != 7 && expires != 365 {
		form.FieldErros["expires"] = "This field must be equal 1, 7, 365"
	}

	if len(form.FieldErros) > 0 {
		data := app.newTempalteData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
