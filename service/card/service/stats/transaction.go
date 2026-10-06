package cardstatsservice

import (
	"context"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"

	cardstatsmencache "github.com/MamangRust/monolith-payment-gateway-card/redis/stats"
	repository "github.com/MamangRust/monolith-payment-gateway-card/repository/stats"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	sharederrorhandler "github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	card_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/card_errors/service"

	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type cardStatsTransactionService struct {
	cache cardstatsmencache.CardStatsTransactionCache

	repository repository.CardStatsTransactionRepository

	logger logger.LoggerInterface

	observability observability.TraceLoggerObservability
}

type cardStatsTransactionServiceDeps struct {
	Cache cardstatsmencache.CardStatsTransactionCache

	Repository repository.CardStatsTransactionRepository

	Logger logger.LoggerInterface

	Observability observability.TraceLoggerObservability
}

func NewCardStatsTransactionService(params *cardStatsTransactionServiceDeps) CardStatsTransactionService {

	return &cardStatsTransactionService{
		cache:         params.Cache,
		repository:    params.Repository,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *cardStatsTransactionService) FindMonthlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.MonthlyTransactionRow, error) {
	const method = "FindMonthlyTransactionAmount"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.cache.GetMonthlyTransactionCache(ctx, year); found {
		logSuccess("Monthly transaction amount cache hit", zap.Int("year", year))
		return data, nil
	}

	res, err := s.repository.GetMonthlyTransactionAmount(ctx, year)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[[]*salemongo.MonthlyTransactionRow](
			s.logger,
			card_errors.ErrFailedFindMonthlyTransactionAmount,
			method,
			span,
			zap.Int("year", year),
		)
	}

	s.cache.SetMonthlyTransactionCache(ctx, year, res)

	logSuccess("Monthly transaction amount retrieved successfully",
		zap.Int("year", year),
		zap.Int("result_count", len(res)),
	)

	return res, nil
}

func (s *cardStatsTransactionService) FindYearlyTransactionAmount(ctx context.Context, year int) ([]*salemongo.YearlyTransactionRow, error) {
	const method = "FindYearlyTransactionAmount"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.cache.GetYearlyTransactionCache(ctx, year); found {
		logSuccess("Yearly transaction amount cache hit", zap.Int("year", year))
		return data, nil
	}

	res, err := s.repository.GetYearlyTransactionAmount(ctx, year)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[[]*salemongo.YearlyTransactionRow](
			s.logger,
			card_errors.ErrFailedFindYearlyTransactionAmount,
			method,
			span,
			zap.Int("year", year),
		)
	}

	s.cache.SetYearlyTransactionCache(ctx, year, res)

	logSuccess("Yearly transaction amount retrieved successfully",
		zap.Int("year", year),
		zap.Int("result_count", len(res)),
	)

	return res, nil
}
