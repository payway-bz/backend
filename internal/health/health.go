package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes returns the router for health-related endpoints.
func Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/healthz", Healthz)
	return r
}

// Healthz is a simple liveness/readiness probe handler.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
