package auth

import (
	"context"
	"net/http"
	"strings"

	firebaseauth "firebase.google.com/go/v4/auth"

	httpx "backend/internal/httpx"
)

// context key type to avoid collisions
type ctxKey string

const (
	ctxUserKey ctxKey = "firebaseUser"
)

// User holds a subset of Firebase token/user info for downstream handlers.
// Extend as needed.
type User struct {
	UID         string
	Email       string
	DisplayName string
	Claims      map[string]any
}

// FirebaseUser extracts the authenticated Firebase user from the context.
func FirebaseUser(w http.ResponseWriter, r *http.Request) (*User, bool) {
	u, ok := r.Context().Value(ctxUserKey).(*User)
	if !ok || u == nil || u.UID == "" {
		httpx.WriteUnauthorized(w)
		return nil, false
	}
	return u, true
}

// NewFirebaseMiddleware returns an HTTP middleware that verifies Firebase ID tokens
// from the Authorization: Bearer header. On success, it injects a User into context.
func NewFirebaseMiddleware(authClient *firebaseauth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Bearer ") {
				http.Error(w, "missing bearer", http.StatusUnauthorized)
				return
			}
			raw := strings.TrimPrefix(authz, "Bearer ")

			token, err := authClient.VerifyIDToken(r.Context(), raw)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Optional: fetch user record for richer info
			var email, displayName string
			if token.UID != "" {
				if ur, err := authClient.GetUser(r.Context(), token.UID); err == nil && ur != nil {
					email = ur.Email
					displayName = ur.DisplayName
				}
			}

			u := &User{UID: token.UID, Email: email, DisplayName: displayName, Claims: token.Claims}
			ctx := context.WithValue(r.Context(), ctxUserKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
