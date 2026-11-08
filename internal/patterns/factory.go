package patterns

import "errors"

// 1. Общий интерфейс продукта (Счет)
type Account interface {
	GetID() string
	GetType() string
}

// 2. Конкретный продукт: Расчетный счет
type CheckingAccount struct {
	ID string
}

func (c *CheckingAccount) GetID() string   { return c.ID }
func (c *CheckingAccount) GetType() string { return "Расчетный (Checking)" }

// 3. Конкретный продукт: Сберегательный счет
type SavingsAccount struct {
	ID string
}

func (s *SavingsAccount) GetID() string   { return s.ID }
func (s *SavingsAccount) GetType() string { return "Сберегательный (Savings)" }

// 4. Сама Фабрика
func CreateAccount(accType string, id string) (Account, error) {
	switch accType {
	case "checking":
		// Здесь в будущем можно добавить логику (например, вызов Singleton БД для проверки уникальности)
		return &CheckingAccount{ID: id}, nil
	case "savings":
		return &SavingsAccount{ID: id}, nil
	default:
		return nil, errors.New("неизвестный тип счета")
	}
}
