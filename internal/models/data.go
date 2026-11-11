package models

import "time"

// Type can be "checking", "savings", or "credit".
type AccountData struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Balance     float64   `json:"balance"`
	CreditLimit float64   `json:"credit_limit,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type LoanData struct {
	Number   string    `json:"number"`
	Borrower string    `json:"borrower"`
	Amount   float64   `json:"amount"`
	Rate     float64   `json:"rate"`
	Months   int       `json:"months"`
	Monthly  float64   `json:"monthly"`
	Total    float64   `json:"total"`
	Date     time.Time `json:"date"`
}

type TransactionRecord struct {
	AccountID  string    `json:"account_id"`
	Name       string    `json:"name"`
	Amount     float64   `json:"amount"`
	Income     bool      `json:"income"`
	Suspicious bool      `json:"suspicious"`
	Reason     string    `json:"reason,omitempty"`
	Date       time.Time `json:"date"`
}
