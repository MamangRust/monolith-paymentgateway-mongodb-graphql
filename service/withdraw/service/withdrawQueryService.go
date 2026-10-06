package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	mencache "github.com/MamangRust/monolith-payment-gateway-withdraw/redis"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type withdrawQueryServiceDeps struct {
	Cache      mencache.WithdrawQueryCache
	Repository repository.WithdrawQueryRepository

	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type withdrawQueryService struct {
	cache                   mencache.WithdrawQueryCache
	withdrawQueryRepository repository.WithdrawQueryRepository

	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewWithdrawQueryService(
	deps *withdrawQueryServiceDeps,
) WithdrawQueryService {
	return &withdrawQueryService{
		cache:                   deps.Cache,
		withdrawQueryRepository: deps.Repository,
		logger:                  deps.Logger,
		observability:           deps.Observability,
	}
}

func (s *withdrawQueryService) FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error) {
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

	if data, total, found := s.cache.GetCachedWithdrawsCache(ctx, req); found {
		logSuccess("Successfully retrieved all withdraw records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	withdraws, err := s.withdrawQueryRepository.FindAll(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.WithdrawQueryResult](
			s.logger,
			sharedErrors.ErrFailed("find all withdraws"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	var totalCount int

	if len(withdraws) > 0 {
		totalCount = int(withdraws[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCachedWithdrawsCache(ctx, req, withdraws, &totalCount)

	logSuccess("Successfully fetched withdraw",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return withdraws, &totalCount, nil
}

func (s *withdrawQueryService) FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*repository.WithdrawQueryResult, *int, error) {
	const method = "FindAllByCardNumber"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search),
		attribute.String("card_number", req.CardNumber))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCachedWithdrawByCardCache(ctx, req); found {
		logSuccess("Successfully retrieved all withdraw records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	withdraws, err := s.withdrawQueryRepository.FindAllByCardNumber(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.WithdrawQueryResult](
			s.logger,
			sharedErrors.ErrFailed("find all by card number"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
			zap.String("card_number", req.CardNumber),
		)
	}

	var totalCount int

	if len(withdraws) > 0 {
		totalCount = int(withdraws[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCachedWithdrawByCardCache(ctx, req, withdraws, &totalCount)

	logSuccess("Successfully fetched withdraw",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return withdraws, &totalCount, nil
}

func (s *withdrawQueryService) FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error) {
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

	if data, total, found := s.cache.GetCachedWithdrawActiveCache(ctx, req); found {
		logSuccess("Successfully retrieved all withdraw records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	withdraws, err := s.withdrawQueryRepository.FindByActive(ctx, req)

	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.WithdrawQueryResult](
			s.logger,
			sharedErrors.ErrFailed("find active withdraws"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	var totalCount int

	if len(withdraws) > 0 {
		totalCount = int(withdraws[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCachedWithdrawActiveCache(ctx, req, withdraws, &totalCount)

	logSuccess("Successfully fetched active withdraw",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return withdraws, &totalCount, nil
}

func (s *withdrawQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error) {
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

	if data, total, found := s.cache.GetCachedWithdrawTrashedCache(ctx, req); found {
		logSuccess("Successfully retrieved all withdraw records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	withdraws, err := s.withdrawQueryRepository.FindByTrashed(ctx, req)

	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*repository.WithdrawQueryResult](
			s.logger,
			sharedErrors.ErrFailed("find trashed withdraws"),
			method,
			span,

			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	var totalCount int

	if len(withdraws) > 0 {
		totalCount = int(withdraws[0].TotalCount)
	} else {
		totalCount = 0
	}

	s.cache.SetCachedWithdrawTrashedCache(ctx, req, withdraws, &totalCount)

	logSuccess("Successfully fetched trashed withdraw",
		zap.Int("totalRecords", totalCount), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return withdraws, &totalCount, nil
}

func (s *withdrawQueryService) FindById(ctx context.Context, withdrawID int) (*salemongo.Withdraw, error) {
	const method = "FindById"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("withdraw_id", withdrawID))

	defer func() {
		end(status, "grpc")
	}()

	withdraw, err := s.withdrawQueryRepository.FindById(ctx, withdrawID)

	if data, found := s.cache.GetCachedWithdrawCache(ctx, withdrawID); found {
		logSuccess("Successfully retrieved withdraw from cache", zap.Int("withdraw_id", withdrawID))
		return data, nil
	}

	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Withdraw](
			s.logger,
			sharedErrors.ErrNotFoundResponse("Withdraw"),
			method,
			span,

			zap.Int("withdraw_id", withdrawID),
		)
	}

	s.cache.SetCachedWithdrawCache(ctx, withdraw)

	logSuccess("Successfully fetched withdraw", zap.Int("withdraw_id", withdrawID))

	return withdraw, nil
}

func (s *withdrawQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
