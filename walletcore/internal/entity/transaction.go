package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAccountFromIsRequired       = errors.New("account from is required")
	ErrAccountToIsRequired         = errors.New("account to is required")
	ErrAmountMustBeGreaterThanZero = errors.New("amount must be greater than zero")
	ErrInsufficientBalance         = errors.New("insufficient balance")
)

type Transaction struct {
	ID          string    `json:"id"`
	AccountFrom *Account  `json:"account_from"`
	AccountTo   *Account  `json:"account_to"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewTransaction(accountFrom, accountTo *Account, amount float64) (*Transaction, error) {
	transaction := &Transaction{
		ID:          uuid.New().String(),
		AccountFrom: accountFrom,
		AccountTo:   accountTo,
		Amount:      amount,
		CreatedAt:   time.Now(),
	}

	err := transaction.Validate()
	if err != nil {
		return nil, err
	}

	transaction.Commit()

	return transaction, nil
}

func (t *Transaction) Commit() {
	t.AccountFrom.Debit(t.Amount)
	t.AccountTo.Credit(t.Amount)
}

func (t *Transaction) Validate() error {
	if t.AccountFrom.ID == "" {
		return ErrAccountFromIsRequired
	}

	if t.AccountTo.ID == "" {
		return ErrAccountToIsRequired
	}

	if t.Amount <= 0 {
		return ErrAmountMustBeGreaterThanZero
	}

	if t.AccountFrom.Balance < t.Amount {
		return ErrInsufficientBalance
	}

	return nil
}

// create custom errors
