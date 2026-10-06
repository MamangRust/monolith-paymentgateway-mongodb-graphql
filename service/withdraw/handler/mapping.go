package handler

import (
	"time"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatTimeWrapper(t *time.Time) *wrapperspb.StringValue {
	if t == nil || t.IsZero() {
		return nil
	}
	return wrapperspb.String(t.Format(time.RFC3339))
}

func toWithdrawResponse(w *salemongo.Withdraw) *pb.WithdrawResponse {
	if w == nil {
		return nil
	}

	return &pb.WithdrawResponse{
		WithdrawId:     w.WithdrawID,
		WithdrawNo:     w.WithdrawNo.String(),
		CardNumber:     w.CardNumber,
		WithdrawAmount: int32(w.WithdrawAmount),
		WithdrawTime:   w.WithdrawTime.Format(time.RFC3339),
		CreatedAt:      formatTime(w.CreatedAt),
		UpdatedAt:      formatTime(w.UpdatedAt),
	}
}

func toWithdrawDeleteAtResponse(w *salemongo.Withdraw) *pb.WithdrawResponseDeleteAt {
	if w == nil {
		return nil
	}

	return &pb.WithdrawResponseDeleteAt{
		WithdrawId:     w.WithdrawID,
		WithdrawNo:     w.WithdrawNo.String(),
		CardNumber:     w.CardNumber,
		WithdrawAmount: int32(w.WithdrawAmount),
		WithdrawTime:   w.WithdrawTime.Format(time.RFC3339),
		CreatedAt:      formatTime(w.CreatedAt),
		UpdatedAt:      formatTime(w.UpdatedAt),
		DeletedAt:      formatTimeWrapper(w.DeletedAt),
	}
}
