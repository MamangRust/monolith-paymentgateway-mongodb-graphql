package handler

import (
	"time"

	"github.com/MamangRust/monolith-payment-gateway-shared/convert"

	transactionstatshandler "github.com/MamangRust/monolith-payment-gateway-transaction/handler/stats"
	"github.com/MamangRust/monolith-payment-gateway-transaction/service"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func formatTimeWrapper(t *time.Time) *wrapperspb.StringValue {
	if t == nil {
		return nil
	}
	return convert.TimeToWrappers(t)
}

type Handler interface {
	TransactionQueryHandleGrpc
	TransactionCommandHandleGrpc
	transactionstatshandler.HandleStats
}

type handler struct {
	TransactionQueryHandleGrpc
	TransactionCommandHandleGrpc
	transactionstatshandler.HandleStats
}

func NewHandler(service service.Service) Handler {
	return &handler{
		TransactionQueryHandleGrpc:   NewTransactionQueryHandleGrpc(service),
		TransactionCommandHandleGrpc: NewTransactionCommandHandleGrpc(service),
		HandleStats:                  transactionstatshandler.NewTransactionStatsHandleGrpc(service),
	}
}
