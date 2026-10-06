package handler

import (
	"time"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/transfer"
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

func toTransferResponse(t *salemongo.Transfer) *pb.TransferResponse {
	if t == nil {
		return nil
	}

	return &pb.TransferResponse{
		Id:             t.TransferID,
		TransferNo:     t.TransferNo.String(),
		TransferFrom:   t.TransferFrom,
		TransferTo:     t.TransferTo,
		TransferAmount: int32(t.TransferAmount),
		TransferTime:   t.TransferTime.Format(time.RFC3339),
		CreatedAt:      formatTime(t.CreatedAt),
		UpdatedAt:      formatTime(t.UpdatedAt),
	}
}

func toTransferDeleteAtResponse(t *salemongo.Transfer) *pb.TransferResponseDeleteAt {
	if t == nil {
		return nil
	}

	return &pb.TransferResponseDeleteAt{
		Id:             t.TransferID,
		TransferNo:     t.TransferNo.String(),
		TransferFrom:   t.TransferFrom,
		TransferTo:     t.TransferTo,
		TransferAmount: int32(t.TransferAmount),
		TransferTime:   t.TransferTime.Format(time.RFC3339),
		CreatedAt:      formatTime(t.CreatedAt),
		UpdatedAt:      formatTime(t.UpdatedAt),
		DeletedAt:      formatTimeWrapper(t.DeletedAt),
	}
}
