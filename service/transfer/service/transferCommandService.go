package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pb/ai_security"
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
	"github.com/MamangRust/monolith-payment-gateway-shared/state"
	mencache "github.com/MamangRust/monolith-payment-gateway-transfer/redis"
	"github.com/MamangRust/monolith-payment-gateway-transfer/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type TransferCommandDeps struct {
	Kafka *kafka.Kafka
	Cache mencache.TransferCommandCache

	CardAdapter  adapter.CardAdapter
	SaldoAdapter adapter.SaldoAdapter

	TransferQueryRepository   repository.TransferQueryRepository
	TransferCommandRepository repository.TransferCommandRepository
	IdempotencyStore          idempotency.Store
	OutboxStore               repository.OutboxRepository

	Logger           logger.LoggerInterface
	Observability    observability.TraceLoggerObservability
	AISecurityClient ai_security.AISecurityServiceClient
}

type transferCommandService struct {
	kafka *kafka.Kafka
	cache mencache.TransferCommandCache

	cardAdapter  adapter.CardAdapter
	saldoAdapter adapter.SaldoAdapter

	transferQueryRepository   repository.TransferQueryRepository
	transferCommandRepository repository.TransferCommandRepository
	idempotencyStore          idempotency.Store
	outboxStore               repository.OutboxRepository

	logger           logger.LoggerInterface
	observability    observability.TraceLoggerObservability
	aiSecurityClient ai_security.AISecurityServiceClient
}

func NewTransferCommandService(
	params *TransferCommandDeps,
) TransferCommandService {
	return &transferCommandService{
		kafka:                     params.Kafka,
		cache:                     params.Cache,
		cardAdapter:               params.CardAdapter,
		saldoAdapter:              params.SaldoAdapter,
		transferQueryRepository:   params.TransferQueryRepository,
		transferCommandRepository: params.TransferCommandRepository,
		idempotencyStore:          params.IdempotencyStore,
		outboxStore:               params.OutboxStore,
		logger:                    params.Logger,
		observability:             params.Observability,
		aiSecurityClient:          params.AISecurityClient,
	}
}

func (s *transferCommandService) CreateTransaction(ctx context.Context, request *requests.CreateTransferRequest) (*salemongo.Transfer, error) {
	const method = "CreateTransaction"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	// Idempotency guard
	var idemHash string
	if request.IdempotencyKey != "" && s.idempotencyStore != nil {
		idemHash = idempotency.HashRequest(request)
		if _, err := s.idempotencyStore.Claim(ctx, "transfer", request.IdempotencyKey, idemHash); err == nil {
			defer func() {
				if status == "error" {
					if relErr := s.idempotencyStore.Release(ctx, "transfer", request.IdempotencyKey, idemHash); relErr != nil {
						s.logger.Error("idempotency: failed to release key", zap.Error(relErr), zap.String("idempotency_key", request.IdempotencyKey))
					}
				}
			}()
		} else {
			existing, gErr := s.idempotencyStore.Get(ctx, "transfer", request.IdempotencyKey)
			if gErr != nil {
				status = "error"
				return errorhandler.HandleError[*salemongo.Transfer](s.logger, gErr, method, span, zap.String("idempotency_key", request.IdempotencyKey))
			}
			if existing.RequestHash != idemHash {
				status = "error"
				return errorhandler.HandleError[*salemongo.Transfer](s.logger, idempotency.ConflictError(), method, span, zap.String("idempotency_key", request.IdempotencyKey))
			}
			if existing.Status == idempotency.StatusSuccess {
				var row salemongo.Transfer
				if uErr := json.Unmarshal(existing.ResponseJSON, &row); uErr != nil {
					status = "error"
					return errorhandler.HandleError[*salemongo.Transfer](s.logger, uErr, method, span, zap.String("idempotency_key", request.IdempotencyKey))
				}
				logSuccess("Replayed idempotent transfer response", zap.String("idempotency_key", request.IdempotencyKey))
				return &row, nil
			}
			status = "error"
			return errorhandler.HandleError[*salemongo.Transfer](s.logger, idempotency.ProcessingError(), method, span, zap.String("idempotency_key", request.IdempotencyKey))
		}
	}

	if err := validation.ValidateAmount(request.TransferAmount); err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(request.TransferFrom)))
	}
	if err := validation.ValidateTransferParties(request.TransferFrom, request.TransferTo); err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(request.TransferFrom)), zap.String("to_card", security.MaskCardNumber(request.TransferTo)))
	}

	senderCard, err := s.cardAdapter.FindUserCardByCardNumber(ctx, request.TransferFrom)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(request.TransferFrom)))
	}

	_, err = s.cardAdapter.FindCardByCardNumber(ctx, request.TransferTo)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("to_card", security.MaskCardNumber(request.TransferTo)))
	}

	senderSaldo, err := s.saldoAdapter.FindByCardNumber(ctx, request.TransferFrom)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(request.TransferFrom)))
	}

	receiverSaldo, err := s.saldoAdapter.FindByCardNumber(ctx, request.TransferTo)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("to_card", security.MaskCardNumber(request.TransferTo)))
	}

	if int(senderSaldo.TotalBalance) < request.TransferAmount {
		status = "error"
		err := sharedErrors.ErrConflict.WithMessage("insufficient balance for transfer")
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(request.TransferFrom)), zap.Float64("balance", float64(senderSaldo.TotalBalance)), zap.Float64("amount", float64(request.TransferAmount)))
	}

	// AI Security Check
	if s.aiSecurityClient != nil {
		secRes, err := s.aiSecurityClient.VerifySecurity(ctx, &ai_security.SecurityRequest{
			Domain:   ai_security.SecurityDomain_TRANSFER,
			EntityId: request.TransferFrom,
			Amount:   float64(request.TransferAmount),
			Metadata: map[string]string{
				"recipient_card": request.TransferTo,
			},
		})
		if err == nil && !secRes.IsSafe {
			status = "error"
			s.logger.Warn("Transfer blocked by AI Security", zap.String("reason", secRes.Reason))
			return nil, errors.New("security block: " + secRes.Reason)
		}
	}

	transfer, err := s.transferCommandRepository.CreateTransfer(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span)
	}
	transferID := int(transfer.TransferID)
	if ok, gErr := s.transferCommandRepository.GuardStatus(ctx, transferID, state.Pending, state.Processing, ""); gErr != nil || !ok {
		status = "error"
		if gErr == nil {
			gErr = errors.New("transfer already being processed")
		}
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, gErr, method, span, zap.Int("transfer_id", transferID))
	}

	senderDebit, err := s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
		CardNumber:  senderSaldo.CardNumber,
		Amount:      request.TransferAmount,
		OperationID: fmt.Sprintf("transfer:%d:sender-debit", transferID),
		SourceType:  "transfer",
		SourceID:    strconv.Itoa(transferID),
	})
	if err != nil {
		_, _ = s.transferCommandRepository.GuardStatus(ctx, transferID, state.Processing, state.Failed, "debit failed: "+err.Error())
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(request.TransferFrom)))
	}
	senderNewBalance := int(senderDebit.TotalBalance)

	receiverCredit, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber:  receiverSaldo.CardNumber,
		Amount:      request.TransferAmount,
		OperationID: fmt.Sprintf("transfer:%d:receiver-credit", transferID),
		SourceType:  "transfer",
		SourceID:    strconv.Itoa(transferID),
	})
	if err != nil {
		_, _ = s.transferCommandRepository.GuardStatus(ctx, transferID, state.Processing, state.Unknown, "receiver credit outcome ambiguous: "+err.Error())
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("failed_to_credit", "receiver"))
	}
	receiverNewBalance := int(receiverCredit.TotalBalance)

	updatedTransfer, err := s.transferCommandRepository.TransitionStatus(ctx, transferID, state.Processing, state.Success, "")
	if err != nil {
		status = "error"
		s.compensateTransfer(ctx, transferID, request.TransferAmount, senderSaldo.CardNumber, receiverSaldo.CardNumber, method)
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.Int("transfer_id", transferID))
	}

	s.enqueueTransferEvents(ctx, transfer, updatedTransfer, senderCard, request, senderNewBalance, receiverNewBalance)

	logSuccess("Transfer created successfully", zap.Int("transfer_id", int(updatedTransfer.TransferID)), zap.String("from", security.MaskCardNumber(request.TransferFrom)), zap.String("to", security.MaskCardNumber(request.TransferTo)))

	if request.IdempotencyKey != "" && s.idempotencyStore != nil {
		if respBytes, mErr := json.Marshal(updatedTransfer); mErr == nil {
			resourceID := updatedTransfer.TransferID
			if cErr := s.idempotencyStore.Complete(ctx, "transfer", request.IdempotencyKey, idemHash, idempotency.Outcome{
				Status:       idempotency.StatusSuccess,
				ResponseJSON: respBytes,
				ResourceID:   &resourceID,
			}); cErr != nil {
				s.logger.Error("idempotency: failed to complete key", zap.Error(cErr), zap.String("idempotency_key", request.IdempotencyKey))
			}
		}
	}

	return updatedTransfer, nil
}

func (s *transferCommandService) UpdateTransaction(ctx context.Context, request *requests.UpdateTransferRequest) (*salemongo.Transfer, error) {
	const method = "UpdateTransaction"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	transfer, err := s.transferQueryRepository.FindById(ctx, *request.TransferID)
	if err != nil {
		status = "error"
		s.markTransferAsFailed(ctx, *request.TransferID, method, span)
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.Int("transfer_id", *request.TransferID))
	}

	amountDifference := request.TransferAmount - int(transfer.TransferAmount)

	senderSaldo, err := s.saldoAdapter.FindByCardNumber(ctx, transfer.TransferFrom)
	if err != nil {
		status = "error"
		s.markTransferAsFailed(ctx, *request.TransferID, method, span)
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(transfer.TransferFrom)))
	}

	receiverSaldo, err := s.saldoAdapter.FindByCardNumber(ctx, transfer.TransferTo)
	if err != nil {
		status = "error"
		s.markTransferAsFailed(ctx, *request.TransferID, method, span)
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("to_card", security.MaskCardNumber(transfer.TransferTo)))
	}

	if amountDifference > 0 {
		if _, err = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber: senderSaldo.CardNumber, Amount: amountDifference,
			OperationID: fmt.Sprintf("transfer:update:%d:%d:sender-debit", transfer.TransferID, request.TransferAmount),
			SourceType:  "transfer_update", SourceID: strconv.Itoa(int(transfer.TransferID)),
		}); err != nil {
			status = "error"
			s.markTransferAsFailed(ctx, *request.TransferID, method, span)
			return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(transfer.TransferFrom)))
		}
		if _, err = s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
			CardNumber: receiverSaldo.CardNumber, Amount: amountDifference,
			OperationID: fmt.Sprintf("transfer:update:%d:%d:receiver-credit", transfer.TransferID, request.TransferAmount),
			SourceType:  "transfer_update", SourceID: strconv.Itoa(int(transfer.TransferID)),
		}); err != nil {
			status = "error"
			_, _ = s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{CardNumber: senderSaldo.CardNumber, Amount: amountDifference, OperationID: fmt.Sprintf("transfer:update:%d:%d:rollback-sender-credit", transfer.TransferID, request.TransferAmount), SourceType: "transfer_update_compensation", SourceID: strconv.Itoa(int(transfer.TransferID))})
			s.markTransferAsFailed(ctx, *request.TransferID, method, span)
			return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("failed_to_credit", "receiver"))
		}
	} else if amountDifference < 0 {
		adjustment := -amountDifference
		if _, err = s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
			CardNumber: senderSaldo.CardNumber, Amount: adjustment,
			OperationID: fmt.Sprintf("transfer:update:%d:%d:sender-credit", transfer.TransferID, request.TransferAmount),
			SourceType:  "transfer_update", SourceID: strconv.Itoa(int(transfer.TransferID)),
		}); err != nil {
			status = "error"
			s.markTransferAsFailed(ctx, *request.TransferID, method, span)
			return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("from_card", security.MaskCardNumber(transfer.TransferFrom)))
		}
		if _, err = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber: receiverSaldo.CardNumber, Amount: adjustment,
			OperationID: fmt.Sprintf("transfer:update:%d:%d:receiver-debit", transfer.TransferID, request.TransferAmount),
			SourceType:  "transfer_update", SourceID: strconv.Itoa(int(transfer.TransferID)),
		}); err != nil {
			status = "error"
			_, _ = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{CardNumber: senderSaldo.CardNumber, Amount: adjustment, OperationID: fmt.Sprintf("transfer:update:%d:%d:rollback-sender-debit", transfer.TransferID, request.TransferAmount), SourceType: "transfer_update_compensation", SourceID: strconv.Itoa(int(transfer.TransferID))})
			s.markTransferAsFailed(ctx, *request.TransferID, method, span)
			return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.String("failed_to_debit", "receiver"))
		}
	}

	updatedTransfer, err := s.transferCommandRepository.UpdateTransfer(ctx, request)
	if err != nil {
		status = "error"
		s.markTransferAsFailed(ctx, *request.TransferID, method, span)
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span)
	}

	if _, err := s.transferCommandRepository.UpdateTransferStatus(ctx, &requests.UpdateTransferStatus{
		TransferID: *request.TransferID,
		Status:     "success",
	}); err != nil {
		status = "error"
		s.markTransferAsFailed(ctx, *request.TransferID, method, span)
		return errorhandler.HandleError[*salemongo.Transfer](s.logger, err, method, span, zap.Int("transfer_id", *request.TransferID))
	}

	logSuccess("Successfully updated transfer", zap.Int("transfer.id", *request.TransferID))

	return updatedTransfer, nil
}

func (s *transferCommandService) TrashedTransfer(ctx context.Context, transfer_id int) (*salemongo.Transfer, error) {
	const method = "TrashedTransfer"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("transfer_id", transfer_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting trashed transfer process", zap.Int("transfer_id", transfer_id))

	res, err := s.transferCommandRepository.TrashedTransfer(ctx, transfer_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](
			s.logger,
			err,
			method,
			span,

			zap.Int("transfer_id", transfer_id),
		)
	}

	logSuccess("Successfully trashed transfer", zap.Int("transfer_id", transfer_id))

	return res, nil
}

func (s *transferCommandService) RestoreTransfer(ctx context.Context, transfer_id int) (*salemongo.Transfer, error) {
	const method = "RestoreTransfer"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("transfer_id", transfer_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting restore transfer process", zap.Int("transfer_id", transfer_id))

	res, err := s.transferCommandRepository.RestoreTransfer(ctx, transfer_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transfer](
			s.logger,
			err,
			method,
			span,

			zap.Int("transfer_id", transfer_id),
		)
	}

	logSuccess("Successfully restored transfer", zap.Int("transfer_id", transfer_id))

	return res, nil
}

func (s *transferCommandService) DeleteTransferPermanent(ctx context.Context, transfer_id int) (bool, error) {
	const method = "DeleteTransferPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("transfer_id", transfer_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting delete transfer permanent process", zap.Int("transfer_id", transfer_id))

	_, err := s.transferCommandRepository.DeleteTransferPermanent(ctx, transfer_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			err,
			method,
			span,

			zap.Int("transfer_id", transfer_id),
		)
	}

	logSuccess("Successfully deleted transfer permanently", zap.Int("transfer_id", transfer_id))

	return true, nil
}

func (s *transferCommandService) RestoreAllTransfer(ctx context.Context) (bool, error) {
	const method = "RestoreAllTransfer"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Restoring all transfers")

	_, err := s.transferCommandRepository.RestoreAllTransfer(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("restore all transfers"),
			method,
			span,
		)
	}

	logSuccess("Successfully restored all transfers")
	return true, nil
}

func (s *transferCommandService) DeleteAllTransferPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllTransferPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Permanently deleting all transfers")

	_, err := s.transferCommandRepository.DeleteAllTransferPermanent(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("delete all transfers permanently"),
			method,
			span,
		)
	}

	logSuccess("Successfully deleted all transfers permanently")
	return true, nil
}

func (s *transferCommandService) compensateTransfer(ctx context.Context, transferID, amount int, senderCard, receiverCard, method string) {
	ok, err := s.transferCommandRepository.GuardStatus(ctx, transferID, state.Processing, state.Compensating, "partial settlement; compensation started")
	if err != nil || !ok {
		return
	}
	if _, err := s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
		CardNumber: receiverCard, Amount: amount,
		OperationID: fmt.Sprintf("transfer:%d:receiver-compensation", transferID),
		SourceType:  "transfer_compensation", SourceID: strconv.Itoa(transferID),
	}); err != nil {
		_, _ = s.transferCommandRepository.GuardStatus(ctx, transferID, state.Compensating, state.Unknown, "receiver compensation failed: "+err.Error())
		return
	}
	if _, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber: senderCard, Amount: amount,
		OperationID: fmt.Sprintf("transfer:%d:sender-compensation", transferID),
		SourceType:  "transfer_compensation", SourceID: strconv.Itoa(transferID),
	}); err != nil {
		_, _ = s.transferCommandRepository.GuardStatus(ctx, transferID, state.Compensating, state.Unknown, "sender compensation failed: "+err.Error())
		return
	}
	_, _ = s.transferCommandRepository.GuardStatus(ctx, transferID, state.Compensating, state.Compensated, "")
}

func (s *transferCommandService) enqueueTransferEvents(ctx context.Context, transfer *salemongo.Transfer, updatedTransfer *salemongo.Transfer, senderCard *salemongo.CardByEmailRow, request *requests.CreateTransferRequest, senderNewBalance, receiverNewBalance int) {
	if s.outboxStore == nil {
		return
	}
	tID := strconv.Itoa(int(transfer.TransferID))

	htmlBody, err := email.GenerateEmailHTML(map[string]string{
		"Title": "Transfer Successful", "Message": fmt.Sprintf("Your Transfer of %d has been processed successfully.", request.TransferAmount),
		"Button": "View History", "Link": "https://sanedge.example.com/withdraw/history",
	})
	if err == nil {
		emailPayload, _ := json.Marshal(map[string]any{"email": senderCard.Email, "subject": "Transfer Successful - SanEdge", "body": htmlBody})
		if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transfer", AggregateID: tID, EventType: "transfer.created", Payload: emailPayload}); iErr != nil {
			s.logger.Error("outbox: failed to enqueue transfer email event", zap.Error(iErr))
		}
	}

	statsEvent := events.TransferEvent{
		TransferID:      uint64(updatedTransfer.TransferID),
		TransferNo:      updatedTransfer.TransferNo.String(),
		SourceCard:      request.TransferFrom,
		DestinationCard: request.TransferTo,
		Amount:          int64(request.TransferAmount),
		Status:          "success",
		CreatedAt:       time.Now(),
	}
	statsBytes, _ := json.Marshal(statsEvent)
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transfer", AggregateID: tID, EventType: "transfer.stats", Payload: statsBytes}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue transfer stats event", zap.Error(iErr))
	}

	senderSaldoBytes, _ := json.Marshal(events.SaldoEvent{CardNumber: request.TransferFrom, TotalBalance: int64(senderNewBalance), CreatedAt: time.Now()})
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transfer", AggregateID: request.TransferFrom, EventType: "transfer.saldo_sender", Payload: senderSaldoBytes}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue sender saldo event", zap.Error(iErr))
	}

	receiverSaldoBytes, _ := json.Marshal(events.SaldoEvent{CardNumber: request.TransferTo, TotalBalance: int64(receiverNewBalance), CreatedAt: time.Now()})
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transfer", AggregateID: request.TransferTo, EventType: "transfer.saldo_receiver", Payload: receiverSaldoBytes}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue receiver saldo event", zap.Error(iErr))
	}
}

func (s *transferCommandService) markTransferAsFailed(ctx context.Context, transferID int, method string, span trace.Span) {
	req := &requests.UpdateTransferStatus{
		TransferID: transferID,
		Status:     "failed",
	}
	async.RunWithTimeout(5*time.Second, func(ctx context.Context) {
		if _, err := s.transferCommandRepository.UpdateTransferStatus(ctx, req); err != nil {
			s.logger.Error("compensation: failed to mark transfer as failed", zap.Error(err), zap.Int("transfer_id", transferID), zap.String("method", method))
		}
	})
}
