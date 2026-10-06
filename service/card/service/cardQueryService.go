package service

import (
	"context"

	mencache "github.com/MamangRust/monolith-payment-gateway-card/redis"
	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type cardQueryServiceDeps struct {
	Cache               mencache.CardQueryCache
	CardQueryRepository repository.CardQueryRepository
	UserRepository      repository.UserRepository
	Logger              logger.LoggerInterface
	Observability       observability.TraceLoggerObservability
}

type cardQueryService struct {
	cache               mencache.CardQueryCache
	cardQueryRepository repository.CardQueryRepository
	userRepository      repository.UserRepository
	logger              logger.LoggerInterface
	observability       observability.TraceLoggerObservability
}

func NewCardQueryService(params *cardQueryServiceDeps) CardQueryService {
	return &cardQueryService{
		cardQueryRepository: params.CardQueryRepository,
		userRepository:      params.UserRepository,
		logger:              params.Logger,
		observability:       params.Observability,
		cache:               params.Cache,
	}
}

func (s *cardQueryService) FindAll(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, error) {
	const method = "FindAll"
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", req.Search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetFindAllCache(ctx, req); found {
		logSuccess("Successfully fetched card records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	cards, err := s.cardQueryRepository.FindAllCards(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*salemongo.Card](s.logger, err, method, span,
			zap.Int("page", req.Page), zap.Int("pageSize", req.PageSize), zap.String("search", req.Search))
	}

	totalCount, err := s.cardQueryRepository.CountAllCards(ctx, req.Search)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*salemongo.Card](s.logger, err, method, span)
	}

	total := int(totalCount)
	s.cache.SetFindAllCache(ctx, req, cards, &total)
	logSuccess("Successfully fetched card records", zap.Int("totalRecords", total), zap.Int("page", page), zap.Int("pageSize", pageSize))
	return cards, &total, nil
}

func (s *cardQueryService) FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, error) {
	const method = "FindByActive"
	page := req.Page
	pageSize := req.PageSize
	search := req.Search
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetByActiveCache(ctx, req); found {
		logSuccess("Successfully fetched active card records from cache", zap.Int("totalRecords", *total))
		return data, total, nil
	}

	res, err := s.cardQueryRepository.FindByActive(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*salemongo.Card](s.logger, err, method, span)
	}

	totalCount, err := s.cardQueryRepository.CountActiveCards(ctx, search)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*salemongo.Card](s.logger, err, method, span)
	}

	total := int(totalCount)
	s.cache.SetByActiveCache(ctx, req, res, &total)
	logSuccess("Successfully fetched active card records", zap.Int("totalRecords", total))
	return res, &total, nil
}

func (s *cardQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*salemongo.Card, *int, error) {
	const method = "FindByTrashed"
	page := req.Page
	pageSize := req.PageSize
	search := req.Search
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetByTrashedCache(ctx, req); found {
		logSuccess("Successfully fetched trashed card records from cache", zap.Int("totalRecords", *total))
		return data, total, nil
	}

	res, err := s.cardQueryRepository.FindByTrashed(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*salemongo.Card](s.logger, err, method, span)
	}

	totalCount, err := s.cardQueryRepository.CountTrashedCards(ctx, search)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandlerErrorPagination[[]*salemongo.Card](s.logger, err, method, span)
	}

	total := int(totalCount)
	s.cache.SetByTrashedCache(ctx, req, res, &total)
	logSuccess("Successfully fetched trashed card records", zap.Int("totalRecords", total))
	return res, &total, nil
}

func (s *cardQueryService) FindById(ctx context.Context, card_id int) (*salemongo.Card, error) {
	const method = "FindById"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("card_id", card_id))
	defer func() { end(status, "grpc") }()

	if data, found := s.cache.GetByIdCache(ctx, card_id); found {
		logSuccess("Successfully fetched card from cache", zap.Int("card.id", card_id))
		return data, nil
	}

	res, err := s.cardQueryRepository.FindById(ctx, card_id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("card_id", card_id))
	}

	s.cache.SetByIdCache(ctx, card_id, res)
	logSuccess("Successfully fetched card", zap.Int("card_id", card_id))
	return res, nil
}

func (s *cardQueryService) FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Card, error) {
	const method = "FindByCardNumber"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("card_number", card_number))
	defer func() { end(status, "grpc") }()

	if data, found := s.cache.GetByCardNumberCache(ctx, card_number); found {
		logSuccess("Successfully fetched card by card number from cache")
		return data, nil
	}

	res, err := s.cardQueryRepository.FindCardByCardNumber(ctx, card_number)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.String("card_number", card_number))
	}

	s.cache.SetByCardNumberCache(ctx, card_number, res)
	logSuccess("Successfully fetched card by card number")
	return res, nil
}

func (s *cardQueryService) FindByUserID(ctx context.Context, user_id int) (*salemongo.Card, error) {
	const method = "FindByUserID"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("user_id", user_id))
	defer func() { end(status, "grpc") }()

	if data, found := s.cache.GetByUserIDCache(ctx, user_id); found {
		logSuccess("Successfully fetched card by user ID from cache")
		return data, nil
	}

	res, err := s.cardQueryRepository.FindCardByUserId(ctx, user_id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("user_id", user_id))
	}

	s.cache.SetByUserIDCache(ctx, user_id, res)
	logSuccess("Successfully fetched card by user ID")
	return res, nil
}

func (s *cardQueryService) FindUserCardByCardNumber(ctx context.Context, card_number string) (*salemongo.Card, error) {
	const method = "FindUserCardByCardNumber"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("card_number", card_number))
	defer func() { end(status, "grpc") }()

	if data, found := s.cache.GetUserCardByCardNumberCache(ctx, card_number); found {
		logSuccess("Successfully fetched user card by card number from cache")
		return data, nil
	}

	res, err := s.cardQueryRepository.FindUserCardByCardNumber(ctx, card_number)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, sharedErrors.ErrFailed("find card by user ID"), method, span, zap.String("card_number", card_number))
	}

	s.cache.SetUserCardByCardNumberCache(ctx, card_number, res)
	logSuccess("Successfully fetched user card by card number")
	return res, nil
}

func (s *cardQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
