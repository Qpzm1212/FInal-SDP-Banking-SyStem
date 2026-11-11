package patterns

import (
	"fmt"
	"strings"
)

type Account interface {
	GetID() string
	GetType() string
	GetTitle() string
	CreditLimit() float64
	MaxTransfer() float64
}

type CheckingAccount struct {
	ID string
}

func (c *CheckingAccount) GetID() string        { return c.ID }
func (c *CheckingAccount) GetType() string      { return "checking" }
func (c *CheckingAccount) GetTitle() string     { return "Checking" }
func (c *CheckingAccount) CreditLimit() float64 { return 0 }
func (c *CheckingAccount) MaxTransfer() float64 { return 50000 }

type SavingsAccount struct {
	ID string
}

func (s *SavingsAccount) GetID() string        { return s.ID }
func (s *SavingsAccount) GetType() string      { return "savings" }
func (s *SavingsAccount) GetTitle() string     { return "Savings" }
func (s *SavingsAccount) CreditLimit() float64 { return 0 }
func (s *SavingsAccount) MaxTransfer() float64 { return 25000 } // cap large single withdrawals from savings

type CreditAccount struct {
	ID string
}

func (c *CreditAccount) GetID() string        { return c.ID }
func (c *CreditAccount) GetType() string      { return "credit" }
func (c *CreditAccount) GetTitle() string     { return "Credit" }
func (c *CreditAccount) CreditLimit() float64 { return 5000 }
func (c *CreditAccount) MaxTransfer() float64 { return 5000 }

type AccountCreator interface {
	CreateAccount(id string) Account
}

type CheckingCreator struct{}

func (CheckingCreator) CreateAccount(id string) Account {
	return &CheckingAccount{ID: id}
}

type SavingsCreator struct{}

func (SavingsCreator) CreateAccount(id string) Account {
	return &SavingsAccount{ID: id}
}

type CreditCreator struct{}

func (CreditCreator) CreateAccount(id string) Account {
	return &CreditAccount{ID: id}
}

var accountRegistry = map[string]AccountCreator{
	"checking": CheckingCreator{},
	"savings":  SavingsCreator{},
	"credit":   CreditCreator{},
}

func validAccountID(id string) bool {
	if len(id) < 2 || len(id) > 24 {
		return false
	}
	for _, r := range id {
		isAlphaNum := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if !isAlphaNum && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func CreateAccount(accType string, id string) (Account, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: account number is required", ErrInvalidInput)
	}
	if !validAccountID(id) {
		return nil, fmt.Errorf("%w: account number must be 2-24 chars (letters, digits, - or _)", ErrInvalidInput)
	}
	creator, ok := accountRegistry[accType]
	if !ok {
		return nil, fmt.Errorf("%w: unknown account type %q", ErrInvalidInput, accType)
	}
	return creator.CreateAccount(id), nil
}
