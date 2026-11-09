package patterns

import (
	"fmt"
	"strings"
)

type LoanAgreement struct {
	Borrower string
	Amount   float64
	Rate     float64
	Months   int
}

type LoanBuilder interface {
	SetBorrower(name string) LoanBuilder
	SetAmount(amount float64) LoanBuilder
	SetTerms() LoanBuilder
	SetCustomTerms(rate float64, months int) LoanBuilder
	Build() (*LoanAgreement, error)
}

type StandardLoanBuilder struct {
	agreement *LoanAgreement
}

func NewStandardLoanBuilder() *StandardLoanBuilder {
	return &StandardLoanBuilder{agreement: &LoanAgreement{}}
}

func (b *StandardLoanBuilder) SetBorrower(name string) LoanBuilder {
	b.agreement.Borrower = strings.TrimSpace(name)
	return b
}

func (b *StandardLoanBuilder) SetAmount(amount float64) LoanBuilder {
	b.agreement.Amount = amount
	return b
}

// SetTerms - стандартные условия (ставка берется из той же константы, что и в Strategy).
func (b *StandardLoanBuilder) SetTerms() LoanBuilder {
	b.agreement.Rate = StandardLoanRatePercent
	b.agreement.Months = 12
	return b
}

func (b *StandardLoanBuilder) SetCustomTerms(rate float64, months int) LoanBuilder {
	b.agreement.Rate = rate
	b.agreement.Months = months
	return b
}

// Build проверяет готовность договора и возвращает его.
func (b *StandardLoanBuilder) Build() (*LoanAgreement, error) {
	a := b.agreement
	switch {
	case a.Borrower == "":
		return nil, fmt.Errorf("%w: не указан заемщик", ErrInvalidInput)
	case !validAmount(a.Amount):
		return nil, fmt.Errorf("%w: сумма кредита должна быть больше 0", ErrInvalidInput)
	case a.Rate <= 0 || a.Months <= 0:
		return nil, fmt.Errorf("%w: не заданы условия кредита", ErrInvalidInput)
	}
	result := *a // отдаем копию
	return &result, nil
}

// LoanDirector - Director: знает порядок шагов построения.
type LoanDirector struct{}

func (LoanDirector) ConstructStandardLoan(b LoanBuilder, borrower string, amount float64) (*LoanAgreement, error) {
	return b.SetBorrower(borrower).SetAmount(amount).SetTerms().Build()
}
