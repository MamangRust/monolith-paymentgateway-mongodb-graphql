package cardgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/model"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbStats "github.com/MamangRust/monolith-payment-gateway-pb/card/stats"
)

type cardResponseMapper struct {
}

func NewCardResponseMapper() *cardResponseMapper {
	return &cardResponseMapper{}
}

func (s *cardResponseMapper) ToGraphqlResponseCard(res *pb.ApiResponseCard) *model.APIResponseCard {
	return &model.APIResponseCard{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapCardResponse(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlResponsePaginationCard(res *pb.ApiResponsePaginationCard) *model.APIResponsePaginationCard {
	return &model.APIResponsePaginationCard{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapCardResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *cardResponseMapper) ToGraphqlResponseAll(res *pb.ApiResponseCardAll) *model.APIResponseCardAll {
	return &model.APIResponseCardAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *cardResponseMapper) ToGraphqlResponseDelete(res *pb.ApiResponseCardDelete) *model.APIResponseCardDelete {
	return &model.APIResponseCardDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *cardResponseMapper) ToGraphqlResponseCardDeleteAt(res *pb.ApiResponseCardDeleteAt) *model.APIResponseCardDeleteAt {
	return &model.APIResponseCardDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapCardResponseDeleteAt(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlResponsePaginationCardDeleteAt(res *pb.ApiResponsePaginationCardDeleteAt) *model.APIResponsePaginationCardDeleteAt {
	return &model.APIResponsePaginationCardDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapCardDeleteAtResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *cardResponseMapper) ToGraphqlDashboardCard(res *pb.ApiResponseDashboardCard) *model.APIResponseDashboardCard {
	return &model.APIResponseDashboardCard{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapDashboardCard(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlDashboardCardCardNumber(res *pb.ApiResponseDashboardCardNumber) *model.APIResponseDashboardCardNumber {
	return &model.APIResponseDashboardCardNumber{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapDashboardCardCardNumber(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlMonthlyBalances(res *pbStats.ApiResponseMonthlyBalance) *model.APIResponseMonthlyBalance {
	return &model.APIResponseMonthlyBalance{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapMonthlyBalances(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlYearlyBalances(res *pbStats.ApiResponseYearlyBalance) *model.APIResponseYearlyBalance {
	return &model.APIResponseYearlyBalance{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapYearlyBalances(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlMonthlyAmounts(res *pb.ApiResponseMonthlyAmount) *model.APIResponseMonthlyAmount {
	return &model.APIResponseMonthlyAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapMonthlyAmounts(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlYearlyAmounts(res *pb.ApiResponseYearlyAmount) *model.APIResponseYearlyAmount {
	return &model.APIResponseYearlyAmount{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapYearlyAmounts(res.Data),
	}
}

// map

func (s *cardResponseMapper) mapCardResponse(card *pb.CardResponse) *model.CardResponse {
	return &model.CardResponse{
		ID:           int32(card.Id),
		UserID:       int32(card.UserId),
		CardNumber:   card.CardNumber,
		CardType:     card.CardType,
		ExpireDate:   card.ExpireDate,
		Cvv:          card.Cvv,
		CardProvider: card.CardProvider,
		CreatedAt:    card.CreatedAt,
		UpdatedAt:    card.UpdatedAt,
	}
}

func (s *cardResponseMapper) mapCardResponses(cards []*pb.CardResponse) []*model.CardResponse {
	var responseCards []*model.CardResponse

	for _, card := range cards {
		responseCards = append(responseCards, s.mapCardResponse(card))
	}

	return responseCards
}

func (s *cardResponseMapper) mapCardResponseDeleteAt(card *pb.CardResponseDeleteAt) *model.CardResponseDeleteAt {
	var deletedAt string

	if card.DeletedAt != nil {
		deletedAt = card.DeletedAt.Value
	}

	return &model.CardResponseDeleteAt{
		ID:           int32(card.Id),
		UserID:       int32(card.UserId),
		CardNumber:   card.CardNumber,
		CardType:     card.CardType,
		ExpireDate:   card.ExpireDate,
		Cvv:          card.Cvv,
		CardProvider: card.CardProvider,
		CreatedAt:    card.CreatedAt,
		UpdatedAt:    card.UpdatedAt,
		DeletedAt:    &deletedAt,
	}
}

func (s *cardResponseMapper) mapCardDeleteAtResponses(cards []*pb.CardResponseDeleteAt) []*model.CardResponseDeleteAt {
	var responseCards []*model.CardResponseDeleteAt

	for _, card := range cards {
		responseCards = append(responseCards, s.mapCardResponseDeleteAt(card))
	}

	return responseCards
}

func (s *cardResponseMapper) mapDashboardCard(dash *pb.CardResponseDashboard) *model.CardDashboardResponse {
	return &model.CardDashboardResponse{
		TotalBalance:     int32(dash.TotalBalance),
		TotalWithdraw:    int32(dash.TotalWithdraw),
		TotalTopup:       int32(dash.TotalTopup),
		TotalTransfer:    int32(dash.TotalTransfer),
		TotalTransaction: int32(dash.TotalTransaction),
	}
}

func (s *cardResponseMapper) mapDashboardCardCardNumber(dash *pb.CardResponseDashboardCardNumber) *model.CardDashboardByNumberResponse {
	return &model.CardDashboardByNumberResponse{
		TotalBalance:          int32(dash.TotalBalance),
		TotalWithdraw:         int32(dash.TotalWithdraw),
		TotalTopup:            int32(dash.TotalTopup),
		TotalTransferSend:     int32(dash.TotalTransferSend),
		TotalTransferReceiver: int32(dash.TotalTransferReceiver),
		TotalTransaction:      int32(dash.TotalTransaction),
	}
}

func (s *cardResponseMapper) mapMonthlyBalance(cards *pbStats.CardResponseMonthlyBalance) *model.CardMonthlyBalanceResponse {
	return &model.CardMonthlyBalanceResponse{
		Month:        cards.Month,
		TotalBalance: int32(cards.TotalBalance),
	}
}

func (s *cardResponseMapper) mapMonthlyBalances(cards []*pbStats.CardResponseMonthlyBalance) []*model.CardMonthlyBalanceResponse {
	var responseCards []*model.CardMonthlyBalanceResponse

	for _, role := range cards {
		responseCards = append(responseCards, s.mapMonthlyBalance(role))
	}

	return responseCards
}

func (s *cardResponseMapper) mapYearlyBalance(cards *pbStats.CardResponseYearlyBalance) *model.CardYearlyBalanceResponse {
	return &model.CardYearlyBalanceResponse{
		Year:         cards.Year,
		TotalBalance: int32(cards.TotalBalance),
	}
}

func (s *cardResponseMapper) mapYearlyBalances(cards []*pbStats.CardResponseYearlyBalance) []*model.CardYearlyBalanceResponse {
	var responseCards []*model.CardYearlyBalanceResponse

	for _, role := range cards {
		responseCards = append(responseCards, s.mapYearlyBalance(role))
	}

	return responseCards
}

func (s *cardResponseMapper) mapMonthlyAmount(cards *pb.CardResponseMonthlyAmount) *model.CardMonthlyAmountResponse {
	return &model.CardMonthlyAmountResponse{
		Month:       cards.Month,
		TotalAmount: int32(cards.TotalAmount),
	}
}

func (s *cardResponseMapper) mapMonthlyAmounts(cards []*pb.CardResponseMonthlyAmount) []*model.CardMonthlyAmountResponse {
	var responseCards []*model.CardMonthlyAmountResponse

	for _, role := range cards {
		responseCards = append(responseCards, s.mapMonthlyAmount(role))
	}

	return responseCards
}

func (s *cardResponseMapper) mapYearlyAmount(cards *pb.CardResponseYearlyAmount) *model.CardYearlyAmountResponse {
	return &model.CardYearlyAmountResponse{
		Year:        cards.Year,
		TotalAmount: int32(cards.TotalAmount),
	}
}

func (s *cardResponseMapper) mapYearlyAmounts(cards []*pb.CardResponseYearlyAmount) []*model.CardYearlyAmountResponse {
	var responseCards []*model.CardYearlyAmountResponse

	for _, role := range cards {
		responseCards = append(responseCards, s.mapYearlyAmount(role))
	}

	return responseCards
}
