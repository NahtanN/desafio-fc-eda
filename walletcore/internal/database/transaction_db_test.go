package database

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/nahtann/walletcore/internal/entity"
)

type TransactionDBTestSuite struct {
	suite.Suite
	client1       *entity.Client
	client2       *entity.Client
	accountFrom   *entity.Account
	accountTo     *entity.Account
	transactionDB *TransactionDB
}

func (s *TransactionDBTestSuite) SetupSuite() {
	db, err := sql.Open("sqlite3", ":memory:")
	s.Nil(err)
	db.Exec("CREATE TABLE clients (id TEXT PRIMARY KEY, name TEXT, email TEXT, created_at date)")
	db.Exec(
		"CREATE TABLE accounts (id TEXT PRIMARY KEY, client_id TEXT, balance REAL, created_at date)",
	)
	db.Exec(
		"CREATE TABLE transactions (id TEXT PRIMARY KEY, account_id_from TEXT, account_id_to TEXT, amount REAL, created_at date)",
	)

	s.client1, _ = entity.NewClient("John Doe", "e@e.com")
	s.client2, _ = entity.NewClient("Jane Doe", "d@d.com")
	s.accountFrom = entity.NewAccount(s.client1)
	s.accountFrom.Balance = 1000
	s.accountTo = entity.NewAccount(s.client2)
	s.accountTo.Balance = 1000
	s.transactionDB = NewTransactionDB(db)

	s.transactionDB = NewTransactionDB(db)
}

func (s *TransactionDBTestSuite) TearDownSuite() {
	defer s.transactionDB.DB.Close()
	s.transactionDB.DB.Exec("DROP TABLE transactions")
	s.transactionDB.DB.Exec("DROP TABLE accounts")
	s.transactionDB.DB.Exec("DROP TABLE clients")
}

func TestTransactionDBTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionDBTestSuite))
}

func (s *TransactionDBTestSuite) TestCreate() {
	transaction, err := entity.NewTransaction(s.accountFrom, s.accountTo, 100)
	s.Nil(err)

	err = s.transactionDB.Create(transaction)
	s.Nil(err)
}
