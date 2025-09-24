package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
)

func newRoutesTestApp(t *testing.T) *App {
	t.Helper()

	app := newTestApp(t)
	app.sessionManager = scs.New()
	app.templateCache = map[string]*template.Template{
		"login.tmpl.html":  template.Must(template.New("login.tmpl.html").Parse(`{{define "base"}}login{{end}}`)),
		"signup.tmpl.html": template.Must(template.New("signup.tmpl.html").Parse(`{{define "base"}}signup{{end}}`)),
	}

	return app
}

func TestRoutes(t *testing.T) {
	app := newRoutesTestApp(t)
	handler := app.routes()

	tests := []struct {
		name         string
		method       string
		target       string
		wantStatus   int
		wantLocation string
	}{
		{name: "Ping", method: http.MethodGet, target: "/ping", wantStatus: http.StatusOK},
		{name: "UserSignupGet", method: http.MethodGet, target: "/user/signup", wantStatus: http.StatusOK},
		{name: "UserLoginGet", method: http.MethodGet, target: "/user/login", wantStatus: http.StatusOK},
		{name: "SnippetCreateGet", method: http.MethodGet, target: "/snippet/create", wantStatus: http.StatusSeeOther, wantLocation: "/user/login"},
		{name: "SnippetCreatePost", method: http.MethodPost, target: "/snippet/create", wantStatus: http.StatusBadRequest},
		{name: "UserLogoutPost", method: http.MethodPost, target: "/user/logout", wantStatus: http.StatusBadRequest},
		{name: "NotFound", method: http.MethodGet, target: "/does-not-exist", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("%s: got status %d, want %d", tt.name, rr.Code, tt.wantStatus)
			}

			if tt.wantLocation != "" {
				if got := rr.Header().Get("Location"); got != tt.wantLocation {
					t.Fatalf("%s: got location %q, want %q", tt.name, got, tt.wantLocation)
				}
			}
		})
	}
}

func TestStaticFileRoute(t *testing.T) {
	app := newRoutesTestApp(t)
	handler := app.routes()

	req := httptest.NewRequest(http.MethodGet, "/static/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusOK)
	}
}
