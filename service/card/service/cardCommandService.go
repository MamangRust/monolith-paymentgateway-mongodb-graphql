package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	mencache "github.com/MamangRust/monolith-payment-gateway-card/redis"
	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/events"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type cardCommandServiceDeps struct {
	Cache                 mencache.CardCommandCache
	Kafka                 *kafka.Kafka
	UserAdapter           adapter.UserAdapter
	CardCommandRepository repository.CardCommandRepository
	BillingEngine         BillingEngineService
	BillingCycleDay       int
	Logger                logger.LoggerInterface
	Observability         observability.TraceLoggerObservability
}

type cardCommandService struct {
	cache                 mencache.CardCommandCache
	kafka                 *kafka.Kafka
	userAdapter           adapter.UserAdapter
	cardCommandRepository repository.CardCommandRepository
	billingEngine         BillingEngineService
	billingCycleDay       int
	logger                logger.LoggerInterface
	observability         observability.TraceLoggerObservability
}

func NewCardCommandService(params *cardCommandServiceDeps) CardCommandService {
	return &cardCommandService{
		cache:                 params.Cache,
		kafka:                 params.Kafka,
		userAdapter:           params.UserAdapter,
		cardCommandRepository: params.CardCommandRepository,
		billingEngine:         params.BillingEngine,
		billingCycleDay:       params.BillingCycleDay,
		logger:                params.Logger,
		observability:         params.Observability,
	}
}

func (s *cardCommandService) CreateCard(ctx context.Context, request *requests.CreateCardRequest) (*salemongo.Card, error) {
	const method = "CreateCard"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.userAdapter.FindById(ctx, request.UserID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("user_id", request.UserID))
	}

	res, err := s.cardCommandRepository.CreateCard(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("user_id", request.UserID))
	}

	if s.kafka != nil {
		go func() {
			saldoPayload := map[string]any{
				"card_number":   res.CardNumber,
				"total_balance": 0,
			}
			payloadBytes, err := json.Marshal(saldoPayload)
			if err != nil {
				s.logger.Error("failed to marshal saldo payload for new card", zap.Error(err))
				return
			}
			_ = s.kafka.SendMessage("saldo-service-topic-create-saldo", strconv.Itoa(int(res.CardID)), payloadBytes)

			statsEvent := events.CardEvent{
				CardID:       uint64(res.CardID),
				UserID:       uint64(request.UserID),
				CardNumber:   res.CardNumber,
				CardType:     res.CardType,
				CardProvider: res.CardProvider,
				Status:       "active",
				CreatedAt:    time.Now(),
			}
			statsPayloadByte, err := json.Marshal(statsEvent)
			if err != nil {
				s.logger.Error("failed to marshal card stats event", zap.Error(err))
			} else {
				_ = s.kafka.SendMessage("stats-topic-card-events", strconv.Itoa(int(res.CardID)), statsPayloadByte)
			}
		}()
	}

	logSuccess("Successfully created card", zap.Int("card.id", int(res.CardID)))
	return res, nil
}

func (s *cardCommandService) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*salemongo.Card, error) {
	const method = "UpdateCard"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.userAdapter.FindById(ctx, request.UserID)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("user_id", request.UserID))
	}

	res, err := s.cardCommandRepository.UpdateCard(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("card_id", request.CardID))
	}

	s.cache.DeleteCardCommandCache(ctx, request.CardID)
	logSuccess("Successfully updated card", zap.Int("card.id", int(res.CardID)))
	return res, nil
}

func (s *cardCommandService) TrashedCard(ctx context.Context, card_id int) (*salemongo.Card, error) {
	const method = "TrashedCard"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("card_id", card_id))
	defer func() { end(status, "grpc") }()

	res, err := s.cardCommandRepository.TrashedCard(ctx, card_id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("card_id", card_id))
	}

	s.cache.DeleteCardCommandCache(ctx, card_id)
	logSuccess("Successfully trashed card", zap.Int("card_id", int(res.CardID)))
	return res, nil
}

func (s *cardCommandService) RestoreCard(ctx context.Context, card_id int) (*salemongo.Card, error) {
	const method = "RestoreCard"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("card_id", card_id))
	defer func() { end(status, "grpc") }()

	res, err := s.cardCommandRepository.RestoreCard(ctx, card_id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("card_id", card_id))
	}

	s.cache.DeleteCardCommandCache(ctx, card_id)
	logSuccess("Successfully restored card", zap.Int("card_id", int(res.CardID)))
	return res, nil
}

func (s *cardCommandService) DeleteCardPermanent(ctx context.Context, card_id int) (bool, error) {
	const method = "DeleteCardPermanent"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("card_id", card_id))
	defer func() { end(status, "grpc") }()

	_, err := s.cardCommandRepository.DeleteCardPermanent(ctx, card_id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](s.logger, err, method, span, zap.Int("card_id", card_id))
	}

	s.cache.DeleteCardCommandCache(ctx, card_id)
	logSuccess("Successfully deleted card permanently", zap.Int("card_id", card_id))
	return true, nil
}

func (s *cardCommandService) RestoreAllCard(ctx context.Context) (bool, error) {
	const method = "RestoreAllCard"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.cardCommandRepository.RestoreAllCard(ctx)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](s.logger, sharedErrors.ErrFailed("restore all cards"), method, span)
	}
	logSuccess("Successfully restored all cards")
	return true, nil
}

func (s *cardCommandService) DeleteAllCardPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllCardPermanent"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.cardCommandRepository.DeleteAllCardPermanent(ctx)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](s.logger, sharedErrors.ErrFailed("delete all cards permanently"), method, span)
	}
	logSuccess("Successfully deleted all cards permanently")
	return true, nil
}

func (s *cardCommandService) ToggleCardStatus(ctx context.Context, request *requests.ToggleCardStatusRequest) (*salemongo.Card, error) {
	const method = "ToggleCardStatus"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("card_id", request.CardID))
	defer func() { end(status, "grpc") }()

	res, err := s.cardCommandRepository.ToggleCardStatus(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("card_id", request.CardID))
	}
	s.cache.DeleteCardCommandCache(ctx, request.CardID)
	logSuccess("Successfully toggled card status", zap.Int("card_id", request.CardID))
	return res, nil
}

func (s *cardCommandService) UpdateCreditLimit(ctx context.Context, request *requests.UpdateCreditLimitRequest) (*salemongo.Card, error) {
	const method = "UpdateCreditLimit"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("card_id", request.CardID))
	defer func() { end(status, "grpc") }()

	res, err := s.cardCommandRepository.UpdateCreditLimit(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("card_id", request.CardID))
	}
	s.cache.DeleteCardCommandCache(ctx, request.CardID)
	logSuccess("Successfully updated credit limit", zap.Int("card_id", request.CardID))
	return res, nil
}

func (s *cardCommandService) RedeemPoints(ctx context.Context, request *requests.RedeemPointsRequest) (*salemongo.Card, error) {
	const method = "RedeemPoints"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("card_id", request.CardID))
	defer func() { end(status, "grpc") }()

	res, err := s.cardCommandRepository.RedeemPoints(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*salemongo.Card](s.logger, err, method, span, zap.Int("card_id", request.CardID))
	}
	s.cache.DeleteCardCommandCache(ctx, request.CardID)
	logSuccess("Successfully redeemed reward points", zap.Int("card_id", request.CardID))
	return res, nil
}

func (s *cardCommandService) ProcessBillingCycles(ctx context.Context) error {
	if s.billingEngine == nil {
		return sharedErrors.ErrFailed("process billing cycles")
	}
	billingCycleDay := s.billingCycleDay
	if billingCycleDay == 0 {
		billingCycleDay = 1
	}
	_, err := s.billingEngine.TriggerBillingCycle(ctx, billingCycleDay)
	return err
}
