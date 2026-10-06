package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/email"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/async"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/events"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/validation"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/idempotency"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"github.com/MamangRust/monolith-payment-gateway-shared/security"
	mencache "github.com/MamangRust/monolith-payment-gateway-topup/redis"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TopupCommandDeps groups dependencies for top-up command service.
type TopupCommandDeps struct {
	Kafka                  *kafka.Kafka
	Cache                  mencache.TopupCommandCache
	CardAdapter            adapter.CardAdapter
	TopupQueryRepository   repository.TopupQueryRepository
	TopupCommandRepository repository.TopupCommandRepository
	SaldoAdapter           adapter.SaldoAdapter
	IdempotencyStore       idempotency.Store
	OutboxStore            repository.OutboxRepository
	Logger                 logger.LoggerInterface
	Observability          observability.TraceLoggerObservability
}

// topupCommandService handles top-up command operations.
type topupCommandService struct {
	kafka                  *kafka.Kafka
	cache                  mencache.TopupCommandCache
	topupQueryRepository   repository.TopupQueryRepository
	cardAdapter            adapter.CardAdapter
	topupCommandRepository repository.TopupCommandRepository
	saldoAdapter           adapter.SaldoAdapter
	idempotencyStore       idempotency.Store
	outboxStore            repository.OutboxRepository
	logger                 logger.LoggerInterface
	observability          observability.TraceLoggerObservability
}

func NewTopupCommandService(
	params *TopupCommandDeps,
) TopupCommandService {
	return &topupCommandService{
		kafka:                  params.Kafka,
		cache:                  params.Cache,
		topupQueryRepository:   params.TopupQueryRepository,
		topupCommandRepository: params.TopupCommandRepository,
		saldoAdapter:           params.SaldoAdapter,
		idempotencyStore:       params.IdempotencyStore,
		outboxStore:            params.OutboxStore,
		cardAdapter:            params.CardAdapter,
		logger:                 params.Logger,
		observability:          params.Observability,
	}
}

// idempotencyReplayWindow bounds how long a concurrent duplicate waits for the
// in-flight request holding the same Idempotency-Key to finish so it can replay
// the same response instead of failing with a spurious 409.
const (
	idempotencyReplayWindow       = 3 * time.Second
	idempotencyReplayPollInterval = 10 * time.Millisecond
)

// topupUpdateMaxAttempts bounds the compare-and-set retries used when concurrent
// updates race on the same topup amount.
const topupUpdateMaxAttempts = 8

// awaitIdempotentOutcome loads the record owned by a concurrent duplicate and
// waits for it to reach a terminal state, returning a conflict error when the
// stored request hash differs from the incoming one.
func (s *topupCommandService) awaitIdempotentOutcome(ctx context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	deadline := time.Now().Add(idempotencyReplayWindow)

	for {
		record, err := s.idempotencyStore.Get(ctx, scope, key)
		if err != nil {
			return nil, err
		}
		if record == nil {
			return nil, idempotency.ProcessingError()
		}
		if record.RequestHash != requestHash {
			return nil, idempotency.ConflictError()
		}
		if record.Status != idempotency.StatusProcessing || !time.Now().Before(deadline) {
			return record, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(idempotencyReplayPollInterval):
		}
	}
}

func (s *topupCommandService) CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*salemongo.Topup, error) {
	const method = "CreateTopup"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	// Idempotency guard
	var idemHash string
	if request.IdempotencyKey != "" && s.idempotencyStore != nil {
		idemHash = idempotency.HashRequest(request)
		if _, err := s.idempotencyStore.Claim(ctx, "topup", request.IdempotencyKey, idemHash); err == nil {
			defer func() {
				if status == "error" {
					if relErr := s.idempotencyStore.Release(ctx, "topup", request.IdempotencyKey, idemHash); relErr != nil {
						s.logger.Error("idempotency: failed to release key", zap.Error(relErr), zap.String("idempotency_key", request.IdempotencyKey))
					}
				}
			}()
		} else {
			existing, gErr := s.awaitIdempotentOutcome(ctx, "topup", request.IdempotencyKey, idemHash)
			if gErr != nil {
				status = "error"
				return errorhandler.HandleError[*salemongo.Topup](s.logger, gErr, method, span, zap.String("idempotency_key", request.IdempotencyKey))
			}
			if existing.Status == idempotency.StatusSuccess {
				var topup salemongo.Topup
				if uErr := json.Unmarshal(existing.ResponseJSON, &topup); uErr != nil {
					status = "error"
					return errorhandler.HandleError[*salemongo.Topup](s.logger, uErr, method, span, zap.String("idempotency_key", request.IdempotencyKey))
				}
				logSuccess("Replayed idempotent topup response", zap.String("idempotency_key", request.IdempotencyKey))
				return &topup, nil
			}
			status = "error"
			return errorhandler.HandleError[*salemongo.Topup](s.logger, idempotency.ProcessingError(), method, span, zap.String("idempotency_key", request.IdempotencyKey))
		}
	}

	if err := validation.ValidateAmount(request.TopupAmount); err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	cardIdentity, err := s.cardAdapter.FindUserCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}
	card := cardIdentity

	topup, err := s.topupCommandRepository.CreateTopup(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	creditedSaldo, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber:  request.CardNumber,
		Amount:      request.TopupAmount,
		OperationID: "topup:" + strconv.Itoa(int(topup.TopupID)) + ":credit",
		SourceType:  "topup",
		SourceID:    strconv.Itoa(int(topup.TopupID)),
	})
	if err != nil {
		status = "error"
		s.markTopupAsFailed(ctx, int(topup.TopupID), method, span)
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}
	newBalance := int(creditedSaldo.TotalBalance)

	_, err = s.cardAdapter.UpdateCard(ctx, &requests.UpdateCardRequest{
		CardID:       int(card.CardID),
		UserID:       int(card.UserID),
		CardType:     card.CardType,
		ExpireDate:   card.ExpireDate,
		CVV:          card.Cvv,
		CardProvider: card.CardProvider,
	})
	if err != nil {
		status = "error"
		_, rollbackErr := s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      request.TopupAmount,
			OperationID: "topup:" + strconv.Itoa(int(topup.TopupID)) + ":rollback",
			SourceType:  "topup_compensation",
			SourceID:    strconv.Itoa(int(topup.TopupID)),
		})
		if rollbackErr != nil {
			return errorhandler.HandleError[*salemongo.Topup](s.logger, rollbackErr, method, span, zap.String("rollback_for", "saldo"))
		}
		s.markTopupAsFailed(ctx, int(topup.TopupID), method, span)
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	updatedTopup, err := s.topupCommandRepository.UpdateTopupStatus(ctx, &requests.UpdateTopupStatus{
		TopupID: int(topup.TopupID),
		Status:  "success",
	})
	if err != nil {
		status = "error"
		_, rollbackErr := s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      request.TopupAmount,
			OperationID: "topup:" + strconv.Itoa(int(topup.TopupID)) + ":rollback",
			SourceType:  "topup_compensation",
			SourceID:    strconv.Itoa(int(topup.TopupID)),
		})
		if rollbackErr != nil {
			return errorhandler.HandleError[*salemongo.Topup](s.logger, rollbackErr, method, span, zap.String("rollback_for", "saldo"))
		}
		s.markTopupAsFailed(ctx, int(topup.TopupID), method, span)
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.Int("topup_id", int(topup.TopupID)))
	}

	s.enqueueTopupEvents(ctx, topup, updatedTopup, card, request, newBalance)

	logSuccess("Topup created successfully", zap.String("cardNumber", security.MaskCardNumber(request.CardNumber)), zap.Int("topupID", int(topup.TopupID)), zap.Float64("topupAmount", float64(request.TopupAmount)))

	if request.IdempotencyKey != "" && s.idempotencyStore != nil {
		if respBytes, mErr := json.Marshal(updatedTopup); mErr == nil {
			resourceID := updatedTopup.TopupID
			if cErr := s.idempotencyStore.Complete(ctx, "topup", request.IdempotencyKey, idemHash, idempotency.Outcome{
				Status:       idempotency.StatusSuccess,
				ResponseJSON: respBytes,
				ResourceID:   &resourceID,
			}); cErr != nil {
				s.logger.Error("idempotency: failed to complete key", zap.Error(cErr), zap.String("idempotency_key", request.IdempotencyKey))
			}
		}
	}

	return updatedTopup, nil
}

func (s *topupCommandService) UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*salemongo.Topup, error) {
	const method = "UpdateTopup"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.cardAdapter.FindCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		status = "error"
		s.markTopupAsFailed(ctx, *request.TopupID, method, span)
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	// The amount is swapped with compare-and-set so concurrent updates settle on
	// the winner's value instead of each applying a delta from a stale read.
	applied := false
	var previousAmount int
	for attempt := 0; attempt < topupUpdateMaxAttempts && !applied; attempt++ {
		existingTopup, findErr := s.topupQueryRepository.FindById(ctx, *request.TopupID)
		if findErr != nil {
			status = "error"
			s.markTopupAsFailed(ctx, *request.TopupID, method, span)
			return errorhandler.HandleError[*salemongo.Topup](s.logger, findErr, method, span, zap.Int("topup_id", *request.TopupID))
		}

		previousAmount = int(existingTopup.TopupAmount)
		swapped, swapErr := s.topupCommandRepository.CompareAndSwapTopupAmount(ctx, *request.TopupID, previousAmount, request.TopupAmount)
		if swapErr != nil {
			status = "error"
			s.markTopupAsFailed(ctx, *request.TopupID, method, span)
			return errorhandler.HandleError[*salemongo.Topup](s.logger, swapErr, method, span, zap.Int("topup_id", *request.TopupID))
		}
		if !swapped {
			continue
		}

		topupDifference := request.TopupAmount - previousAmount
		if topupDifference > 0 {
			_, err = s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
				CardNumber:  request.CardNumber,
				Amount:      topupDifference,
				OperationID: "topup:update:" + strconv.Itoa(*request.TopupID) + ":credit",
				SourceType:  "topup_update",
				SourceID:    strconv.Itoa(*request.TopupID),
			})
		} else if topupDifference < 0 {
			_, err = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
				CardNumber:  request.CardNumber,
				Amount:      -topupDifference,
				OperationID: "topup:update:" + strconv.Itoa(*request.TopupID) + ":debit",
				SourceType:  "topup_update",
				SourceID:    strconv.Itoa(*request.TopupID),
			})
		}
		if err != nil {
			status = "error"
			_, rollbackErr := s.topupCommandRepository.CompareAndSwapTopupAmount(ctx, *request.TopupID, request.TopupAmount, previousAmount)
			if rollbackErr != nil {
				return errorhandler.HandleError[*salemongo.Topup](s.logger, rollbackErr, method, span, zap.Int("topup_id", *request.TopupID))
			}
			s.markTopupAsFailed(ctx, *request.TopupID, method, span)
			return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
		}

		// Persist the non-amount metadata of the update.
		if _, err = s.topupCommandRepository.UpdateTopup(ctx, request); err != nil {
			status = "error"
			s.markTopupAsFailed(ctx, *request.TopupID, method, span)
			return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.Int("topup_id", *request.TopupID))
		}

		applied = true
	}

	if !applied {
		status = "error"
		s.markTopupAsFailed(ctx, *request.TopupID, method, span)
		return errorhandler.HandleError[*salemongo.Topup](s.logger, sharedErrors.NewConflictError("topup was updated concurrently, retry the request"), method, span, zap.Int("topup_id", *request.TopupID))
	}

	updatedTopup, err := s.topupCommandRepository.UpdateTopupStatus(ctx, &requests.UpdateTopupStatus{
		TopupID: *request.TopupID,
		Status:  "success",
	})
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Topup](s.logger, err, method, span, zap.Int("topup_id", *request.TopupID))
	}

	logSuccess("UpdateTopup process completed", zap.Int("topup_id", *request.TopupID))

	return updatedTopup, nil
}

func (s *topupCommandService) TrashedTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error) {
	const method = "TrashedTopup"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("topup_id", topup_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting TrashedTopup process", zap.Int("topup_id", topup_id))

	res, err := s.topupCommandRepository.TrashedTopup(ctx, topup_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Topup](
			s.logger,
			err,
			method,
			span,

			zap.Int("topup_id", topup_id),
		)
	}

	logSuccess("TrashedTopup process completed", zap.Int("topup_id", topup_id))

	return res, nil
}

func (s *topupCommandService) RestoreTopup(ctx context.Context, topup_id int) (*salemongo.Topup, error) {
	const method = "RestoreTopup"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("topup_id", topup_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting RestoreTopup process", zap.Int("topup_id", topup_id))

	res, err := s.topupCommandRepository.RestoreTopup(ctx, topup_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Topup](
			s.logger,
			err,
			method,
			span,

			zap.Int("topup_id", topup_id),
		)
	}

	logSuccess("RestoreTopup process completed", zap.Int("topup_id", topup_id))

	return res, nil
}

func (s *topupCommandService) DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error) {
	const method = "DeleteTopupPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("topup_id", topup_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting DeleteTopupPermanent process", zap.Int("topup_id", topup_id))

	_, err := s.topupCommandRepository.DeleteTopupPermanent(ctx, topup_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			err,
			method,
			span,

			zap.Int("topup_id", topup_id),
		)
	}

	logSuccess("DeleteTopupPermanent process completed", zap.Int("topup_id", topup_id))

	return true, nil
}

func (s *topupCommandService) RestoreAllTopup(ctx context.Context) (bool, error) {
	const method = "RestoreAllTopup"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Restoring all topups")

	_, err := s.topupCommandRepository.RestoreAllTopup(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("restore all topups"),
			method,
			span,
		)
	}

	logSuccess("Successfully restored all topups")
	return true, nil
}

func (s *topupCommandService) DeleteAllTopupPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllTopupPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Permanently deleting all topups")

	_, err := s.topupCommandRepository.DeleteAllTopupPermanent(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("delete all topups permanently"),
			method,
			span,
		)
	}

	logSuccess("Successfully deleted all topups permanently")
	return true, nil
}

func (s *topupCommandService) enqueueTopupEvents(ctx context.Context, topup *salemongo.Topup, updatedTopup *salemongo.Topup, card *salemongo.CardByEmailRow, request *requests.CreateTopupRequest, newBalance int) {
	if s.outboxStore == nil {
		return
	}
	topupID := strconv.Itoa(int(topup.TopupID))

	// Email enrichment
	if card.Email != "" {
		htmlBody, err := email.GenerateEmailHTML(map[string]string{
			"Title":   "Topup Successful",
			"Message": fmt.Sprintf("Your topup of %d has been processed successfully.", request.TopupAmount),
			"Button":  "View History",
			"Link":    "https://sanedge.example.com/topup/history",
		})
		if err == nil {
			emailPayload, _ := json.Marshal(map[string]any{"email": card.Email, "subject": "Topup Successful - SanEdge", "body": htmlBody})
			if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{
				AggregateType: "topup", AggregateID: topupID,
				EventType: "topup.created", Payload: emailPayload,
			}); iErr != nil {
				s.logger.Error("outbox: failed to enqueue topup email event", zap.Error(iErr))
			}
		}
	}

	// Stats event
	statsEvent := events.TopupEvent{
		TopupID:       uint64(updatedTopup.TopupID),
		TopupNo:       updatedTopup.TopupNo.String(),
		CardNumber:    request.CardNumber,
		CardType:      card.CardType,
		CardProvider:  card.CardProvider,
		Amount:        int64(request.TopupAmount),
		PaymentMethod: request.TopupMethod,
		Status:        "success",
		CreatedAt:     time.Now(),
	}
	statsBytes, _ := json.Marshal(statsEvent)
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{
		AggregateType: "topup", AggregateID: topupID,
		EventType: "topup.stats", Payload: statsBytes,
	}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue topup stats event", zap.Error(iErr))
	}

	// Saldo event
	saldoEvent := events.SaldoEvent{CardNumber: request.CardNumber, TotalBalance: int64(newBalance), CreatedAt: time.Now()}
	saldoBytes, _ := json.Marshal(saldoEvent)
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{
		AggregateType: "topup", AggregateID: request.CardNumber,
		EventType: "topup.saldo", Payload: saldoBytes,
	}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue topup saldo event", zap.Error(iErr))
	}
}

func (s *topupCommandService) markTopupAsFailed(ctx context.Context, topupID int, method string, span trace.Span) {
	req := requests.UpdateTopupStatus{
		TopupID: topupID,
		Status:  "failed",
	}
	async.RunWithTimeout(5*time.Second, func(ctx context.Context) {
		if _, err := s.topupCommandRepository.UpdateTopupStatus(ctx, &req); err != nil {
			s.logger.Error("compensation: failed to mark topup as failed", zap.Error(err), zap.Int("topup_id", topupID), zap.String("method", method))
		}
	})
}
