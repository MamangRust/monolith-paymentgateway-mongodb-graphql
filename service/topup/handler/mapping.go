package handler

import (
	"time"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/topup"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func toTopupResponse(t *salemongo.Topup) *pb.TopupResponse {
	if t == nil {
		return nil
	}

	return &pb.TopupResponse{
		Id:          t.TopupID,
		CardNumber:  t.CardNumber,
		TopupNo:     t.TopupNo.String(),
		TopupAmount: int32(t.TopupAmount),
		TopupMethod: t.TopupMethod,
		TopupTime:   t.TopupTime.Format(time.RFC3339),
		CreatedAt:   formatTime(t.CreatedAt),
		UpdatedAt:   formatTime(t.UpdatedAt),
	}
}

func toTopupDeleteAtResponse(t *salemongo.Topup) *pb.TopupResponseDeleteAt {
	if t == nil {
		return nil
	}

	var deletedAt *wrapperspb.StringValue
	if t.DeletedAt != nil {
		deletedAt = wrapperspb.String(t.DeletedAt.Format(time.RFC3339))
	}

	return &pb.TopupResponseDeleteAt{
		Id:          t.TopupID,
		CardNumber:  t.CardNumber,
		TopupNo:     t.TopupNo.String(),
		TopupAmount: int32(t.TopupAmount),
		TopupMethod: t.TopupMethod,
		TopupTime:   t.TopupTime.Format(time.RFC3339),
		CreatedAt:   formatTime(t.CreatedAt),
		UpdatedAt:   formatTime(t.UpdatedAt),
		DeletedAt:   deletedAt,
	}
}
