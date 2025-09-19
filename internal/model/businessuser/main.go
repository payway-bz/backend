package businessuser

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"backend/internal/auth"
	httpx "backend/internal/httpx"
)

// AssertUserBelongsToBusiness checks whether the given user belongs to the given business.
// It logs and writes an HTTP error to the ResponseWriter when the check fails or on internal error.
// Returns true if membership exists and the request may proceed; false otherwise (an error response has been written).
func AssertUserBelongsToBusiness(ctx context.Context, db *sql.DB, w http.ResponseWriter, businessID string, u *auth.User) bool {
	var exists bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM business_user bu
			JOIN "user" usr ON usr.id = bu.user_id
			WHERE bu.business_id = $1::uuid AND usr.firebase_id = $2
		)`, businessID, u.UID,
	).Scan(&exists); err != nil {
		slog.Default().With(
			slog.String("component", "businessuser"),
			slog.String("op", "AssertUserBelongsToBusiness"),
			slog.String("business_id", businessID),
			slog.String("firebase_id", u.UID),
		).Error("membership check failed", slog.Any("err", err))
		httpx.WriteInternalServerError(w)
		return false
	}
	if !exists {
		httpx.WriteForbidden(w)
		return false
	}
	return true
}
