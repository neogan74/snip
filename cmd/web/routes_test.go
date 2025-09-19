package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/neogan74/snip/ui"
)

// Define a minimal interface for session management used in App
type sessionManagerInterface interface {
	LoadAndSave(http.Handler) http.Handler
}

func newMockApp() *App {
	return &App{
		// Use interface{} or a custom interface for sessionManager in tests
		sessionManager: &mockSessionManager{},
	}
}

// --- Mock dependencies ---

type mockSessionManager struct{}

// Ensure mockSessionManager implements sessionManagerInterface
var _ sessionManagerInterface = (*mockSessionManager)(nil)

// --- Mock dependencies ---

type mockSessionManager struct{}

func (m *mockSessionManager) LoadAndSave(next http.Handler) http.Handler {
	return next
}

// Dummy middleware and handlers for testing
func noSurf(next http.Handler) http.Handler                           { return next }
func (app *App) authenticate(next http.Handler) http.Handler          { return next }
func (app *App) requireAuthentication(next http.Handler) http.Handler { return next }
func (app *App) recoverPanic(next http.Handler) http.Handler          { return next }
func (app *App) logRequest(next http.Handler) http.Handler            { return next }
func secureHeaders(next http.Handler) http.Handler                    { return next }

func (app *App) notFound(w http.ResponseWriter) {
	http.Error(w, "custom 404", http.StatusNotFound)
}

func (app *App) home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("home"))
}
func (app *App) snippetView(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("snippet view"))
}
func (app *App) userSignUp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("signup"))
}
func (app *App) userSignUpPost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("signup post"))
}
func (app *App) userLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("login"))
}
func (app *App) userLoginPost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("login post"))
}
func (app *App) snippetCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("snippet create"))
}
func (app *App) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("snippet create post"))
}
func (app *App) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logout"))
}

// ping handler for /ping route
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// Mock ui.Files for static file serving
type mockFS struct{}

func (m mockFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}

var _ fs.FS = mockFS{}

func init() {
	// Patch ui.Files for tests
	ui.Files = mockFS{}
}

func TestRoutes(t *testing.T) {
	app := newMockApp()
	handler := app.routes()

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantBody   string
	}{
		{"Home", "GET", "/", http.StatusOK, "home"},
		{"SnippetView", "GET", "/snippet/view/1", http.StatusOK, "snippet view"},
		{"UserSignupGet", "GET", "/user/signup", http.StatusOK, "signup"},
		{"UserSignupPost", "POST", "/user/signup", http.StatusOK, "signup post"},
		{"UserLoginGet", "GET", "/user/login", http.StatusOK, "login"},
		{"UserLoginPost", "POST", "/user/login", http.StatusOK, "login post"},
		{"SnippetCreateGet", "GET", "/snippet/create", http.StatusOK, "snippet create"},
		{"SnippetCreatePost", "POST", "/snippet/create", http.StatusOK, "snippet create post"},
		{"UserLogoutPost", "POST", "/user/logout", http.StatusOK, "logout"},
		{"Ping", "GET", "/ping", http.StatusOK, "OK"},
		{"NotFound", "GET", "/doesnotexist", http.StatusNotFound, "custom 404"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("got body %q, want to contain %q", rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestStaticFileRoute(t *testing.T) {
	app := newMockApp()
	handler := app.routes()

	req := httptest.NewRequest("GET", "/static/test.txt", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Since mockFS always returns ErrNotExist, expect 404
	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
	}
}
