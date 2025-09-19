package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"backend/internal/auth"
	"backend/internal/httpx"
	"backend/internal/model/businessuser"
)

// attachGetRoutes registers the list (GET) endpoint.
func attachGetRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { getOrders(db, w, r) })
}

// getOrders handles GET /api/orders?business_id=...
//
// @Summary      List orders by business ID
// @Description  Returns all orders for the provided business_id. If business_id is omitted and the authenticated user belongs to exactly one business, that business will be used automatically. If the user belongs to zero or more than one business, an error is returned.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        business_id  query     string  false  "Business ID"
// @Success      200          {array}   object
// @Failure      400          {object}  ErrorResponse
// @Router       /api/orders [get]
func getOrders(db *sql.DB, w http.ResponseWriter, r *http.Request) {

	u, ok := auth.FirebaseUser(w, r)
	if !ok {
		return
	}

	// Create a scoped logger for this operation
	logger := slog.Default().With(
		slog.String("component", "orders"),
		slog.String("op", "getOrders"),
	)

	// Resolve internal user UUID from Firebase UID
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	bizID := strings.TrimSpace(r.URL.Query().Get("business_id"))

	if !businessuser.AssertUserBelongsToBusiness(ctx, db, w, bizID, u) {
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, created_at, updated_at, business_id, created_by, status, amount, description, customer_email, currency
		FROM "order"
		WHERE business_id = $1::uuid AND created_by = (SELECT id FROM "user" WHERE firebase_id = $2)
		ORDER BY created_at DESC`,
		bizID,
		u.UID,
	)
	if err != nil {
		logger.Error("query orders failed", slog.String("business_id", bizID), slog.Any("err", err))
		httpx.WriteInternalServerError(w)
		return
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		var o Order
		var desc sql.NullString
		var email sql.NullString
		if err := rows.Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt, &o.BusinessID, &o.CreatedBy, &o.Status, &o.Amount, &desc, &email, &o.Currency); err != nil {
			logger.Error("scan order row failed", slog.Any("err", err))
			httpx.WriteInternalServerError(w)
			return
		}
		if desc.Valid {
			o.Description = &desc.String
		}
		if email.Valid {
			o.CustomerEmail = &email.String
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		logger.Error("rows error after iteration", slog.Any("err", err))
		httpx.WriteInternalServerError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(orders)
}
