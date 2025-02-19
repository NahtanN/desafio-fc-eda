package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {
	client1, _ := NewClient("John Doe", "a@c.com")
	account1 := NewAccount(client1)

	client2, _ := NewClient("Jane Doe", "d@e.com")
	account2 := NewAccount(client2)

	account1.Credit(1000)
	account2.Credit(1000)

	transaction, err := NewTransaction(account1, account2, 100)
	assert.Nil(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, account1.Balance, 900.0)
	assert.Equal(t, account2.Balance, 1100.0)
}

func TestCreateTransactionWithInsufficientBalance(t *testing.T) {
	client1, _ := NewClient("John Doe", "a@c.com")
	account1 := NewAccount(client1)

	client2, _ := NewClient("Jane Doe", "d@e.com")
	account2 := NewAccount(client2)

	account1.Credit(1000)
	account2.Credit(1000)

	transaction, err := NewTransaction(account1, account2, 1001)
	assert.NotNil(t, err)
	assert.Error(t, err, ErrInsufficientBalance)
	assert.Nil(t, transaction)
	assert.Equal(t, account1.Balance, 1000.0)
	assert.Equal(t, account2.Balance, 1000.0)
}
