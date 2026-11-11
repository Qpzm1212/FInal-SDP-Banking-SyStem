package patterns

import (
	"fmt"
	"math"
)

const (
	CheckingRatePercent     = 0.0
	SavingsRatePercent      = 5.0
	StandardLoanRatePercent = 15.0 // kept in one place so LoanBuilder and LoanStrategy stay aligned
)

func validAmount(v float64) bool {
	return v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0)
}

type InterestStrategy interface {
	CalculateInterest(balance float64) float64
}

type CheckingStrategy struct{}

func (c *CheckingStrategy) CalculateInterest(balance float64) float64 {
	return balance * CheckingRatePercent / 100
}

type SavingsStrategy struct{}

func (s *SavingsStrategy) CalculateInterest(balance float64) float64 {
	return balance * SavingsRatePercent / 100
}

type LoanStrategy struct{}

func (l *LoanStrategy) CalculateInterest(balance float64) float64 {
	return balance * StandardLoanRatePercent / 100
}

type InterestContext struct {
	strategy InterestStrategy
}

func (c *InterestContext) SetStrategy(s InterestStrategy) {
	c.strategy = s
}

func (c *InterestContext) ExecuteStrategy(balance float64) float64 {
	if c.strategy == nil {
		return 0
	}
	return c.strategy.CalculateInterest(balance)
}

var interestStrategies = map[string]InterestStrategy{
	"checking": &CheckingStrategy{},
	"savings":  &SavingsStrategy{},
	"credit":   &LoanStrategy{},
	"loan":     &LoanStrategy{},
}

func GetStrategy(name string) (InterestStrategy, error) {
	s, ok := interestStrategies[name]
	if !ok {
		return nil, fmt.Errorf("%w: unknown calculation type %q (available: checking, savings, credit)", ErrInvalidInput, name)
	}
	return s, nil
}
