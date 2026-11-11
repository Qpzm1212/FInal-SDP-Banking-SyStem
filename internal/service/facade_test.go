package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "bank-test")
	os.Setenv("BANK_DB_PATH", filepath.Join(dir, "db.json"))
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestTransferFlow(t *testing.T) {
	f := NewBankFacade()

	res, err := f.Transfer("ACC-1", 500)
	if err != nil || res.Suspicious || res.Balance != 47750.75 {
		t.Fatalf("regular transfer: %+v err=%v", res, err)
	}
	res, err = f.Transfer("ACC-1", 15000)
	if err != nil || !res.Suspicious || res.Reason == "" {
		t.Fatalf("large transfer must be flagged: %+v err=%v", res, err)
	}
	if _, err := f.Transfer("ACC-2", 1e9); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("want ErrInsufficientFunds, got %v", err)
	}
	if _, err := f.Transfer("NOPE", 10); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("want ErrAccountNotFound, got %v", err)
	}
	if _, err := f.Transfer("ACC-1", -5); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	if got := f.ListTransactions(100); len(got) < 7 || !got[0].Suspicious {
		t.Fatalf("history must be newest first and include the flag, got %d records", len(got))
	}
}

func TestAccountsAndLoansAndInterest(t *testing.T) {
	f := NewBankFacade()

	acc, err := f.OpenAccount("credit", " CR-9 ")
	if err != nil || acc.ID != "CR-9" || acc.Type != "credit" || acc.CreditLimit != 5000 {
		t.Fatalf("open: %+v err=%v", acc, err)
	}
	if _, err := f.OpenAccount("credit", "CR-9"); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("duplicate: %v", err)
	}
	if res, err := f.Transfer("CR-9", 300); err != nil || res.Balance != -300 {
		t.Fatalf("credit transfer: %+v err=%v", res, err)
	}
	if _, err := f.Transfer("CR-9", 4900); !errors.Is(err, ErrCreditLimitExceeded) {
		t.Fatalf("want ErrCreditLimitExceeded, got %v", err)
	}

	loan, err := f.ApplyForLoan("Ali", 5000)
	if err != nil || loan.Monthly <= 0 {
		t.Fatalf("loan: %+v err=%v", loan, err)
	}
	if _, err := f.ApplyForLoan("", 5000); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("loan validation: %v", err)
	}

	if v, err := f.CalculateYield("savings", 1000); err != nil || v != 50 {
		t.Fatalf("savings: %v %v", v, err)
	}
	if _, err := f.CalculateYield("bogus", 1000); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("unknown type: %v", err)
	}
}
