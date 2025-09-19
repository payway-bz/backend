package user

import (
	"database/sql"
	"net/http"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
)

// Routes exposes both public and private user endpoints under one router.
// Public endpoints (e.g., registration) are mounted without middleware.
// Private endpoints (e.g., get profile) are mounted with the provided middleware.
func Routes(db *sql.DB, fbAuth *firebaseauth.Client, mw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	// Public
	attachRegisterRoutes(r, db, fbAuth)

	// Private (apply middleware to the subrouter passed into attachGetRoutes)
	attachGetRoutes(r.With(mw), db)

	return r
}
