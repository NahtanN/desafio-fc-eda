package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/go-sql-driver/mysql"

	"github.com/nahtann/balance-service/pkg/kafka"
)

type BalancePayload struct {
	AccountIDFrom        string  `json:"account_id_from"`
	AccountIDTo          string  `json:"account_id_to"`
	BalanceAccountIDFrom float64 `json:"balance_account_id_from"`
	BalanceAccountIDTo   float64 `json:"balance_account_id_to"`
}

type BalanceMessage struct {
	Name    string         `json:"Name"`
	Payload BalancePayload `json:"Payload"`
}

type AccountNotFound struct {
	Message string `json:"message"`
}

type AccountBalance struct {
	AccountID string  `json:"account_id"`
	Balance   float64 `json:"balance"`
}

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(balance-service-db:3306)/balance?parseTime=true")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	migrations(db)

	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"group.id":          "wallet",
	}

	consumer := kafka.NewConsumer(configMap, []string{"balances"})
	msgChan := make(chan *ckafka.Message)

	go startConsumer(consumer, msgChan)
	go handleMessages(db, msgChan)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /balance/{account_id}", func(w http.ResponseWriter, r *http.Request) {
		accountID := r.PathValue("account_id")

		exists := accountExists(db, accountID)

		w.Header().Set("Content-Type", "application/json")
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(AccountNotFound{Message: "Account not found"})
			return
		}

		balance := getBalance(db, accountID)
		json.NewEncoder(w).Encode(AccountBalance{AccountID: accountID, Balance: balance})
	})

	http.ListenAndServe(":3003", mux)
}

func getBalance(db *sql.DB, accountID string) float64 {
	var balance float64
	err := db.QueryRow("SELECT balance FROM balances WHERE id = ?", accountID).Scan(&balance)
	if err != nil {
		fmt.Println(err.Error())
	}

	return balance
}

func startConsumer(consumer *kafka.Consumer, msgChan chan *ckafka.Message) {
	if err := consumer.Consume(msgChan); err != nil {
		panic(err)
	}
}

func handleMessages(db *sql.DB, msgChan chan *ckafka.Message) {
	for msg := range msgChan {
		fmt.Println(string(msg.Value))

		balancePayload := BalanceMessage{}
		err := json.Unmarshal(msg.Value, &balancePayload)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(balancePayload)
		if balancePayload.Payload.AccountIDFrom != "" {
			updateBalance(
				db,
				balancePayload.Payload.AccountIDFrom,
				balancePayload.Payload.BalanceAccountIDFrom,
			)
		}

		if balancePayload.Payload.AccountIDTo != "" {
			updateBalance(
				db,
				balancePayload.Payload.AccountIDTo,
				balancePayload.Payload.BalanceAccountIDTo,
			)
		}
	}
}

func updateBalance(db *sql.DB, accountID string, balance float64) {
	exists := accountExists(db, accountID)

	if !exists {
		_, err := db.Exec(
			"INSERT INTO balances (id, balance, created_at) VALUES (?, ?, NOW())",
			accountID,
			balance,
		)
		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	_, err := db.Exec(
		"UPDATE balances SET balance = ? WHERE id = ?",
		balance,
		accountID,
	)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func accountExists(db *sql.DB, accountID string) bool {
	exists := false
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM balances WHERE id = ?)", accountID).
		Scan(&exists)
	if err != nil {
		fmt.Println(err.Error())
	}

	return exists
}

func migrations(db *sql.DB) {
	db.Exec(
		"CREATE TABLE IF NOT EXISTS balances (id VARCHAR(255) PRIMARY KEY, balance REAL, created_at date)",
	)

	fmt.Println("Migrations executed successfully")
}
