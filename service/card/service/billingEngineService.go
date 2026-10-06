package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	sharederrorhandler "github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type BillingEngineServiceDeps struct {
	BillingCycleRepository repository.BillingCycleRepository
	Kafka                  *kafka.Kafka
	Logger                 logger.LoggerInterface
	Observability          observability.TraceLoggerObservability
	Now                    func() time.Time
}

type billingEngineService struct {
	billingCycleRepository repository.BillingCycleRepository
	kafka                  *kafka.Kafka
	logger                 logger.LoggerInterface
	observability          observability.TraceLoggerObservability
	now                    func() time.Time
}

func NewBillingEngineService(deps *BillingEngineServiceDeps) BillingEngineService {
	return &billingEngineService{
		billingCycleRepository: deps.BillingCycleRepository,
		kafka:                  deps.Kafka,
		logger:                 deps.Logger,
		observability:          deps.Observability,
		now: func() time.Time {
			if deps.Now != nil {
				return deps.Now()
			}
			return time.Now()
		},
	}
}

func (s *billingEngineService) TriggerBillingCycle(ctx context.Context, billingCycleDay int) (int, error) {
	const method = "TriggerBillingCycle"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("billing_cycle_day", billingCycleDay))
	defer func() { end(status, "internal") }()

	if billingCycleDay < 1 || billingCycleDay > 31 {
		status = "error"
		return sharederrorhandler.HandleError[int](s.logger,
			sharedErrors.NewBadRequestError("billing cycle day must be between 1 and 31"),
			method, span, zap.Int("billing_cycle_day", billingCycleDay))
	}

	now := s.now()
	cycleStart, cycleEnd := billingPeriod(now, billingCycleDay)
	created, err := s.billingCycleRepository.CreateBillingCycles(ctx, cycleStart, cycleEnd, cycleEnd)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[int](s.logger, err, method, span)
	}

	if s.kafka != nil && created > 0 {
		payload, marshalErr := json.Marshal(map[string]any{
			"billing_cycle_day": billingCycleDay,
			"cycle_start":       cycleStart,
			"cycle_end":         cycleEnd,
			"created":           created,
		})
		if marshalErr != nil {
			s.logger.Error("failed to marshal billing cycle event", zap.Error(marshalErr))
		} else if sendErr := s.kafka.SendMessage("card.statement.generated", strconv.Itoa(billingCycleDay), payload); sendErr != nil {
			s.logger.Error("failed to publish billing cycle event", zap.Error(sendErr))
		}
	}

	logSuccess("Successfully triggered billing cycle", zap.Int("created", created))
	return created, nil
}

func (s *billingEngineService) GetStatement(ctx context.Context, cardNumber string) (*salemongo.BillingCycle, error) {
	const method = "GetStatement"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", cardNumber))
	defer func() { end(status, "internal") }()

	cardNumber, err := normalizeCardNumber(cardNumber)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.BillingCycle](s.logger, err, method, span)
	}

	cycles, err := s.billingCycleRepository.GetBillingCyclesByCardNumber(ctx, cardNumber)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.BillingCycle](s.logger, err, method, span)
	}
	if len(cycles) == 0 {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.BillingCycle](s.logger,
			sharedErrors.ErrNotFoundResponse("Billing statement"), method, span)
	}

	logSuccess("Successfully retrieved statement", zap.Int("billing_id", int(cycles[0].BillingID)))
	return cycles[0], nil
}

func (s *billingEngineService) GetStatementsByCard(ctx context.Context, cardNumber string, page, pageSize int) ([]*salemongo.BillingCycle, error) {
	const method = "GetStatementsByCard"
	page, pageSize = normalizePagination(page, pageSize)

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", cardNumber))
	defer func() { end(status, "internal") }()

	cardNumber, err := normalizeCardNumber(cardNumber)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[[]*salemongo.BillingCycle](s.logger, err, method, span)
	}

	cycles, err := s.billingCycleRepository.GetBillingCyclesByCardNumber(ctx, cardNumber)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[[]*salemongo.BillingCycle](s.logger, err, method, span)
	}

	start := (page - 1) * pageSize
	if start >= len(cycles) {
		return []*salemongo.BillingCycle{}, nil
	}
	endIndex := start + pageSize
	if endIndex > len(cycles) {
		endIndex = len(cycles)
	}

	result := cycles[start:endIndex]
	logSuccess("Successfully retrieved billing statements", zap.Int("count", len(result)))
	return result, nil
}

func (s *billingEngineService) GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*salemongo.BillingCycle, error) {
	const method = "GetBillingCyclesByCardNumber"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("card_number", cardNumber))
	defer func() { end(status, "internal") }()

	cardNumber, err := normalizeCardNumber(cardNumber)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[[]*salemongo.BillingCycle](s.logger, err, method, span)
	}

	cycles, err := s.billingCycleRepository.GetBillingCyclesByCardNumber(ctx, cardNumber)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[[]*salemongo.BillingCycle](s.logger, err, method, span)
	}

	logSuccess("Successfully retrieved billing cycles", zap.Int("count", len(cycles)))
	return cycles, nil
}

func normalizeCardNumber(cardNumber string) (string, error) {
	cardNumber = strings.TrimSpace(cardNumber)
	if cardNumber == "" {
		return "", sharedErrors.NewBadRequestError("card number is required")
	}
	return cardNumber, nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func billingPeriod(now time.Time, billingDay int) (time.Time, time.Time) {
	location := now.Location()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	currentEnd := dateInMonth(monthStart, billingDay)
	if now.Before(currentEnd) {
		currentEnd = dateInMonth(monthStart.AddDate(0, -1, 0), billingDay)
	}
	previousMonth := time.Date(currentEnd.Year(), currentEnd.Month()-1, 1, 0, 0, 0, 0, location)
	previousStart := dateInMonth(previousMonth, billingDay)
	return previousStart, currentEnd
}

func dateInMonth(month time.Time, day int) time.Time {
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)
	lastDay := nextMonth.AddDate(0, 0, -1).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(monthStart.Year(), monthStart.Month(), day, 0, 0, 0, 0, monthStart.Location())
}
