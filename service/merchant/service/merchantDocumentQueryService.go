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

type merchantDocumentQueryDeps struct {
	Cache         mencache.MerchantDocumentQueryCache
	Repository    repository.MerchantDocumentQueryRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type merchantDocumentQueryService struct {
	cache         mencache.MerchantDocumentQueryCache
	queryRepo     repository.MerchantDocumentQueryRepository
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewMerchantDocumentQueryService(params *merchantDocumentQueryDeps) MerchantDocumentQueryService {
	return &merchantDocumentQueryService{
		cache:         params.Cache,
		queryRepo:     params.Repository,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *merchantDocumentQueryService) FindAll(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, error) {
	const method = "FindAll"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCachedMerchantDocuments(ctx, req); found {
		logSuccess("Successfully retrieved all merchant document records from cache", zap.Int("total", *total), zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))
		return data, total, nil
	}

	merchantDocuments, err := s.queryRepo.FindAllDocuments(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.MerchantDocument](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.queryRepo.CountAllDocuments(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.MerchantDocument](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCachedMerchantDocuments(ctx, req, merchantDocuments, &total)

	logSuccess("Successfully retrieved all merchant document records", zap.Int("total", total), zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return merchantDocuments, &total, nil
}

func (s *merchantDocumentQueryService) FindById(ctx context.Context, documentID int) (*salemongo.MerchantDocument, error) {
	const method = "FindById"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("merchantDocument.id", documentID))
	defer func() { end(status, "grpc") }()

	if cachedMerchant, found := s.cache.GetCachedMerchantDocument(ctx, documentID); found {
		logSuccess("Successfully found merchant document by ID from cache", zap.Int("merchantDocument.id", documentID))
		return cachedMerchant, nil
	}

	merchantDocument, err := s.queryRepo.FindByIdDocument(ctx, documentID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("merchantDocument.id", documentID))
	}

	s.cache.SetCachedMerchantDocument(ctx, documentID, merchantDocument)

	logSuccess("Successfully found merchant document by ID", zap.Int("merchantDocument.id", documentID))

	return merchantDocument, nil
}

func (s *merchantDocumentQueryService) FindByActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, error) {
	const method = "FindByActive"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCachedMerchantDocumentsActive(ctx, req); found {
		logSuccess("Successfully retrieved active merchant document records from cache", zap.Int("total", *total), zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))
		return data, total, nil
	}

	merchantDocuments, err := s.queryRepo.FindByActiveDocuments(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.MerchantDocument](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.queryRepo.CountActiveDocuments(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.MerchantDocument](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCachedMerchantDocumentsActive(ctx, req, merchantDocuments, &total)

	logSuccess("Successfully retrieved active merchant document records", zap.Int("total", total), zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return merchantDocuments, &total, nil
}

func (s *merchantDocumentQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*salemongo.MerchantDocument, *int, error) {
	const method = "FindByTrashed"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCachedMerchantDocumentsTrashed(ctx, req); found {
		logSuccess("Successfully retrieved trashed merchant document records from cache", zap.Int("total", *total), zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))
		return data, total, nil
	}

	merchantDocuments, err := s.queryRepo.FindByTrashedDocuments(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.MerchantDocument](s.logger, err, method, span, zap.String("search", search))
	}

	totalCount, err := s.queryRepo.CountTrashedDocuments(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.MerchantDocument](s.logger, err, method, span, zap.String("search", search))
	}

	total := int(totalCount)
	s.cache.SetCachedMerchantDocumentsTrashed(ctx, req, merchantDocuments, &total)

	logSuccess("Successfully retrieved trashed merchant document records", zap.Int("total", total), zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return merchantDocuments, &total, nil
}

func (s *merchantDocumentQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
