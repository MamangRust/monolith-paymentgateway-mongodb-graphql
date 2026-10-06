package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	mencache "github.com/MamangRust/monolith-payment-gateway-topup/redis"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type topupQueryDeps struct {
	Cache         mencache.TopupQueryCache
	Repository    repository.TopupQueryRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type topupQueryService struct {
	cache                mencache.TopupQueryCache
	topupQueryRepository repository.TopupQueryRepository
	logger               logger.LoggerInterface

	observability observability.TraceLoggerObservability
}

func NewTopupQueryService(
	params *topupQueryDeps,
) TopupQueryService {
	return &topupQueryService{
		cache:                params.Cache,
		topupQueryRepository: params.Repository,
		logger:               params.Logger,
		observability:        params.Observability,
	}
}

func (s *topupQueryService) FindAll(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error) {
	const method = "FindAll"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCachedTopupsCache(ctx, req); found {
		logSuccess("Successfully retrieved all topup records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	topups, err := s.topupQueryRepository.FindAllTopups(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.TopupQueryResult](
			s.logger,
			sharedErrors.ErrFailed("find all topups"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	var totalCount int

	if len(topups) > 0 {
		totalCount = int(topups[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCachedTopupsCache(ctx, req, topups, &totalCount)

	logSuccess("Successfully fetched topup",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return topups, &totalCount, nil
}

func (s *topupQueryService) FindAllByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*repository.TopupByCardResult, *int, error) {
	const method = "FindAllByCardNumber"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search
	card_number := req.CardNumber

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCacheTopupByCardCache(ctx, req); found {
		logSuccess("Successfully retrieved all topup records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	topups, err := s.topupQueryRepository.FindAllTopupByCardNumber(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.TopupByCardResult](
			s.logger,
			sharedErrors.ErrFailed("find all topups by card number"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
			zap.String("card_number", card_number),
		)
	}

	var totalCount int

	if len(topups) > 0 {
		totalCount = int(topups[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCacheTopupByCardCache(ctx, req, topups, &totalCount)

	logSuccess("Successfully fetched topup by card number",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("card_number", card_number))

	return topups, &totalCount, nil
}

func (s *topupQueryService) FindById(ctx context.Context, topupID int) (*salemongo.Topup, error) {
	const method = "FindById"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("topup_id", topupID))

	defer func() {
		end(status, "grpc")
	}()

	if data, found := s.cache.GetCachedTopupCache(ctx, topupID); found {
		logSuccess("Successfully retrieved topup from cache", zap.Int("topup.id", topupID))
		return data, nil
	}

	topup, err := s.topupQueryRepository.FindById(ctx, topupID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Topup](
			s.logger,
			sharedErrors.ErrNotFoundResponse("Topup"),
			method,
			span,

			zap.Int("topup_id", topupID),
		)
	}

	s.cache.SetCachedTopupCache(ctx, topup)

	logSuccess("Successfully fetched topup", zap.Int("topup_id", topupID))

	return topup, nil
}

func (s *topupQueryService) FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error) {
	const method = "FindByActive"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCachedTopupActiveCache(ctx, req); found {
		logSuccess("Successfully retrieved all topup records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	topups, err := s.topupQueryRepository.FindByActive(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.TopupQueryResult](
			s.logger,
			sharedErrors.ErrFailed("find active topups"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	var totalCount int

	if len(topups) > 0 {
		totalCount = int(topups[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCachedTopupActiveCache(ctx, req, topups, &totalCount)

	logSuccess("Successfully fetched active topup",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return topups, &totalCount, nil
}

func (s *topupQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error) {
	const method = "FindByTrashed"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCachedTopupTrashedCache(ctx, req); found {
		logSuccess("Successfully retrieved all topup records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	topups, err := s.topupQueryRepository.FindByTrashed(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.TopupQueryResult](
			s.logger,
			sharedErrors.ErrFailed("find trashed topups"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	var totalCount int

	if len(topups) > 0 {
		totalCount = int(topups[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCachedTopupTrashedCache(ctx, req, topups, &totalCount)

	logSuccess("Successfully fetched trashed topup",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return topups, &totalCount, nil
}

func (s *topupQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
