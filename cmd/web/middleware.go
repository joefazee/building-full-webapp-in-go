package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"abahjoseph.com/snippetbox/pkg/models"
	"github.com/justinas/nosurf"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})

}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.isAuthenticated(r) {
			http.Redirect(w, r, fmt.Sprintf("/user/login?redirectTo=%s", r.URL.Path), http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// Fetches the user’s ID from their session data, checks the database to see
// if the ID is valid and for an active user, and then updates the request context
// to include this information.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if a authenticatedUserID value exists in the session. If this *isn't
		// present* then call the next handler in the chain as normal.
		exists := app.session.Exists(r, authenticatedUserID)
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		// Fetch the details of the current user from the database. If no matching
		// record is found, remove the (invalid) authenticatedUserID value from their
		// session and call the next handler in the chain as normal.
		user, err := app.users.Get(app.session.GetInt(r, authenticatedUserID))
		if errors.Is(err, models.ErrNoRecord) {
			app.session.Remove(r, authenticatedUserID)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}

		// Likewise, if the the current user is has been deactivated remove the
		// authenticatedUserID value from their session and call the next handler in
		// the chain as normal.
		if !user.Active {
			app.session.Remove(r, authenticatedUserID)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we know that the request is coming from a active, authenticated,
		// user. We create a new copy of the request, with a true boolean value
		// added to the request context to indicate this, and call the next handler
		// in the chain *using this new copy of the request*.
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}
