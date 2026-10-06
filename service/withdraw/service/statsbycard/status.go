package withdrawstatsbycardservice

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	withdraw_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/withdraw_errors/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	mencache "github.com/MamangRust/monolith-payment-gateway-withdraw/redis/statsbycard"
	repository "github.com/MamangRust/monolith-payment-gateway-withdraw/repository/statsbycard"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type WithdrawStatsByCardStatusDeps struct {
	Cache mencache.WithdrawStatsByCardStatusCache

	Repository repository.WithdrawStatsByCardStatusRepository

	Logger logger.LoggerInterface

	Observability observability.TraceLoggerObservability
}

type withdrawStatsByCardStatusService struct {
	cache mencache.WithdrawStatsByCardStatusCache

	repository repository.WithdrawStatsByCardStatusRepository

	logger logger.LoggerInterface

	observability observability.TraceLoggerObservability
}

func NewWithdrawStatsByCardStatusService(params *WithdrawStatsByCardStatusDeps) WithdrawStatsByCardStatusService {

	return &withdrawStatsByCardStatusService{
		cache:         params.Cache,
		repository:    params.Repository,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *withdrawStatsByCardStatusService) FindMonthWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusSuccessByCardRow, error) {
	const method = "FindMonthWithdrawStatusSuccessByCardNumber"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		attribute.Int("year", req.Year),
		attribute.Int("month", req.Month))

	defer func() {
		end(status)
	}()

	if dbRows, found := s.cache.GetCachedMonthWithdrawStatusSuccessByCardNumber(ctx, req); found {
		logSuccess("Successfully fetched monthly withdraw status success by card number (from cache)",
			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year),
			zap.Int("month", req.Month))
		return dbRows, nil
	}

	dbRows, err := s.repository.GetMonthWithdrawStatusSuccessByCardNumber(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[[]*salemongo.WithdrawMonthlyStatusSuccessByCardRow](
			s.logger,
			withdraw_errors.ErrFailedFindMonthWithdrawStatusSuccess,
			method,
			span,

			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year),
			zap.Int("month", req.Month),
		)
	}

	s.cache.SetCachedMonthWithdrawStatusSuccessByCardNumber(ctx, req, dbRows)

	logSuccess("Successfully fetched monthly withdraw status success by card number (from DB)",
		zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		zap.Int("year", req.Year),
		zap.Int("month", req.Month))

	return dbRows, nil
}

func (s *withdrawStatsByCardStatusService) FindYearlyWithdrawStatusSuccessByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusSuccessByCardRow, error) {
	const method = "FindYearlyWithdrawStatusSuccessByCardNumber"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		attribute.Int("year", req.Year))

	defer func() {
		end(status)
	}()

	if dbRows, found := s.cache.GetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx, req); found {
		logSuccess("Successfully fetched yearly withdraw status success by card number (from cache)",
			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year))
		return dbRows, nil
	}

	dbRows, err := s.repository.GetYearlyWithdrawStatusSuccessByCardNumber(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[[]*salemongo.WithdrawYearlyStatusSuccessByCardRow](
			s.logger,
			withdraw_errors.ErrFailedFindYearWithdrawStatusSuccess,
			method,
			span,

			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year),
		)
	}

	s.cache.SetCachedYearlyWithdrawStatusSuccessByCardNumber(ctx, req, dbRows)

	logSuccess("Successfully fetched yearly withdraw status success by card number (from DB)",
		zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		zap.Int("year", req.Year))

	return dbRows, nil
}

func (s *withdrawStatsByCardStatusService) FindMonthWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.MonthStatusWithdrawCardNumber) ([]*salemongo.WithdrawMonthlyStatusFailedByCardRow, error) {
	const method = "FindMonthWithdrawStatusFailedByCardNumber"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		attribute.Int("year", req.Year),
		attribute.Int("month", req.Month))

	defer func() {
		end(status)
	}()

	if dbRows, found := s.cache.GetCachedMonthWithdrawStatusFailedByCardNumber(ctx, req); found {
		logSuccess("Successfully fetched monthly withdraw status failed by card number (from cache)",
			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year),
			zap.Int("month", req.Month))
		return dbRows, nil
	}

	dbRows, err := s.repository.GetMonthWithdrawStatusFailedByCardNumber(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[[]*salemongo.WithdrawMonthlyStatusFailedByCardRow](
			s.logger,
			withdraw_errors.ErrFailedFindMonthWithdrawStatusFailed,
			method,
			span,

			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year),
			zap.Int("month", req.Month),
		)
	}

	s.cache.SetCachedMonthWithdrawStatusFailedByCardNumber(ctx, req, dbRows)

	logSuccess("Successfully fetched monthly withdraw status failed by card number (from DB)",
		zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		zap.Int("year", req.Year),
		zap.Int("month", req.Month))

	return dbRows, nil
}

func (s *withdrawStatsByCardStatusService) FindYearlyWithdrawStatusFailedByCardNumber(ctx context.Context, req *requests.YearStatusWithdrawCardNumber) ([]*salemongo.WithdrawYearlyStatusFailedByCardRow, error) {
	const method = "FindYearlyWithdrawStatusFailedByCardNumber"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		attribute.Int("year", req.Year))

	defer func() {
		end(status)
	}()

	if dbRows, found := s.cache.GetCachedYearlyWithdrawStatusFailedByCardNumber(ctx, req); found {
		logSuccess("Successfully fetched yearly withdraw status failed by card number (from cache)",
			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year))
		return dbRows, nil
	}

	dbRows, err := s.repository.GetYearlyWithdrawStatusFailedByCardNumber(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[[]*salemongo.WithdrawYearlyStatusFailedByCardRow](
			s.logger,
			withdraw_errors.ErrFailedFindYearWithdrawStatusFailed,
			method,
			span,

			zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
			zap.Int("year", req.Year),
		)
	}

	s.cache.SetCachedYearlyWithdrawStatusFailedByCardNumber(ctx, req, dbRows)

	logSuccess("Successfully fetched yearly withdraw status failed by card number (from DB)",
		zap.String("card_number", observability.MaskIdentifier(req.CardNumber)),
		zap.Int("year", req.Year))

	return dbRows, nil
}
