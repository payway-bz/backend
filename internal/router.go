package internal

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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
	target, err := url.Parse(cfg.WebsiteURL)
	if err != nil {
		slog.Error("parsing website url failed", slog.Any("err", err))
		os.Exit(1)
	}

	r.Handle("/*", httputil.NewSingleHostReverseProxy(target))

	return &httpServer{r}, nil
}
