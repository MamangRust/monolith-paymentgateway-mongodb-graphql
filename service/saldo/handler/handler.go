package handler

import (
	"time"

	"github.com/MamangRust/monolith-payment-gateway-shared/convert"

	saldostatshandler "github.com/MamangRust/monolith-payment-gateway-saldo/handler/stats"
	"github.com/MamangRust/monolith-payment-gateway-saldo/service"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func formatPtrTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func int32Ptr(v *int64) *int32 {
	if v == nil {
		return nil
	}
	result := int32(*v)
	return &result
}

func formatPtrDeletedAt(d *time.Time) *wrapperspb.StringValue {
	if d == nil {
		return nil
	}
	return convert.TimeToWrappers(d)
}

type handler struct {
	SaldoQueryHandleGrpc
	SaldoCommandHandleGrpc
	saldostatshandler.HandleStats
}

type Handler interface {
	SaldoQueryHandleGrpc
	SaldoCommandHandleGrpc
	saldostatshandler.HandleStats
}

func NewHandler(service service.Service) Handler {
	return &handler{
		SaldoQueryHandleGrpc:   NewSaldoQueryHandleGrpc(service),
		SaldoCommandHandleGrpc: NewSaldoCommandHandleGrpc(service),
		HandleStats:            saldostatshandler.NewSaldoStatsHandle(service),
	}
}
