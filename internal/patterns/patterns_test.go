package patterns

import (
	"errors"
	"math"
	"testing"
)

func TestStrategySelection(t *testing.T) {
	s, err := GetStrategy("checking")
	if err != nil || s.CalculateInterest(1000) != 0 {
		t.Fatalf("checking: expected 0, got err=%v", err)
	}
	s, err = GetStrategy("savings")
	if err != nil || s.CalculateInterest(1000) != 50 {
		t.Fatalf("savings: expected 50, got err=%v", err)
	}
	for _, name := range []string{"credit", "loan"} {
		s, err = GetStrategy(name)
		if err != nil || s.CalculateInterest(1000) != 150 {
			t.Fatalf("%s: expected 150, got err=%v", name, err)
		}
	}
	if _, err := GetStrategy("savngs"); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("unknown type must return ErrInvalidInput")
	}
}

func TestBuilderValidationAndPayments(t *testing.T) {
	loan, err := LoanDirector{}.ConstructStandardLoan(NewStandardLoanBuilder(), "Ali", 5000)
	if err != nil || loan.Rate != StandardLoanRatePercent || loan.Months != DefaultLoanMonths {
		t.Fatalf("unexpected loan: %+v err=%v", loan, err)
	}
	if math.Abs(loan.Monthly-451.29) > 0.02 {
		t.Fatalf("monthly payment: got %.2f", loan.Monthly)
	}
	if math.Abs(loan.Total-loan.Monthly*12) > 0.1 || loan.Number == "" {
		t.Fatalf("bad total or number: %+v", loan)
	}
	if _, err := NewStandardLoanBuilder().SetBorrower("").SetAmount(500).SetTerms().Build(); err == nil {
		t.Fatal("empty borrower must fail")
	}
	if _, err := NewStandardLoanBuilder().SetBorrower("A").SetAmount(500).SetTerms().Build(); err == nil {
		t.Fatal("1-char borrower must fail")
	}
	if _, err := NewStandardLoanBuilder().SetBorrower("Ali").SetAmount(50).SetTerms().Build(); err == nil {
		t.Fatal("below-minimum loan must fail")
	}
	if _, err := NewStandardLoanBuilder().SetBorrower("Ali").SetAmount(900000).SetTerms().Build(); err == nil {
		t.Fatal("above-maximum loan must fail")
	}
}

func TestFactoryMethod(t *testing.T) {
	for _, typ := range []string{"checking", "savings", "credit"} {
		a, err := CreateAccount(typ, "A1")
		if err != nil || a.GetType() != typ {
			t.Fatalf("%s: %v %v", typ, a, err)
		}
	}
	if _, err := CreateAccount("gold", "A2"); err == nil {
		t.Fatal("unknown type must fail")
	}
	if _, err := CreateAccount("checking", " "); err == nil {
		t.Fatal("empty id must fail")
	}
	if _, err := CreateAccount("checking", "bad id!"); err == nil {
		t.Fatal("invalid id format must fail")
	}
}

type spyObserver struct {
	got []*TransactionEvent
}

func (s *spyObserver) Update(e *TransactionEvent) {
	s.got = append(s.got, e)
}

func TestObserverNotifiesAndDetectsVelocity(t *testing.T) {
	m := NewTransactionMonitor()
	spy := &spyObserver{}
	m.AddObserver(&FraudDetector{Threshold: FraudThreshold})
	m.AddObserver(spy)

	if e := m.ProcessTransaction(500, "A1"); e.Suspicious {
		t.Fatal("small amount must not be suspicious")
	}
	if e := m.ProcessTransaction(20000, "A1"); !e.Suspicious || e.Reason == "" {
		t.Fatal("large amount must be flagged with reason")
	}
	// Trigger velocity rule on a separate account with small transfers (>3 inside 1 min)
	for i := 0; i < 3; i++ {
		if e := m.ProcessTransaction(100, "VEL-1"); e.Suspicious {
			t.Fatalf("transfer %d should not trigger velocity yet", i+1)
		}
	}
	if e := m.ProcessTransaction(100, "VEL-1"); !e.Suspicious {
		t.Fatal("4th rapid transfer must trigger velocity check")
	}

	prevCount := len(spy.got)
	m.RemoveObserver(spy)
	m.ProcessTransaction(1, "A1")
	if len(spy.got) != prevCount {
		t.Fatal("removed observer must not be notified")
	}
}
