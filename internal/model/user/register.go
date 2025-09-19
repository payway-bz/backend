package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend/internal/httpx"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
)

// attachRegisterRoutes registers register (POST) endpoint(s).
func attachRegisterRoutes(r chi.Router, db *sql.DB, fbAuth *firebaseauth.Client) {
	r.Post("/", func(w http.ResponseWriter, r *http.Request) { registerUser(db, w, r, fbAuth) })
}

type registerInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
}

type registerResponse struct {
	ID         string `json:"id"`
	FirebaseID string `json:"firebase_id"`
	Name       string `json:"name"`
	LastName   string `json:"last_name"`
}

// registerUser creates (or idempotently upserts) a user row
// and creates a Firebase user.
func registerUser(db *sql.DB, w http.ResponseWriter, r *http.Request, fbAuth *firebaseauth.Client) {
	var in registerInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Email = strings.TrimSpace(in.Email)
	in.Password = strings.TrimSpace(in.Password)
	in.Name = strings.TrimSpace(in.Name)
	in.LastName = strings.TrimSpace(in.LastName)
	if in.Email == "" || in.Password == "" || in.Name == "" || in.LastName == "" {
		httpx.WriteErr(w, http.StatusBadRequest, "email, password, name and last_name are required")
		return
	}

	fbCtx, fbCancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer fbCancel()
	fbUser, err := createFirebaseUser(fbCtx, fbAuth, in.Email, in.Password)
	if err != nil {
		httpx.WriteErr(w, http.StatusInternalServerError, "firebase: "+err.Error())
		return
	}

	// Insert into DB
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var id string
	row := db.QueryRowContext(ctx, `
			INSERT INTO "user" (firebase_id, name, last_name)
			VALUES ($1, $2, $3)
			RETURNING id
		`, fbUser.UID, in.Name, in.LastName)
	if err := row.Scan(&id); err != nil {
		httpx.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(registerResponse{
		ID:         id,
		FirebaseID: fbUser.UID,
		Name:       in.Name,
		LastName:   in.LastName,
	})
}

// createFirebaseUser returns a user or an error; no HTTP writes inside.
func createFirebaseUser(ctx context.Context, fbAuth *firebaseauth.Client, email string, password string) (*firebaseauth.UserRecord, error) {
	params := (&firebaseauth.UserToCreate{}).
		Email(email).
		Password(password)
	user, err := fbAuth.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return user, nil
}
