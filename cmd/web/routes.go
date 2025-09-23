package main

import (
	"io/fs"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/neogan74/snip/ui"
)

// routes sets up the application's HTTP routes and middleware stack.
// It configures custom handlers for not found routes, serves static files,
// and defines both protected and unprotected endpoints for snippets and user authentication.
// Middleware chains are used for session management, CSRF protection, authentication,
// panic recovery, request logging, and secure headers.
// Returns an http.Handler that is ready to be used by the HTTP server.
func (app *App) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	staticFS, err := fs.Sub(ui.Files, "static")
	if err != nil {
		app.errorLog.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(staticFS))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// unprotected wuth auth group of endpoints
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	// User related login and logout
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignUp))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignUpPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
