package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
	httpx "backend/internal/httpx"
	"backend/internal/model/businessuser"

	"github.com/go-chi/chi/v5"
)

// OrderPayload matches the fields the client sends
// Adjust types/validation as your API evolves.
type OrderPayload struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Email       string  `json:"email"`
	Currency    string  `json:"currency"`
	BusinessID  string  `json:"business_id"`
}

// attachCreateRoutes registers the create (POST) endpoint.
func attachCreateRoutes(r chi.Router, db *sql.DB) {
	r.Post("/", func(w http.ResponseWriter, r *http.Request) { createOrder(db, w, r) })
}

// createOrder handles POST /api/orders
//
// @Summary      Create an order
// @Description  Creates a new order and returns its ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        payload  body      OrderPayload          true  "Order payload"
// @Success      200      {object}  CreateOrderResponse
// @Failure      400      {object}  ErrorResponse
// @Router       /api/orders [post]
func createOrder(db *sql.DB, w http.ResponseWriter, r *http.Request) {

	u, ok := auth.FirebaseUser(w, r)
	if !ok {
		return
	}

	defer r.Body.Close()
	var p OrderPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		httpx.WriteBadRequest(w)
		return
	}

	if err := normalizeAndValidate(&p); err != nil {
		slog.Error("normalize and validate error", slog.Any("err", err))
		httpx.WriteBadRequest(w)
		return
	}

	businessID := strings.TrimSpace(p.BusinessID)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if !businessuser.AssertUserBelongsToBusiness(ctx, db, w, businessID, u) {
		return
	}

	order, ok := insertOrder(w, ctx, db, p, u)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

// normalizeAndValidate uppercases currency and performs minimal validations.
func normalizeAndValidate(p *OrderPayload) error {
	if p.Amount <= 0 {
		return errors.New("amount must be > 0")
	}
	if strings.TrimSpace(p.BusinessID) == "" {
		return errors.New("business_id is required")
	}
	p.Currency = strings.ToUpper(strings.TrimSpace(p.Currency))
	if len(p.Currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}
	return nil
}

// insertOrder performs the INSERT and returns the created Order. It resolves created_by via firebase_id in a subquery.
func insertOrder(w http.ResponseWriter, ctx context.Context, db *sql.DB, p OrderPayload, u *auth.User) (Order, bool) {
	query := `INSERT INTO "order" (business_id, created_by, amount, description, customer_email, currency)
		VALUES ($1, (SELECT id FROM "user" WHERE firebase_id = $2), $3, $4, $5, $6)
		RETURNING id, created_at, updated_at, business_id, created_by, status, amount, currency, description, customer_email`

	var ord Order
	var descNS, emailNS sql.NullString
	if err := db.QueryRowContext(ctx, query, p.BusinessID, u.UID, p.Amount, nullIfEmpty(p.Description), nullIfEmpty(p.Email), p.Currency).
		Scan(&ord.ID, &ord.CreatedAt, &ord.UpdatedAt, &ord.BusinessID, &ord.CreatedBy, &ord.Status, &ord.Amount, &ord.Currency, &descNS, &emailNS); err != nil {
		slog.Error("create order insert error", slog.Any("err", err))
		httpx.WriteInternalServerError(w)
		return Order{}, false
	}
	if descNS.Valid {
		d := descNS.String
		ord.Description = &d
	}
	if emailNS.Valid {
		e := emailNS.String
		ord.CustomerEmail = &e
	}
	return ord, true
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{String: "", Valid: false}
	}
	return s
}
