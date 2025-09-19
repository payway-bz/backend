package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/httpx"

	"github.com/go-chi/chi/v5"
)

// attachGetRoutes registers GET endpoints for user
func attachGetRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { getUser(db, w, r) })
}

// MeResponse is the payload for the GET / endpoint
type GetUserResponse struct {
	ID         string           `json:"id"`
	Name       *string          `json:"name,omitempty"`
	LastName   *string          `json:"last_name,omitempty"`
	Businesses []BusinessRecord `json:"businesses"`
}

type BusinessRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// getUser returns profile + businesses for the Firebase-authenticated principal.
// Assumes the user was previously created via POST /api/user.
func getUser(db *sql.DB, w http.ResponseWriter, r *http.Request) {

	u, ok := auth.FirebaseUser(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Fetch user by firebase_id (must already exist)
	var userID string
	var name sql.NullString
	var lastName sql.NullString
	err := db.QueryRowContext(ctx, `SELECT id, name, last_name FROM "user" WHERE firebase_id = $1`, u.UID).Scan(&userID, &name, &lastName)
	if err == sql.ErrNoRows {
		// This should not happen if the client called POST /api/user beforehand
		httpx.WriteErr(w, http.StatusInternalServerError, "user not initialized")
		return
	} else if err != nil {
		slog.Error("query user failed", slog.Any("err", err))
		httpx.WriteInternalServerError(w)
		return
	}

	// Fetch associated businesses
	bizRows, err := db.QueryContext(ctx, `
		SELECT b.id, b.name
		FROM business_user bu
		JOIN business b ON b.id = bu.business_id
		WHERE bu.user_id = $1
		ORDER BY b.created_at DESC
	`, userID)
	if err != nil {
		slog.Error("query businesses failed", slog.Any("err", err))
		httpx.WriteInternalServerError(w)
		return
	}
	defer bizRows.Close()

	var businesses []BusinessRecord
	for bizRows.Next() {
		var b BusinessRecord
		if err := bizRows.Scan(&b.ID, &b.Name); err != nil {
			slog.Error("scan business row failed", slog.Any("err", err))
			httpx.WriteInternalServerError(w)
			return
		}
		businesses = append(businesses, b)
	}
	if err := bizRows.Err(); err != nil {
		slog.Error("rows error after iteration", slog.Any("err", err))
		httpx.WriteInternalServerError(w)
		return
	}

	var namePtr *string
	if name.Valid {
		n := name.String
		namePtr = &n
	}
	var lastNamePtr *string
	if lastName.Valid {
		ln := lastName.String
		lastNamePtr = &ln
	}

	resp := GetUserResponse{ID: userID, Name: namePtr, LastName: lastNamePtr, Businesses: businesses}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
