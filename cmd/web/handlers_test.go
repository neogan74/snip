package main

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/neogan74/snip/internal/assert"
	"github.com/neogan74/snip/internal/models"
)

func TestPing(t *testing.T) {
	app := newTestApp(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}

func TestAppHome(t *testing.T) {
	app := newTestApp(t)
	app.snippets = &stubSnippetModel{
		latestFn: func() ([]*models.Snippet, error) {
			return []*models.Snippet{{Title: "example"}}, nil
		},
	}
	app.templateCache["home.tmpl.html"] = template.Must(template.New("home.tmpl.html").Parse(`{{define "base"}}{{range .Snippets}}{{.Title}}{{end}}{{end}}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.home(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "example")
}

func TestAppHomeError(t *testing.T) {
	app := newTestApp(t)
	app.snippets = &stubSnippetModel{
		latestFn: func() ([]*models.Snippet, error) {
			return nil, errors.New("latest failed")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.home(rr, req)

	assert.Equal(t, rr.Code, http.StatusInternalServerError)
}

func TestAppSnippetView(t *testing.T) {
	tests := []struct {
		name       string
		idValue    string
		model      models.SnippetModelInterface
		wantStatus int
		wantBody   string
	}{
		{
			name:       "invalid id",
			idValue:    "abc",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "not found",
			idValue:    "2",
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "ok",
			idValue: "1",
			model: &stubSnippetModel{
				getFn: func(id int) (*models.Snippet, error) {
					return &models.Snippet{ID: id, Title: "hello"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(t)
			if tt.model != nil {
				app.snippets = tt.model
			}
			app.templateCache["view.tmpl.html"] = template.Must(template.New("view.tmpl.html").Parse(`{{define "base"}}{{if .Snippet}}{{.Snippet.Title}}{{end}}{{end}}`))

			req := httptest.NewRequest(http.MethodGet, "/snippet/view/", nil)
			ctx, err := app.sessionManager.Load(req.Context(), "")
			if err != nil {
				t.Fatal(err)
			}
			ctx = context.WithValue(ctx, httprouter.ParamsKey, httprouter.Params{{Key: "id", Value: tt.idValue}})
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			app.snippetView(rr, req)

			assert.Equal(t, rr.Code, tt.wantStatus)
			if tt.wantBody != "" {
				assert.Equal(t, strings.TrimSpace(rr.Body.String()), tt.wantBody)
			}
		})
	}
}

func TestAppSnippetCreate(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["create.tmpl.html"] = template.Must(template.New("create.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{.Expires}}{{end}}{{end}}`))

	req := httptest.NewRequest(http.MethodGet, "/snippet/create", nil)
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.snippetCreate(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "365")
}

func TestAppSnippetCreatePostValidationErrors(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["create.tmpl.html"] = template.Must(template.New("create.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{index .FieldErrors "title"}}{{end}}{{end}}`))

	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/snippet/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.snippetCreatePost(rr, req)

	assert.Equal(t, rr.Code, http.StatusUnprocessableEntity)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "Title field cannot be blank")
}

func TestAppSnippetCreatePostSuccess(t *testing.T) {
	app := newTestApp(t)
	form := url.Values{}
	form.Set("title", "Test Title")
	form.Set("content", "Test Content")
	form.Set("expires", "7")

	req := httptest.NewRequest(http.MethodPost, "/snippet/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.snippetCreatePost(rr, req)

	assert.Equal(t, rr.Code, http.StatusSeeOther)
	assert.Equal(t, rr.Header().Get("Location"), "/snippet/view/2")
	flash := app.sessionManager.PopString(ctx, "flash")
	assert.Equal(t, flash, "Snippet successfully created")
}

func TestAppUserSignUp(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["signup.tmpl.html"] = template.Must(template.New("signup.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{.Email}}{{end}}{{end}}`))

	req := httptest.NewRequest(http.MethodGet, "/user/signup", nil)
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userSignUp(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "")
}

func TestAppUserSignUpPostValidationErrors(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["signup.tmpl.html"] = template.Must(template.New("signup.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{index .FieldErrors "name"}}{{end}}{{end}}`))

	req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(url.Values{}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userSignUpPost(rr, req)

	assert.Equal(t, rr.Code, http.StatusUnprocessableEntity)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "Name field cannot be blank")
}

func TestAppUserSignUpPostDuplicateEmail(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["signup.tmpl.html"] = template.Must(template.New("signup.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{index .FieldErrors "email"}}{{end}}{{end}}`))

	form := url.Values{}
	form.Set("name", "Test")
	form.Set("email", "dupe@example.com")
	form.Set("password", "12345678")

	req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userSignUpPost(rr, req)

	assert.Equal(t, rr.Code, http.StatusUnprocessableEntity)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "This email address is already in use")
}

func TestAppUserSignUpPostSuccess(t *testing.T) {
	app := newTestApp(t)
	form := url.Values{}
	form.Set("name", "Tester")
	form.Set("email", "new@example.com")
	form.Set("password", "12345678")

	req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userSignUpPost(rr, req)

	assert.Equal(t, rr.Code, http.StatusSeeOther)
	assert.Equal(t, rr.Header().Get("Location"), "/user/login")
	flash := app.sessionManager.PopString(ctx, "flash")
	assert.Equal(t, flash, "User successfully created")
}

func TestAppUserLogin(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["login.tmpl.html"] = template.Must(template.New("login.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{.Email}}{{end}}{{end}}`))

	req := httptest.NewRequest(http.MethodGet, "/user/login", nil)
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userLogin(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "")
}

func TestAppUserLoginPostValidationErrors(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["login.tmpl.html"] = template.Must(template.New("login.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{index .FieldErrors "email"}}{{end}}{{end}}`))

	req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(url.Values{}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userLoginPost(rr, req)

	assert.Equal(t, rr.Code, http.StatusUnprocessableEntity)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "Email field cannot be blank")
}

func TestAppUserLoginPostInvalidCredentials(t *testing.T) {
	app := newTestApp(t)
	app.templateCache["login.tmpl.html"] = template.Must(template.New("login.tmpl.html").Parse(`{{define "base"}}{{with .Form}}{{range .NonFieldErrors}}{{.}}{{end}}{{end}}{{end}}`))

	form := url.Values{}
	form.Set("email", "nope@example.com")
	form.Set("password", "wrongpass")

	req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userLoginPost(rr, req)

	assert.Equal(t, rr.Code, http.StatusUnprocessableEntity)
	assert.Equal(t, strings.TrimSpace(rr.Body.String()), "Email or password is incorrect")
}

func TestAppUserLoginPostSuccess(t *testing.T) {
	app := newTestApp(t)
	form := url.Values{}
	form.Set("email", "arina@neogan.com")
	form.Set("password", "123")

	req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx, err := app.sessionManager.Load(req.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userLoginPost(rr, req)

	assert.Equal(t, rr.Code, http.StatusSeeOther)
	assert.Equal(t, rr.Header().Get("Location"), "/snippet/create")
	id := app.sessionManager.GetInt(ctx, "authenticatedUserID")
	assert.Equal(t, id, 1)
}

func TestAppUserLogoutPost(t *testing.T) {
	app := newTestApp(t)
	ctx, err := app.sessionManager.Load(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	app.sessionManager.Put(ctx, "authenticatedUserID", 99)
	req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	app.userLogoutPost(rr, req)

	assert.Equal(t, rr.Code, http.StatusSeeOther)
	assert.Equal(t, rr.Header().Get("Location"), "/")
	id := app.sessionManager.GetInt(ctx, "authenticatedUserID")
	assert.Equal(t, id, 0)
	flash := app.sessionManager.PopString(ctx, "flash")
	assert.Equal(t, flash, "You've been logged out successfully!")
}

type stubSnippetModel struct {
	insertFn func(title, content string, expires int) (int, error)
	getFn    func(id int) (*models.Snippet, error)
	latestFn func() ([]*models.Snippet, error)
}

func (m *stubSnippetModel) Insert(title, content string, expires int) (int, error) {
	if m.insertFn != nil {
		return m.insertFn(title, content, expires)
	}
	return 0, nil
}

func (m *stubSnippetModel) Get(id int) (*models.Snippet, error) {
	if m.getFn != nil {
		return m.getFn(id)
	}
	return nil, models.ErrNoRecord
}

func (m *stubSnippetModel) Latest() ([]*models.Snippet, error) {
	if m.latestFn != nil {
		return m.latestFn()
	}
	return nil, nil
}
