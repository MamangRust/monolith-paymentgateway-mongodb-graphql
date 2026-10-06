package handler

import (
	"time"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/transaction"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatTxnDeletedAt(t *time.Time) *wrapperspb.StringValue {
	if t == nil || t.IsZero() {
		return nil
	}
	return wrapperspb.String(t.Format(time.RFC3339))
}

func toTransactionResponse(t *salemongo.Transaction) *pb.TransactionResponse {
	if t == nil {
		return nil
	}

	return &pb.TransactionResponse{
		Id:              t.TransactionID,
		CardNumber:      t.CardNumber,
		TransactionNo:   t.TransactionNo.String(),
		Amount:          int32(t.Amount),
		PaymentMethod:   t.PaymentMethod,
		MerchantId:      t.MerchantID,
		TransactionTime: t.TransactionTime.Format(time.RFC3339),
		CreatedAt:       formatTime(t.CreatedAt),
		UpdatedAt:       formatTime(t.UpdatedAt),
	}
}

func toTransactionDeleteAtResponse(t *salemongo.Transaction) *pb.TransactionResponseDeleteAt {
	if t == nil {
		return nil
	}

	return &pb.TransactionResponseDeleteAt{
		Id:              t.TransactionID,
		CardNumber:      t.CardNumber,
		TransactionNo:   t.TransactionNo.String(),
		Amount:          int32(t.Amount),
		PaymentMethod:   t.PaymentMethod,
		MerchantId:      t.MerchantID,
		TransactionTime: t.TransactionTime.Format(time.RFC3339),
		CreatedAt:       formatTime(t.CreatedAt),
		UpdatedAt:       formatTime(t.UpdatedAt),
		DeletedAt:       formatTxnDeletedAt(t.DeletedAt),
	}
}
