package patterns

type InterestStrategy interface {
	CalculateInterest(balance float64) float64
}

type SavingsStrategy struct{}

func (s *SavingsStrategy) CalculateInterest(balance float64) float64 {
	return balance * 0.05 // 5% годовых
}

type LoanStrategy struct{}

func (l *LoanStrategy) CalculateInterest(balance float64) float64 {
	return balance * 0.15 // 15% годовых по кредиту
}

// Контекст
type InterestContext struct {
	strategy InterestStrategy
}

func (c *InterestContext) SetStrategy(s InterestStrategy) { c.strategy = s }
func (c *InterestContext) ExecuteStrategy(balance float64) float64 {
	return c.strategy.CalculateInterest(balance)
}
