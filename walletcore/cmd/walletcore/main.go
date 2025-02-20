package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/go-sql-driver/mysql"

	"github.com/nahtann/walletcore/internal/database"
	"github.com/nahtann/walletcore/internal/event"
	"github.com/nahtann/walletcore/internal/event/handler"
	createaccount "github.com/nahtann/walletcore/internal/usecases/create_account"
	createclient "github.com/nahtann/walletcore/internal/usecases/create_client"
	createtransaction "github.com/nahtann/walletcore/internal/usecases/create_transaction"
	"github.com/nahtann/walletcore/internal/web"
	"github.com/nahtann/walletcore/internal/web/webserver"
	"github.com/nahtann/walletcore/pkg/events"
	"github.com/nahtann/walletcore/pkg/kafka"
	"github.com/nahtann/walletcore/pkg/uow"
)

func main() {
	// Mysql connection on localhost
	db, err := sql.Open("mysql", "root:root@tcp(mysql:3306)/wallet?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	migrations(db)
	seed(db)

	configMap := ckafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"group.id":          "wallet",
	}
	kafkaProducer := kafka.NewKafkaProducer(&configMap)

	eventDispatcher := events.NewEventDispatcher()
	transactionCreatedEvent := event.NewTransactionCreated()
	eventDispatcher.Register(
		"TransactionCreated",
		handler.NewTransactionCreatedKafkaHandler(kafkaProducer),
	)

	balanceUpdatedEvent := event.NewBalanceUpdated()
	eventDispatcher.Register(
		"BalanceUpdated",
		handler.NewBalanceUpdatedKafkaHandler(kafkaProducer),
	)

	clientDb := database.NewClientDB(db)
	accountDb := database.NewAccountDB(db)
	/*transactionDb := database.NewTransactionDB(db)*/

	ctx := context.Background()
	uow := uow.NewUow(ctx, db)

	uow.Register("AccountDB", func(tx *sql.Tx) interface{} {
		return database.NewAccountDB(db)
	})
	uow.Register("TransactionDB", func(tx *sql.Tx) interface{} {
		return database.NewTransactionDB(db)
	})

	createClientUsecase := createclient.NewCreateClientUseCase(clientDb)
	createAccountUseCase := createaccount.NewCreateAccountUseCase(accountDb, clientDb)
	createTransactionUsecase := createtransaction.NewTransactionUseCase(
		uow,
		eventDispatcher,
		transactionCreatedEvent,
		balanceUpdatedEvent,
	)

	webserver := webserver.NewWebServer(":8080")

	clientHandler := web.NewWebClientHandler(*createClientUsecase)
	accountHandler := web.NewWebAccountHandler(*createAccountUseCase)
	transactionHandler := web.NewWebTransactionHandler(*createTransactionUsecase)

	webserver.AddHandler("/clients", clientHandler.CreateClient)
	webserver.AddHandler("/accounts", accountHandler.CreateAccount)
	webserver.AddHandler("/transactions", transactionHandler.CreateTransaction)

	fmt.Println("Server started on port 8080")
	webserver.Start()
}

func migrations(db *sql.DB) {
	db.Exec(
		"CREATE TABLE IF NOT EXISTS clients (id VARCHAR(255) PRIMARY KEY, name TEXT, email TEXT, created_at date)",
	)
	db.Exec(
		"CREATE TABLE IF NOT EXISTS accounts (id VARCHAR(255) PRIMARY KEY, client_id TEXT, balance REAL, created_at date)",
	)
	db.Exec(
		"CREATE TABLE IF NOT EXISTS transactions (id VARCHAR(255) PRIMARY KEY, account_id_from TEXT, account_id_to TEXT, amount REAL, created_at date)",
	)

	fmt.Println("Migrations executed successfully")
}

func seed(db *sql.DB) {
	clientId1 := "d77d5d9b-6959-4637-8aaf-4c677e2fa83e"
	db.Exec(
		"INSERT INTO clients (id, name, email, created_at) VALUES (?, ?, ?, ?)",
		clientId1,
		"foo",
		"foo@bar",
		time.Now(),
	)
	clientBalanceId1 := "1fe35ef4-bbc7-4a23-80c4-48c966dbbc5f"
	db.Exec(
		"INSERT INTO accounts (id, client_id, balance, created_at) VALUES (?, ?, ?, ?)",
		clientBalanceId1,
		clientId1,
		1000,
		time.Now(),
	)

	clientId2 := "611d3cef-f1c6-4fa7-a821-0b5edec151d6"
	db.Exec(
		"INSERT INTO clients (id, name, email, created_at) VALUES (?, ?, ?, ?)",
		clientId2,
		"bar",
		"bar@foo",
		time.Now(),
	)
	clientBalanceId2 := "4be4234a-156a-4679-8a8a-35d8f1293502"
	db.Exec(
		"INSERT INTO accounts (id, client_id, balance, created_at) VALUES (?, ?, ?, ?)",
		clientBalanceId2,
		clientId2,
		1000,
		time.Now(),
	)

	fmt.Println("Seeds executed successfully")
}
