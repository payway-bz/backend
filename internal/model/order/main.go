package order

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes aggregates all order submodule routes (create, get, etc.)
func Routes(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	attachCreateRoutes(r, db)
	attachGetRoutes(r, db)
	return r
}
