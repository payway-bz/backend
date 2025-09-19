package internal

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	httpx "backend/internal/httpx"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/firebaseapp"
	"backend/internal/health"
	"backend/internal/model/order"
	"backend/internal/model/user"
)

type httpServer struct{ http.Handler }

func NewHTTPServer(cfg config.Config, db *sql.DB) (http.Handler, error) {
	r := chi.NewRouter()

	// Core middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Public routes
	r.Get("/ping", httpx.Ping)
	r.Mount("/", health.Routes())

	// Swagger UI (available at /swagger/index.html)
	// Note: The actual spec will appear after running `swag init` and importing the generated docs package.
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Firebase client
	var mw func(http.Handler) http.Handler
	_, fbAuth, err := firebaseapp.New(context.Background(), cfg.FirebaseCredentialsFile)
	if err != nil {
		return nil, err
	}
	mw = auth.NewFirebaseMiddleware(fbAuth)

	// API endpoints
	r.Route("/api", func(api chi.Router) {
		// User endpoints (public and private combined)
		api.Mount("/user", user.Routes(db, fbAuth, mw))

		// Private API endpoints (with auth middleware)
		api.With(mw).Mount("/orders", order.Routes(db))
	})

	// Catch-all must be last so it doesn't shadow /api/* and /swagger/*
	// Use the Docker service name "website" so the backend can reach the Vite dev server via the shared Docker network.
	r.Handle("/*", newReverseProxy("http://website:5173"))

	return &httpServer{r}, nil
}

// newReverseProxy returns an HTTP handler that proxies requests to the targetURL.
// This is used in development to forward unknown routes to the Vite dev server.
func newReverseProxy(targetURL string) http.Handler {
	u, err := url.Parse(targetURL)
	if err != nil {
		// If the target URL is invalid, fail fast in development.
		panic("invalid reverse proxy target: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// Preserve the incoming path and query while swapping scheme/host to the target.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Ensure Host header matches the target for some dev servers.
		req.Host = u.Host
		// Some servers are sensitive to an absent User-Agent; set empty if missing.
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "proxy error: "+err.Error(), http.StatusBadGateway)
	}

	return proxy
}
