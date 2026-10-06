package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	mencache "github.com/MamangRust/monolith-payment-gateway-saldo/redis"
	"github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type saldoQueryParams struct {
	Cache         mencache.SaldoQueryCache
	Repository    repository.SaldoQueryRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type saldoQueryService struct {
	mencache             mencache.SaldoQueryCache
	saldoQueryRepository repository.SaldoQueryRepository
	logger               logger.LoggerInterface
	observability        observability.TraceLoggerObservability
}

func NewSaldoQueryService(params *saldoQueryParams) SaldoQueryService {
	return &saldoQueryService{
		mencache:             params.Cache,
		saldoQueryRepository: params.Repository,
		logger:               params.Logger,
		observability:        params.Observability,
	}
}

func (s *saldoQueryService) ListReconciliationQueue(ctx context.Context, status string, limit int32) ([]*repository.ReconciliationQueueResult, error) {
	return s.saldoQueryRepository.ListReconciliationQueue(ctx, status, limit)
}

func (s *saldoQueryService) ListLedgerEntries(ctx context.Context, cardNumber string, limit int32) ([]*repository.LedgerEntryResult, error) {
	return s.saldoQueryRepository.ListLedgerEntries(ctx, cardNumber, limit)
}

func (s *saldoQueryService) FindAll(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, error) {
	const method = "FindAll"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	res, err := s.saldoQueryRepository.FindAllSaldos(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.SaldoResult](s.logger, err, method, span)
	}

	totalCount, err := s.saldoQueryRepository.CountAllSaldos(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.SaldoResult](s.logger, err, method, span)
	}

	total := int(totalCount)
	logSuccess("Successfully fetched saldo", zap.Int("totalRecords", total))
	return res, &total, nil
}

func (s *saldoQueryService) FindByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, error) {
	const method = "FindByActive"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	res, err := s.saldoQueryRepository.FindByActive(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.SaldoResult](s.logger, err, method, span)
	}

	totalCount, err := s.saldoQueryRepository.CountActiveSaldos(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.SaldoResult](s.logger, err, method, span)
	}

	total := int(totalCount)
	logSuccess("Successfully fetched active saldo", zap.Int("totalRecords", total))
	return res, &total, nil
}

func (s *saldoQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, error) {
	const method = "FindByTrashed"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	res, err := s.saldoQueryRepository.FindByTrashed(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.SaldoResult](s.logger, err, method, span)
	}

	totalCount, err := s.saldoQueryRepository.CountTrashedSaldos(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.SaldoResult](s.logger, err, method, span)
	}

	total := int(totalCount)
	logSuccess("Successfully fetched trashed saldo", zap.Int("totalRecords", total))
	return res, &total, nil
}

func (s *saldoQueryService) FindById(ctx context.Context, saldo_id int) (*repository.SaldoResult, error) {
	const method = "FindById"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("saldo_id", saldo_id))
	defer func() { end(status, "grpc") }()

	res, err := s.saldoQueryRepository.FindById(ctx, saldo_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoResult](s.logger, sharedErrors.ErrNotFoundResponse("Saldo"), method, span)
	}

	logSuccess("Successfully fetched saldo", zap.Int("saldo_id", saldo_id))
	return res, nil
}

func (s *saldoQueryService) FindByCardNumber(ctx context.Context, card_number string) (*salemongo.Saldo, error) {
	const method = "FindByCardNumber"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", card_number))
	defer func() { end(status, "grpc") }()

	res, err := s.saldoQueryRepository.FindByCardNumber(ctx, card_number)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Saldo](s.logger, sharedErrors.ErrNotFoundResponse("Saldo"), method, span)
	}

	logSuccess("Successfully fetched saldo by card number", zap.String("card_number", card_number))
	return res, nil
}

func (s *saldoQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
