package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

func secureHeadersTest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src'self'; style-src'self' fonts.googleapis.com; img-src'self'; script-src'self'; font-src fonts.gstatic.com")
		next.ServeHTTP(w, r)
	})
}

func (app *App) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *App) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				//app.errorLog.Printf("panic: %v", err)
				app.serverError(w, fmt.Errorf("internal server error %s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *App) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// noSurf is a middleware that wraps the provided http.Handler with CSRF protection
// using the nosurf package. It configures the CSRF cookie to be HttpOnly, Secure,
// and available for the entire site (Path: "/"). This helps prevent Cross-Site
// Request Forgery attacks by requiring a valid CSRF token on state-changing requests.
func TestnoSurf(next http.Handler) http.Handler {
	csrfHadler := nosurf.New(next)
	csrfHadler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHadler
}

// authenticate is a middleware that checks if a user is authenticated by retrieving
// the "authenticatedUserID" from the session. If the user is authenticated and exists
// in the database, it adds an authentication flag to the request context. Otherwise,
// it passes the request to the next handler without modification. If a database error
// occurs during the existence check, it responds with a server error.
func (app *App) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}

		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})

}
