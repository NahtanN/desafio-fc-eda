package web

import (
	"encoding/json"
	"net/http"

	createtransaction "github.com/nahtann/walletcore/internal/usecases/create_transaction"
)

type WebTransactionHandler struct {
	TransactionUseCase createtransaction.CreateTransactionUseCase
}

func NewWebTransactionHandler(
	transactionUseCase createtransaction.CreateTransactionUseCase,
) *WebTransactionHandler {
	return &WebTransactionHandler{
		TransactionUseCase: transactionUseCase,
	}
}

func (h *WebTransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var dto createtransaction.CreateTransactionInputDTO

	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	output, err := h.TransactionUseCase.Execute(ctx, dto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(output)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
