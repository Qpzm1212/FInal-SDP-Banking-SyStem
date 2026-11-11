package patterns

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	MinLoanAmount     = 100.0
	MaxLoanAmount     = 500000.0
	DefaultLoanMonths = 12
)

type LoanAgreement struct {
	Number   string    `json:"number"`
	Borrower string    `json:"borrower"`
	Amount   float64   `json:"amount"`
	Rate     float64   `json:"rate"`
	Months   int       `json:"months"`
	Monthly  float64   `json:"monthly"`
	Total    float64   `json:"total"`
	Date     time.Time `json:"date"`
}

type LoanBuilder interface {
	SetBorrower(name string) LoanBuilder
	SetAmount(amount float64) LoanBuilder
	SetTerms() LoanBuilder
	SetCustomTerms(rate float64, months int) LoanBuilder
	Build() (*LoanAgreement, error)
}

type StandardLoanBuilder struct {
	draft *LoanAgreement
}

func NewStandardLoanBuilder() *StandardLoanBuilder {
	return &StandardLoanBuilder{draft: &LoanAgreement{}}
}

func (b *StandardLoanBuilder) SetBorrower(name string) LoanBuilder {
	b.draft.Borrower = strings.TrimSpace(name)
	return b
}

func (b *StandardLoanBuilder) SetAmount(amount float64) LoanBuilder {
	b.draft.Amount = amount
	return b
}

func (b *StandardLoanBuilder) SetTerms() LoanBuilder {
	b.draft.Rate = StandardLoanRatePercent
	b.draft.Months = DefaultLoanMonths
	return b
}

func (b *StandardLoanBuilder) SetCustomTerms(rate float64, months int) LoanBuilder {
	b.draft.Rate = rate
	b.draft.Months = months
	return b
}

func (b *StandardLoanBuilder) Build() (*LoanAgreement, error) {
	agreement := *b.draft

	if agreement.Borrower == "" {
		return nil, fmt.Errorf("%w: borrower name is required", ErrInvalidInput)
	}
	if len([]rune(agreement.Borrower)) < 2 {
		return nil, fmt.Errorf("%w: borrower name is too short", ErrInvalidInput)
	}
	if !validAmount(agreement.Amount) {
		return nil, fmt.Errorf("%w: loan amount must be greater than 0", ErrInvalidInput)
	}
	if agreement.Amount < MinLoanAmount || agreement.Amount > MaxLoanAmount {
		return nil, fmt.Errorf("%w: loan amount must be between $%.0f and $%.0f", ErrInvalidInput, MinLoanAmount, MaxLoanAmount)
	}
	if agreement.Rate <= 0 || agreement.Months <= 0 {
		return nil, fmt.Errorf("%w: loan terms are not set", ErrInvalidInput)
	}
	if agreement.Rate > 45 || agreement.Months < 3 || agreement.Months > 120 {
		return nil, fmt.Errorf("%w: loan terms out of allowed policy range", ErrInvalidInput)
	}

	// monthly annuity formula: P * r / (1 - (1+r)^-n)
	monthlyRate := agreement.Rate / 100 / 12
	payment := agreement.Amount * monthlyRate / (1 - math.Pow(1+monthlyRate, -float64(agreement.Months)))
	agreement.Monthly = math.Round(payment*100) / 100
	agreement.Total = math.Round(payment*float64(agreement.Months)*100) / 100
	agreement.Date = time.Now()
	agreement.Number = fmt.Sprintf("LN-%06d", agreement.Date.UnixNano()/1000%1000000)

	return &agreement, nil
}

type LoanDirector struct{}

func (LoanDirector) ConstructStandardLoan(b LoanBuilder, borrower string, amount float64) (*LoanAgreement, error) {
	return b.SetBorrower(borrower).SetAmount(amount).SetTerms().Build()
}
