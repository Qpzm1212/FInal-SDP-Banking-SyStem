package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"ESHKEREE/internal/db"
	"ESHKEREE/internal/models"
	"ESHKEREE/internal/patterns"
)

var (
	ErrInvalidInput        = patterns.ErrInvalidInput
	ErrAccountNotFound     = db.ErrAccountNotFound
	ErrAccountExists       = db.ErrAccountExists
	ErrInsufficientFunds   = db.ErrInsufficientFunds
	ErrCreditLimitExceeded = db.ErrCreditLimitExceeded
)

type TransferResult struct {
	Suspicious bool                     `json:"suspicious"`
	Reason     string                   `json:"reason,omitempty"`
	Balance    float64                  `json:"balance"`
	Record     models.TransactionRecord `json:"transaction"`
}

// BankFacade ties together the ledger store and domain patterns for the HTTP layer.
type BankFacade struct {
	txMonitor *patterns.TransactionMonitor
}

func NewBankFacade() *BankFacade {
	monitor := patterns.NewTransactionMonitor()
	// FraudDetector has to run before SecurityTeamAlert so Suspicious is set when the alert checks it.
	monitor.AddObserver(&patterns.FraudDetector{Threshold: patterns.FraudThreshold})
	monitor.AddObserver(&patterns.AuditLogger{})
	monitor.AddObserver(&patterns.SecurityTeamAlert{})
	return &BankFacade{txMonitor: monitor}
}

func (f *BankFacade) OpenAccount(accType string, accID string) (models.AccountData, error) {
	acct, err := patterns.CreateAccount(accType, accID)
	if err != nil {
		return models.AccountData{}, err
	}
	record := models.AccountData{
		ID:          acct.GetID(),
		Type:        acct.GetType(),
		CreditLimit: acct.CreditLimit(),
		CreatedAt:   time.Now(),
	}
	if err := db.GetDBInstance().SaveAccount(record); err != nil {
		return models.AccountData{}, err
	}
	return record, nil
}

func (f *BankFacade) ListAccounts() []models.AccountData {
	return db.GetDBInstance().ListAccounts()
}

func (f *BankFacade) Transfer(accID string, amount float64) (TransferResult, error) {
	accID = strings.TrimSpace(accID)
	if accID == "" {
		return TransferResult{}, fmt.Errorf("%w: account number is required", ErrInvalidInput)
	}
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return TransferResult{}, fmt.Errorf("%w: amount must be greater than 0", ErrInvalidInput)
	}

	ledger := db.GetDBInstance()
	acct, ok := ledger.GetAccount(accID)
	if !ok {
		return TransferResult{}, ErrAccountNotFound
	}

	// Check balance / credit line before firing observers so rejected transfers stay out of audit logs.
	if acct.Type == db.CreditType {
		limit := acct.CreditLimit
		if limit <= 0 {
			limit = db.DefaultCreditLimit
		}
		if acct.Balance-amount < -limit {
			return TransferResult{}, ErrCreditLimitExceeded
		}
	} else if acct.Balance < amount {
		return TransferResult{}, ErrInsufficientFunds
	}

	if typedAcct, err := patterns.CreateAccount(acct.Type, acct.ID); err == nil {
		if amount > typedAcct.MaxTransfer() {
			return TransferResult{}, fmt.Errorf("%w: single transfer exceeds $%.0f limit for %s accounts", ErrInvalidInput, typedAcct.MaxTransfer(), acct.Type)
		}
	}

	ev := f.txMonitor.ProcessTransaction(amount, accID)

	entry := models.TransactionRecord{
		AccountID:  accID,
		Name:       "Transfer",
		Amount:     amount,
		Suspicious: ev.Suspicious,
		Reason:     ev.Reason,
		Date:       ev.Time,
	}
	balance, err := ledger.ApplyDebit(entry)
	if err != nil {
		return TransferResult{}, err
	}
	return TransferResult{
		Suspicious: ev.Suspicious,
		Reason:     ev.Reason,
		Balance:    balance,
		Record:     entry,
	}, nil
}

func (f *BankFacade) ListTransactions(limit int) []models.TransactionRecord {
	return db.GetDBInstance().ListTransactions(limit)
}

func (f *BankFacade) ApplyForLoan(borrower string, amount float64) (*patterns.LoanAgreement, error) {
	loan, err := patterns.LoanDirector{}.ConstructStandardLoan(patterns.NewStandardLoanBuilder(), borrower, amount)
	if err != nil {
		return nil, err
	}
	err = db.GetDBInstance().SaveLoan(loan.Number+"_"+loan.Borrower, models.LoanData{
		Number:   loan.Number,
		Borrower: loan.Borrower,
		Amount:   loan.Amount,
		Rate:     loan.Rate,
		Months:   loan.Months,
		Monthly:  loan.Monthly,
		Total:    loan.Total,
		Date:     loan.Date,
	})
	if err != nil {
		return nil, fmt.Errorf("could not save the loan: %w", err)
	}
	return loan, nil
}

func (f *BankFacade) CalculateYield(accountType string, balance float64) (float64, error) {
	if balance < 0 || math.IsNaN(balance) || math.IsInf(balance, 0) {
		return 0, fmt.Errorf("%w: amount must be a non-negative number", ErrInvalidInput)
	}
	strategy, err := patterns.GetStrategy(accountType)
	if err != nil {
		return 0, err
	}
	ctx := &patterns.InterestContext{}
	ctx.SetStrategy(strategy)
	return ctx.ExecuteStrategy(balance), nil
}
