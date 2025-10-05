package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neogan74/snip/internal/assert"
	"github.com/neogan74/snip/internal/models"
)

// Test_secureHeaders verifies that the secureHeaders middleware correctly sets
// the expected HTTP security headers on the response, including Content-Security-Policy,
// X-Content-Type-Options, X-Frame-Options, and X-XSS-Protection. It also checks that
// the middleware passes through the response body and status code as expected.
func Test_secureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	secureHeaders(next).ServeHTTP(rr, r)

	rs := rr.Result()

	expectedValue := "default-src'self'; style-src'self' fonts.googleapis.com; img-src'self'; script-src'self'; font-src fonts.gstatic.com"
	assert.Equal(t, rs.Header.Get("Content-Security-Policy"), expectedValue)

	expectedValue = "nosniff"
	assert.Equal(t, rs.Header.Get("X-Content-Type-Options"), expectedValue)

	expectedValue = "deny"
	assert.Equal(t, rs.Header.Get("X-Frame-Options"), expectedValue)

	expectedValue = "1; mode=block"
	assert.Equal(t, rs.Header.Get("X-XSS-Protection"), expectedValue)

	assert.Equal(t, rs.StatusCode, http.StatusOK)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	assert.Equal(t, string(body), "OK")
}

func TestRecoverPanic(t *testing.T) {
	app := newTestApp(t)
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	handler := app.recoverPanic(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	rs := rr.Result()
	assert.Equal(t, rs.StatusCode, http.StatusInternalServerError)
	assert.Equal(t, rs.Header.Get("Connection"), "close")

	defer rs.Body.Close()
}

func TestRequireAuthenticationRedirectsUnauthenticated(t *testing.T) {
	app := newTestApp(t)
	called := false
	handler := app.requireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/snippet/create", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	rs := rr.Result()
	assert.Equal(t, rs.StatusCode, http.StatusSeeOther)
	assert.Equal(t, rs.Header.Get("Location"), "/user/login")
	assert.Equal(t, called, false)
	defer rs.Body.Close()
}

func TestRequireAuthenticationAllowsAuthenticated(t *testing.T) {
	app := newTestApp(t)
	called := false
	handler := app.requireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if !app.isAuthenticated(r) {
			t.Fatal("expected authenticated request")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/snippet/create", nil)
	ctx := context.WithValue(req.Context(), isAuthenticatedContextKey, true)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	rs := rr.Result()
	assert.Equal(t, rs.StatusCode, http.StatusOK)
	assert.Equal(t, rs.Header.Get("Cache-Control"), "no-store")
	assert.Equal(t, called, true)
	defer rs.Body.Close()
}

func TestAuthenticateMiddleware(t *testing.T) {
	t.Run("no session", func(t *testing.T) {
		app := newTestApp(t)
		var called bool
		handler := app.authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			if app.isAuthenticated(r) {
				t.Fatal("did not expect authenticated user")
			}
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx, err := app.sessionManager.Load(req.Context(), "")
		if err != nil {
			t.Fatal(err)
		}
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, called, true)
	})

	t.Run("authenticated user", func(t *testing.T) {
		app := newTestApp(t)
		app.users = &stubUserModel{
			existsFn: func(id int) (bool, error) {
				if id != 1 {
					t.Fatalf("unexpected id %d", id)
				}
				return true, nil
			},
		}

		var called bool
		handler := app.authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			if !app.isAuthenticated(r) {
				t.Fatal("expected authenticated request")
			}
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx, err := app.sessionManager.Load(req.Context(), "")
		if err != nil {
			t.Fatal(err)
		}
		app.sessionManager.Put(ctx, "authenticatedUserID", 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, called, true)
	})

	t.Run("user lookup error", func(t *testing.T) {
		app := newTestApp(t)
		app.users = &stubUserModel{
			existsFn: func(id int) (bool, error) {
				return false, errors.New("db down")
			},
		}

		handler := app.authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("middleware should not call next")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx, err := app.sessionManager.Load(req.Context(), "")
		if err != nil {
			t.Fatal(err)
		}
		app.sessionManager.Put(ctx, "authenticatedUserID", 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, http.StatusInternalServerError)
	})
}

func TestNoSurfSetsCookie(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := noSurf(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	rs := rr.Result()
	defer rs.Body.Close()
	assert.Equal(t, rs.StatusCode, http.StatusOK)

	var csrfCookie *http.Cookie
	for _, c := range rs.Cookies() {
		if c.Name == "csrf_token" {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("expected csrf_token cookie to be set")
	}

	assert.Equal(t, csrfCookie.Path, "/")
	assert.Equal(t, csrfCookie.HttpOnly, true)
	assert.Equal(t, csrfCookie.Secure, true)
}

type stubUserModel struct {
	insertFn       func(name, email, password string) error
	authenticateFn func(email, password string) (int, error)
	existsFn       func(id int) (bool, error)
}

func (m *stubUserModel) Insert(name, email, password string) error {
	if m.insertFn != nil {
		return m.insertFn(name, email, password)
	}
	return nil
}

func (m *stubUserModel) Authenticate(email, password string) (int, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(email, password)
	}
	return 0, models.ErrInvalidCredentials
}

func (m *stubUserModel) Exists(id int) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(id)
	}
	return false, nil
}
