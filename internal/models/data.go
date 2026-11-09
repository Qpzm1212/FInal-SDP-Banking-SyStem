package models

import "time"

// AccountData - запись о счете в хранилище.
// Type хранит КОД типа ("checking" / "savings"), а не текст для интерфейса.
type AccountData struct {
	ID        string
	Type      string
	CreatedAt time.Time
}

// LoanData - запись о кредите.
type LoanData struct {
	Borrower string
	Amount   float64
	Rate     float64
	Months   int
	Date     time.Time
}

// TransactionRecord - запись о транзакции.
type TransactionRecord struct {
	AccountID  string
	Amount     float64
	Suspicious bool
	Date       time.Time
}
