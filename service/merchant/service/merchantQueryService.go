package service

import (
	"context"

	cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis"
	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type merchantQueryDeps struct {
	Repository    repository.Repositories
	Cache         cache.MerchantQueryCache
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type merchantQueryService struct {
	queryRepo     repository.MerchantQueryRepository
	cache         cache.MerchantQueryCache
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewMerchantQueryService(params *merchantQueryDeps) MerchantQueryService {
	return &merchantQueryService{
		queryRepo:     params.Repository,
		cache:         params.Cache,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *merchantQueryService) FindAll(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, error) {
	const method = "FindAll"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCachedMerchants(ctx, req); found {
		logSuccess("Successfully retrieved all merchant records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	merchants, err := s.queryRepo.FindAllMerchants(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Merchant](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.queryRepo.CountAllMerchants(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Merchant](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCachedMerchants(ctx, req, merchants, &total)

	logSuccess("Successfully retrieved all merchant records", zap.Int("totalRecords", total), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return merchants, &total, nil
}

func (s *merchantQueryService) FindById(ctx context.Context, merchantID int) (*salemongo.Merchant, error) {
	const method = "FindById"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("merchant.id", merchantID))
	defer func() { end(status, "grpc") }()

	if cachedMerchant, found := s.cache.GetCachedMerchant(ctx, merchantID); found {
		logSuccess("Successfully retrieved merchant from cache", zap.Int("merchant.id", merchantID))
		return cachedMerchant, nil
	}

	merchant, err := s.queryRepo.FindByMerchantId(ctx, merchantID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("merchant.id", merchantID))
	}

	s.cache.SetCachedMerchant(ctx, merchant)

	logSuccess("Successfully retrieved merchant", zap.Int("merchant.id", merchantID))

	return merchant, nil
}

func (s *merchantQueryService) FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, error) {
	const method = "FindByActive"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCachedMerchantActive(ctx, req); found {
		logSuccess("Successfully fetched active merchants from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	merchants, err := s.queryRepo.FindByActive(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Merchant](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.queryRepo.CountActiveMerchants(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Merchant](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCachedMerchantActive(ctx, req, merchants, &total)

	logSuccess("Successfully fetched active merchants", zap.Int("totalRecords", total), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return merchants, &total, nil
}

func (s *merchantQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*salemongo.Merchant, *int, error) {
	const method = "FindByTrashed"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCachedMerchantTrashed(ctx, req); found {
		logSuccess("Successfully fetched trashed merchants from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	merchants, err := s.queryRepo.FindByTrashed(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Merchant](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.queryRepo.CountTrashedMerchants(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Merchant](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCachedMerchantTrashed(ctx, req, merchants, &total)

	logSuccess("Successfully fetched trashed merchants", zap.Int("totalRecords", total), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return merchants, &total, nil
}

func (s *merchantQueryService) FindByApiKey(ctx context.Context, apiKey string) (*salemongo.Merchant, error) {
	const method = "FindByApiKey"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("api_key", apiKey))
	defer func() { end(status, "grpc") }()

	if cachedMerchant, found := s.cache.GetCachedMerchantByApiKey(ctx, apiKey); found {
		logSuccess("Successfully found merchant by API key from cache", zap.String("api_key", apiKey))
		return cachedMerchant, nil
	}

	merchant, err := s.queryRepo.FindByApiKey(ctx, apiKey)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.String("api_key", apiKey))
	}

	s.cache.SetCachedMerchantByApiKey(ctx, apiKey, merchant)

	logSuccess("Successfully found merchant by API key", zap.String("api_key", apiKey))

	return merchant, nil
}

func (s *merchantQueryService) FindByMerchantUserId(ctx context.Context, userID int) ([]*salemongo.Merchant, error) {
	const method = "FindByMerchantUserId"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("user.id", userID))
	defer func() { end(status, "grpc") }()

	if cachedMerchants, found := s.cache.GetCachedMerchantsByUserId(ctx, userID); found {
		logSuccess("Successfully found merchants by user ID from cache", zap.Int("user.id", userID), zap.Int("count", len(cachedMerchants)))
		return cachedMerchants, nil
	}

	merchants, err := s.queryRepo.FindByMerchantUserId(ctx, userID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[[]*salemongo.Merchant](s.logger, err, method, span, zap.Int("user.id", userID))
	}

	s.cache.SetCachedMerchantsByUserId(ctx, userID, merchants)

	logSuccess("Successfully found merchants by user ID", zap.Int("user.id", userID), zap.Int("count", len(merchants)))

	return merchants, nil
}

func (s *merchantQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
