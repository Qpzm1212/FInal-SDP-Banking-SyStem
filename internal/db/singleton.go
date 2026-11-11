package db

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"ESHKEREE/internal/models"
)

var (
	ErrAccountExists       = errors.New("an account with this number already exists")
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrCreditLimitExceeded = errors.New("credit limit exceeded")
)

const (
	defaultLedgerPath  = "data/database.json"
	CreditType         = "credit"
	DefaultCreditLimit = 5000.0
)

type LedgerData struct {
	Loans        map[string]models.LoanData            `json:"loans"`
	Transactions map[string][]models.TransactionRecord `json:"transactions"`
	Accounts     map[string]models.AccountData         `json:"accounts"`
}

// LedgerStore wraps the JSON file behind an RWMutex because Echo handles requests concurrently.
type LedgerStore struct {
	FilePath string
	mu       sync.RWMutex
	data     LedgerData
}

var (
	sharedLedger *LedgerStore
	initOnce     sync.Once
)

// GetDBInstance returns the singleton store; BANK_DB_PATH overrides the file path in tests.
func GetDBInstance() *LedgerStore {
	initOnce.Do(func() {
		path := os.Getenv("BANK_DB_PATH")
		if path == "" {
			path = defaultLedgerPath
		}
		sharedLedger = openLedger(path)
	})
	return sharedLedger
}

func openLedger(path string) *LedgerStore {
	store := &LedgerStore{FilePath: path}
	if err := store.readFromDisk(); err != nil {
		log.Printf("ledger read failed (%s), starting fresh: %v", path, err)
	}
	store.initCollections()
	if len(store.data.Accounts) == 0 {
		store.loadStarterAccounts()
	}
	return store
}

func (s *LedgerStore) initCollections() {
	if s.data.Loans == nil {
		s.data.Loans = map[string]models.LoanData{}
	}
	if s.data.Transactions == nil {
		s.data.Transactions = map[string][]models.TransactionRecord{}
	}
	if s.data.Accounts == nil {
		s.data.Accounts = map[string]models.AccountData{}
	}
}

// First run without an existing DB file gets a couple of accounts so the dashboard isn't empty.
func (s *LedgerStore) loadStarterAccounts() {
	now := time.Now()
	day := 24 * time.Hour

	s.data.Accounts["ACC-1"] = models.AccountData{
		ID:        "ACC-1",
		Type:      "checking",
		Balance:   48250.75,
		CreatedAt: now.Add(-30 * day),
	}
	s.data.Accounts["ACC-2"] = models.AccountData{
		ID:        "ACC-2",
		Type:      "savings",
		Balance:   12400,
		CreatedAt: now.Add(-29 * day),
	}
	s.data.Transactions["ACC-1"] = []models.TransactionRecord{
		{AccountID: "ACC-1", Name: "Starbucks Coffee", Amount: 6.8, Date: now.Add(-4 * day)},
		{AccountID: "ACC-1", Name: "Yandex Go", Amount: 12.5, Date: now.Add(-3 * day)},
		{AccountID: "ACC-1", Name: "Payroll Deposit", Amount: 3200, Income: true, Date: now.Add(-2 * day)},
		{AccountID: "ACC-1", Name: "Magnum Market", Amount: 84.2, Date: now.Add(-7 * time.Hour)},
	}
	s.data.Transactions["ACC-2"] = []models.TransactionRecord{
		{AccountID: "ACC-2", Name: "Monthly Deposit Yield", Amount: 52, Income: true, Date: now.Add(-6 * day)},
	}

	if err := s.writeLedgerFile(); err != nil {
		log.Printf("failed to write starter accounts: %v", err)
	}
}

func (s *LedgerStore) readFromDisk() error {
	raw, err := os.ReadFile(s.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, &s.data)
}

// Caller must hold s.mu.
func (s *LedgerStore) writeLedgerFile() error {
	encoded, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.FilePath), 0755); err != nil {
		return err
	}
	tmpFile := s.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, encoded, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, s.FilePath)
}

func (s *LedgerStore) GetAccount(id string) (models.AccountData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	acct, ok := s.data.Accounts[id]
	return acct, ok
}

func (s *LedgerStore) ListAccounts() []models.AccountData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]models.AccountData, 0, len(s.data.Accounts))
	for _, a := range s.data.Accounts {
		accounts = append(accounts, a)
	}
	sort.Slice(accounts, func(i, j int) bool {
		if accounts[i].CreatedAt.Equal(accounts[j].CreatedAt) {
			return accounts[i].ID < accounts[j].ID
		}
		return accounts[i].CreatedAt.Before(accounts[j].CreatedAt)
	})
	return accounts
}

func (s *LedgerStore) SaveAccount(acct models.AccountData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data.Accounts[acct.ID]; exists {
		return ErrAccountExists
	}
	if acct.Type == CreditType && acct.CreditLimit == 0 {
		acct.CreditLimit = DefaultCreditLimit
	}
	s.data.Accounts[acct.ID] = acct
	if err := s.writeLedgerFile(); err != nil {
		delete(s.data.Accounts, acct.ID)
		return err
	}
	return nil
}

func (s *LedgerStore) SaveLoan(key string, loan models.LoanData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Loans[key] = loan
	if err := s.writeLedgerFile(); err != nil {
		delete(s.data.Loans, key)
		return err
	}
	return nil
}

// ApplyDebit checks account rules (overdraft / credit limit), updates the balance, and saves the entry.
func (s *LedgerStore) ApplyDebit(entry models.TransactionRecord) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acct, ok := s.data.Accounts[entry.AccountID]
	if !ok {
		return 0, ErrAccountNotFound
	}

	nextBalance := math.Round((acct.Balance-entry.Amount)*100) / 100
	if acct.Type == CreditType {
		limit := acct.CreditLimit
		if limit <= 0 {
			limit = DefaultCreditLimit
		}
		if nextBalance < -limit {
			return acct.Balance, ErrCreditLimitExceeded
		}
	} else if acct.Balance < entry.Amount {
		return acct.Balance, ErrInsufficientFunds
	}

	prevBalance := acct.Balance
	acct.Balance = nextBalance
	s.data.Accounts[entry.AccountID] = acct
	s.data.Transactions[entry.AccountID] = append(s.data.Transactions[entry.AccountID], entry)

	if err := s.writeLedgerFile(); err != nil {
		acct.Balance = prevBalance
		s.data.Accounts[entry.AccountID] = acct
		txs := s.data.Transactions[entry.AccountID]
		s.data.Transactions[entry.AccountID] = txs[:len(txs)-1]
		return prevBalance, err
	}
	return acct.Balance, nil
}

func (s *LedgerStore) ListTransactions(limit int) []models.TransactionRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var all []models.TransactionRecord
	for _, entries := range s.data.Transactions {
		all = append(all, entries...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Date.After(all[j].Date)
	})
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all
}
