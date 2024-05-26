package main

import (
	"context"
	"fmt"
	"net/http"
)

func (app *application) applogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infolog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self' https://s3.amazonaws.com https://gravatar.com http://lorempixum.com http://www.ultraimg.com; " +
			"style-src 'self' https://cdnjs.cloudflare.com https://fonts.googleapis.com https://netdna.bootstrapcdn.com; " +
			"font-src 'self' https://fonts.gstatic.com https://netdna.bootstrapcdn.com; " +
			"script-src 'self' https://code.jquery.com 'unsafe-inline'; " +
			"img-src 'self' data: https://s3.amazonaws.com https://gravatar.com http://lorempixum.com http://www.ultraimg.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// requierAuthentification is a middleware function that checks if the current
// request is authenticated. If the request is not authenticated, it redirects
// the user to the login page. If the request is authenticated, it sets the
// "Cache-Control" header to "no-store" to prevent caching of the response.
func (app *application) requierAuthentification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			// Add the path that the user is trying to access to their session
			// data.
			app.sessionManager.Put(r.Context(), "redirectedPathAfterLogging", r.URL.Path)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the authenticatedUserID value from the session using the
		// GetInt() method. This will return the zero value for an int (0) if no
		// "authenticatedUserID" value is in the session -- in which case we
		// call the next handler in the chain as normal and return.
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}
		exists, err := app.userData.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		// If a matching user is found, we know that the request is
		// coming from an authenticated user who exists in our database. We
		// create a new copy of the request (with an isAuthenticatedContextKey
		// value of true in the request context) and assign it to r.
		if exists {
			r = r.WithContext(context.WithValue(r.Context(), isAuthenticatedContextKey, true))
		}
		next.ServeHTTP(w, r)

	})
}
