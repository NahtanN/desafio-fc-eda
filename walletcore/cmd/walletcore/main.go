package main

import (
	"context"
	"database/sql"
	"fmt"

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
