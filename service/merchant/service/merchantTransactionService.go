package service

import (
	"context"

	mencache "github.com/MamangRust/monolith-payment-gateway-merchant/redis"
	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type merchantTransactionDeps struct {
	Repository    repository.MerchantTransactionRepository
	Cache         mencache.MerchantTransactionCache
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type merchantTransactionService struct {
	repo          repository.MerchantTransactionRepository
	cache         mencache.MerchantTransactionCache
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewMerchantTransactionService(params *merchantTransactionDeps) MerchantTransactionService {
	return &merchantTransactionService{
		repo:          params.Repository,
		cache:         params.Cache,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *merchantTransactionService) FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*salemongo.Transaction, *int, error) {
	const method = "FindAllTransactions"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCacheAllMerchantTransactions(ctx, req); found {
		logSuccess("Successfully retrieved all merchant transactions from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	transactions, err := s.repo.FindAllTransactions(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Transaction](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.repo.CountAllTransactions(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Transaction](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCacheAllMerchantTransactions(ctx, req, transactions, &total)

	logSuccess("Successfully retrieved all merchant transactions", zap.Int("totalRecords", total), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return transactions, &total, nil
}

func (s *merchantTransactionService) FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*salemongo.Transaction, *int, error) {
	const method = "FindAllTransactionsByMerchant"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCacheMerchantTransactions(ctx, req); found {
		logSuccess("Successfully retrieved merchant transactions from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	transactions, err := s.repo.FindAllTransactionsByMerchant(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Transaction](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.repo.CountTransactionsByMerchant(ctx, req.MerchantID, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Transaction](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCacheMerchantTransactions(ctx, req, transactions, &total)

	logSuccess("Successfully retrieved merchant transactions", zap.Int("totalRecords", total), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return transactions, &total, nil
}

func (s *merchantTransactionService) FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*salemongo.Transaction, *int, error) {
	const method = "FindAllTransactionsByApikey"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCacheMerchantTransactionApikey(ctx, req); found {
		logSuccess("Successfully retrieved merchant transactions by apikey from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	transactions, err := s.repo.FindAllTransactionsByApikey(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Transaction](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.repo.CountTransactionsByApikey(ctx, req.ApiKey, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.Transaction](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCacheMerchantTransactionApikey(ctx, req, transactions, &total)

	logSuccess("Successfully retrieved merchant transactions by apikey", zap.Int("totalRecords", total), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return transactions, &total, nil
}

func (s *merchantTransactionService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
