package db

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"ESHKEREE/internal/models"
)

func openTempLedger(t *testing.T) *LedgerStore {
	t.Helper()
	return openLedger(filepath.Join(t.TempDir(), "db.json"))
}

func TestSeedAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	store := openLedger(path)
	if len(store.ListAccounts()) != 2 {
		t.Fatalf("expected 2 starter accounts, got %d", len(store.ListAccounts()))
	}
	if err := store.SaveAccount(models.AccountData{ID: "X-1", Type: "checking", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	reopened := openLedger(path)
	if _, ok := reopened.GetAccount("X-1"); !ok {
		t.Fatal("account missing after reopening ledger")
	}
}

func TestDuplicateAccount(t *testing.T) {
	store := openTempLedger(t)
	err := store.SaveAccount(models.AccountData{ID: "ACC-1", Type: "checking"})
	if !errors.Is(err, ErrAccountExists) {
		t.Fatalf("expected ErrAccountExists, got %v", err)
	}
}

func TestDebitRulesAndCreditLimit(t *testing.T) {
	store := openTempLedger(t)
	tx := func(id string, amt float64) models.TransactionRecord {
		return models.TransactionRecord{AccountID: id, Name: "Transfer", Amount: amt, Date: time.Now()}
	}

	bal, err := store.ApplyDebit(tx("ACC-2", 400))
	if err != nil || bal != 12000 {
		t.Fatalf("debit failed: bal=%v err=%v", bal, err)
	}
	if _, err := store.ApplyDebit(tx("ACC-2", 999999)); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
	if _, err := store.ApplyDebit(tx("NOPE", 1)); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}

	_ = store.SaveAccount(models.AccountData{ID: "CR-1", Type: CreditType, CreditLimit: 5000, CreatedAt: time.Now()})
	bal, err = store.ApplyDebit(tx("CR-1", 250))
	if err != nil || bal != -250 {
		t.Fatalf("credit account should allow negative balance within limit: bal=%v err=%v", bal, err)
	}
	if _, err := store.ApplyDebit(tx("CR-1", 4900)); !errors.Is(err, ErrCreditLimitExceeded) {
		t.Fatalf("expected ErrCreditLimitExceeded past -$5000, got %v", err)
	}
}

func TestConcurrentDebits(t *testing.T) {
	store := openTempLedger(t)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.ApplyDebit(models.TransactionRecord{
				AccountID: "ACC-1",
				Name:      "Transfer",
				Amount:    10,
				Date:      time.Now(),
			})
		}()
	}
	wg.Wait()
	acc, _ := store.GetAccount("ACC-1")
	if acc.Balance != 48250.75-500 {
		t.Fatalf("balance after 50 concurrent debits: %v", acc.Balance)
	}
}
