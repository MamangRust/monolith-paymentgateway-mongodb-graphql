package merchantstatsbymerchantservice

import (
	"context"

	mencache "github.com/MamangRust/monolith-payment-gateway-merchant/redis/statsbymerchant"
	repository "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbymerchant"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	merchant_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/merchant_errors/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type MerchantStatsByMerchantMethodDeps struct {
	Cache mencache.MerchantStatsMethodByMerchantCache

	Repository repository.MerchantStatsMethodByMerchantRepository

	Logger logger.LoggerInterface

	Observability observability.TraceLoggerObservability
}

type merchantStatsByMerchantMethodService struct {
	cache mencache.MerchantStatsMethodByMerchantCache

	repository repository.MerchantStatsMethodByMerchantRepository

	logger logger.LoggerInterface

	observability observability.TraceLoggerObservability
}

func NewMerchantStatsByMerchantMethodService(params *MerchantStatsByMerchantMethodDeps) *merchantStatsByMerchantMethodService {
	return &merchantStatsByMerchantMethodService{
		cache:         params.Cache,
		repository:    params.Repository,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *merchantStatsByMerchantMethodService) FindMonthlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantMonthlyPaymentMethodRow, error) {
	const method = "FindMonthlyPaymentMethodByMerchants"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("year", req.Year),
		attribute.Int("merchant_id", req.MerchantID))

	defer func() {
		end(status)
	}()

	if cachedMerchant, found := s.cache.GetMonthlyPaymentMethodByMerchantsCache(ctx, req); found {
		logSuccess("Successfully fetched merchant from cache", zap.Int("merchant_id", req.MerchantID))
		return cachedMerchant, nil
	}

	dbRows, err := s.repository.GetMonthlyPaymentMethodByMerchants(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[[]*salemongo.MerchantMonthlyPaymentMethodRow](
			s.logger,
			merchant_errors.ErrFailedFindMonthlyPaymentMethodByMerchants,
			method,
			span,

			zap.Int("merchant_id", req.MerchantID),
			zap.Int("year", req.Year),
		)
	}

	s.cache.SetMonthlyPaymentMethodByMerchantsCache(ctx, req, dbRows)

	logSuccess("Successfully found monthly payment methods by merchant (from DB)", zap.Int("merchant_id", req.MerchantID), zap.Int("year", req.Year))
	return dbRows, nil
}

func (s *merchantStatsByMerchantMethodService) FindYearlyPaymentMethodByMerchants(ctx context.Context, req *requests.MonthYearPaymentMethodMerchant) ([]*salemongo.MerchantYearlyPaymentMethodRow, error) {
	const method = "FindYearlyPaymentMethodByMerchants"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("year", req.Year),
		attribute.Int("merchant_id", req.MerchantID))

	defer func() {
		end(status)
	}()

	if cachedMerchant, found := s.cache.GetYearlyPaymentMethodByMerchantsCache(ctx, req); found {
		logSuccess("Successfully fetched merchant from cache", zap.Int("merchant_id", req.MerchantID))
		return cachedMerchant, nil
	}

	dbRows, err := s.repository.GetYearlyPaymentMethodByMerchants(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[[]*salemongo.MerchantYearlyPaymentMethodRow](
			s.logger,
			merchant_errors.ErrFailedFindYearlyPaymentMethodByMerchants,
			method,
			span,

			zap.Int("merchant_id", req.MerchantID),
			zap.Int("year", req.Year),
		)
	}

	s.cache.SetYearlyPaymentMethodByMerchantsCache(ctx, req, dbRows)

	logSuccess("Successfully found yearly payment methods by merchant (from DB)", zap.Int("merchant_id", req.MerchantID), zap.Int("year", req.Year))
	return dbRows, nil
}
