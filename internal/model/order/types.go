package order

import "time"

type Order struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	BusinessID    string    `json:"business_id"`
	CreatedBy     string    `json:"created_by"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Description   *string   `json:"description,omitempty"`
	CustomerEmail *string   `json:"customer_email,omitempty"`
}
