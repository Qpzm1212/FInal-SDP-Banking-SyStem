package patterns

import (
	"errors"
	"testing"
)

func TestStrategySelection(t *testing.T) {
	s, err := GetStrategy("savings")
	if err != nil || s.CalculateInterest(1000) != 50 {
		t.Fatalf("savings: want 50, got err=%v", err)
	}
	s, _ = GetStrategy("loan")
	if s.CalculateInterest(1000) != 150 {
		t.Fatal("loan: want 150")
	}
	if _, err := GetStrategy("savngs"); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("unknown type must return ErrInvalidInput")
	}
}

func TestBuilderValidation(t *testing.T) {
	loan, err := LoanDirector{}.ConstructStandardLoan(NewStandardLoanBuilder(), "Ali", 5000)
	if err != nil || loan.Rate != StandardLoanRatePercent || loan.Months != 12 {
		t.Fatalf("bad loan: %+v err=%v", loan, err)
	}
	if _, err := NewStandardLoanBuilder().SetBorrower("").SetAmount(100).SetTerms().Build(); err == nil {
		t.Fatal("empty borrower must fail")
	}
	if _, err := NewStandardLoanBuilder().SetBorrower("Ali").SetAmount(-1).SetTerms().Build(); err == nil {
		t.Fatal("negative amount must fail")
	}
}

func TestFactoryMethod(t *testing.T) {
	a, err := CreateAccount("savings", "A1")
	if err != nil || a.GetType() != "savings" {
		t.Fatalf("savings: %v %v", a, err)
	}
	if _, err := CreateAccount("gold", "A2"); err == nil {
		t.Fatal("unknown type must fail")
	}
	if _, err := CreateAccount("checking", " "); err == nil {
		t.Fatal("empty id must fail")
	}
}

type spy struct{ got []*TransactionEvent }

func (s *spy) Update(e *TransactionEvent) { s.got = append(s.got, e) }

func TestObserverNotifiesAll(t *testing.T) {
	m := NewTransactionMonitor()
	s := &spy{}
	m.AddObserver(&FraudDetector{Threshold: FraudThreshold})
	m.AddObserver(s)

	if e := m.ProcessTransaction(500, "A1"); e.Suspicious {
		t.Fatal("small amount must not be suspicious")
	}
	if e := m.ProcessTransaction(20000, "A1"); !e.Suspicious {
		t.Fatal("large amount must be suspicious")
	}
	if len(s.got) != 2 {
		t.Fatalf("observer must get every event, got %d", len(s.got))
	}

	m.RemoveObserver(s)
	m.ProcessTransaction(1, "A1")
	if len(s.got) != 2 {
		t.Fatal("removed observer must not be notified")
	}
}
