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
	mencache "github.com/MamangRust/monolith-payment-gateway-transaction/redis"
	"github.com/MamangRust/monolith-payment-gateway-transaction/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// transactionCommandServiceDeps groups dependencies for transaction command service.
type TransactionCommandServiceDeps struct {
	Kafka                        *kafka.Kafka
	Mencache                     mencache.TransactionCommandCache
	Tracer                       trace.Tracer
	MerchantAdapter              adapter.MerchantAdapter
	CardAdapter                  adapter.CardAdapter
	SaldoAdapter                 adapter.SaldoAdapter
	TransactionQueryRepository   repository.TransactionQueryRepository
	TransactionCommandRepository repository.TransactionCommandRepository
	IdempotencyStore             idempotency.Store
	OutboxStore                  repository.OutboxRepository
	Logger                       logger.LoggerInterface
	Observability                observability.TraceLoggerObservability
	AISecurityClient             ai_security.AISecurityServiceClient
}

// transactionCommandService handles transaction write operations.
type transactionCommandService struct {
	kafka                        *kafka.Kafka
	cache                        mencache.TransactionCommandCache
	merchantAdapter              adapter.MerchantAdapter
	cardAdapter                  adapter.CardAdapter
	saldoAdapter                 adapter.SaldoAdapter
	transactionQueryRepository   repository.TransactionQueryRepository
	transactionCommandRepository repository.TransactionCommandRepository
	idempotencyStore             idempotency.Store
	outboxStore                  repository.OutboxRepository
	logger                       logger.LoggerInterface
	observability                observability.TraceLoggerObservability
	aiSecurityClient             ai_security.AISecurityServiceClient
}

func NewTransactionCommandService(
	params *TransactionCommandServiceDeps,
) TransactionCommandService {
	return &transactionCommandService{
		kafka:                        params.Kafka,
		cache:                        params.Mencache,
		merchantAdapter:              params.MerchantAdapter,
		cardAdapter:                  params.CardAdapter,
		saldoAdapter:                 params.SaldoAdapter,
		transactionCommandRepository: params.TransactionCommandRepository,
		transactionQueryRepository:   params.TransactionQueryRepository,
		idempotencyStore:             params.IdempotencyStore,
		outboxStore:                  params.OutboxStore,
		logger:                       params.Logger,
		observability:                params.Observability,
		aiSecurityClient:             params.AISecurityClient,
	}
}

func (s *transactionCommandService) Create(ctx context.Context, apiKey string, request *requests.CreateTransactionRequest) (*salemongo.Transaction, error) {
	const method = "CreateTransaction"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("apikey", security.MaskAPIKey(apiKey)))
	defer func() { end(status, "grpc") }()

	// Idempotency guard
	var idemHash string
	if request.IdempotencyKey != "" && s.idempotencyStore != nil {
		idemHash = idempotency.HashRequest(request)
		if _, err := s.idempotencyStore.Claim(ctx, "transaction", request.IdempotencyKey, idemHash); err == nil {
			defer func() {
				if status == "error" {
					if relErr := s.idempotencyStore.Release(ctx, "transaction", request.IdempotencyKey, idemHash); relErr != nil {
						s.logger.Error("idempotency: failed to release key", zap.Error(relErr), zap.String("idempotency_key", request.IdempotencyKey))
					}
				}
			}()
		} else {
			existing, gErr := s.idempotencyStore.Get(ctx, "transaction", request.IdempotencyKey)
			if gErr != nil {
				status = "error"
				return errorhandler.HandleError[*salemongo.Transaction](s.logger, gErr, method, span, zap.String("idempotency_key", request.IdempotencyKey))
			}
			if existing.RequestHash != idemHash {
				status = "error"
				return errorhandler.HandleError[*salemongo.Transaction](s.logger, idempotency.ConflictError(), method, span, zap.String("idempotency_key", request.IdempotencyKey))
			}
			if existing.Status == idempotency.StatusSuccess {
				var txn salemongo.Transaction
				if uErr := json.Unmarshal(existing.ResponseJSON, &txn); uErr != nil {
					status = "error"
					return errorhandler.HandleError[*salemongo.Transaction](s.logger, uErr, method, span, zap.String("idempotency_key", request.IdempotencyKey))
				}
				logSuccess("Replayed idempotent transaction response", zap.String("idempotency_key", request.IdempotencyKey))
				return &txn, nil
			}
			status = "error"
			return errorhandler.HandleError[*salemongo.Transaction](s.logger, idempotency.ProcessingError(), method, span, zap.String("idempotency_key", request.IdempotencyKey))
		}
	}

	if err := validation.ValidateAmount(request.Amount); err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	merchant, err := s.merchantAdapter.FindByApiKey(ctx, apiKey)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("api_key", security.MaskAPIKey(apiKey)))
	}

	card, err := s.cardAdapter.FindUserCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	saldo, err := s.saldoAdapter.FindByCardNumber(ctx, card.CardNumber)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(card.CardNumber)))
	}

	if int(saldo.TotalBalance) < request.Amount {
		status = "error"
		err := sharedErrors.ErrConflict.WithMessage("insufficient balance")
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Float64("current_balance", float64(saldo.TotalBalance)), zap.Float64("requested_amount", float64(request.Amount)))
	}

	// AI Security Check
	if s.aiSecurityClient != nil {
		securityRes, err := s.aiSecurityClient.DetectFraud(ctx, &ai_security.FraudRequest{
			TransactionId: strconv.Itoa(int(time.Now().UnixNano())),
			MerchantId:    int32(merchant.MerchantID),
			UserId:        int32(card.UserID),
			Amount:        float64(request.Amount),
			PaymentMethod: request.PaymentMethod,
		})
		if err == nil && securityRes.IsFraudulent {
			status = "error"
			s.logger.Warn("Transaction blocked by AI Security", zap.String("reason", securityRes.Reason))
			return nil, errors.New("security block: " + securityRes.Reason)
		}
	}

	merchantId := int(merchant.MerchantID)
	request.MerchantID = &merchantId
	transaction, err := s.transactionCommandRepository.CreateTransaction(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(card.CardNumber)))
	}
	txID := int(transaction.TransactionID)

	if ok, gErr := s.transactionCommandRepository.GuardStatus(ctx, txID, state.Pending, state.Processing, ""); gErr != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, gErr, method, span, zap.Int("transaction_id", txID))
	} else if !ok {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, errors.New("transaction already being processed"), method, span, zap.Int("transaction_id", txID))
	}

	debitedSaldo, err := s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
		CardNumber:  card.CardNumber,
		Amount:      request.Amount,
		OperationID: fmt.Sprintf("transaction:%d:user-debit", txID),
		SourceType:  "transaction",
		SourceID:    strconv.Itoa(txID),
	})
	if err != nil {
		_, _ = s.transactionCommandRepository.GuardStatus(ctx, txID, state.Processing, state.Failed, "debit failed: "+err.Error())
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(card.CardNumber)))
	}
	newUserBalance := int(debitedSaldo.TotalBalance)

	merchantCard, err := s.cardAdapter.FindCardByUserId(ctx, int(merchant.UserID))
	if err != nil {
		s.compensateTransaction(ctx, txID, request.Amount, card.CardNumber, merchantId, false, method, span)
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("merchant_id", merchantId))
	}

	creditedMerchantSaldo, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber:  merchantCard.CardNumber,
		Amount:      request.Amount,
		OperationID: fmt.Sprintf("transaction:%d:merchant-credit", txID),
		SourceType:  "transaction",
		SourceID:    strconv.Itoa(txID),
	})
	if err != nil {
		_, _ = s.transactionCommandRepository.GuardStatus(ctx, txID, state.Processing, state.Unknown, "merchant credit outcome ambiguous: "+err.Error())
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("merchant_card_number", security.MaskCardNumber(merchantCard.CardNumber)))
	}
	newMerchantBalance := int(creditedMerchantSaldo.TotalBalance)

	updatedTransaction, err := s.transactionCommandRepository.TransitionStatus(ctx, txID, state.Processing, state.Success, "")
	if err != nil {
		s.compensateTransaction(ctx, txID, request.Amount, card.CardNumber, merchantId, true, method, span)
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", txID))
	}

	s.enqueueTransactionEvents(ctx, transaction, updatedTransaction, card, merchant, merchantCard, request, newUserBalance, newMerchantBalance)

	logSuccess("Successfully created transaction", zap.Int("transaction.id", int(updatedTransaction.TransactionID)))

	if request.IdempotencyKey != "" && s.idempotencyStore != nil {
		if respBytes, mErr := json.Marshal(updatedTransaction); mErr == nil {
			resourceID := updatedTransaction.TransactionID
			if cErr := s.idempotencyStore.Complete(ctx, "transaction", request.IdempotencyKey, idemHash, idempotency.Outcome{
				Status:       idempotency.StatusSuccess,
				ResponseJSON: respBytes,
				ResourceID:   &resourceID,
			}); cErr != nil {
				s.logger.Error("idempotency: failed to complete key", zap.Error(cErr), zap.String("idempotency_key", request.IdempotencyKey))
			}
		}
	}

	return updatedTransaction, nil
}

func (s *transactionCommandService) Update(ctx context.Context, apiKey string, request *requests.UpdateTransactionRequest) (*salemongo.Transaction, error) {
	const method = "UpdateTransaction"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	transaction, err := s.transactionQueryRepository.FindById(ctx, *request.TransactionID)
	if err != nil {
		status = "error"
		s.markTransactionAsFailed(ctx, *request.TransactionID, method, span)
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", *request.TransactionID))
	}

	merchant, err := s.merchantAdapter.FindByApiKey(ctx, apiKey)
	if err != nil || transaction.MerchantID != merchant.MerchantID {
		status = "error"
		s.markTransactionAsFailed(ctx, *request.TransactionID, method, span)
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("api_key", security.MaskAPIKey(apiKey)))
	}

	card, err := s.cardAdapter.FindCardByCardNumber(ctx, transaction.CardNumber)
	if err != nil {
		status = "error"
		s.markTransactionAsFailed(ctx, *request.TransactionID, method, span)
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(transaction.CardNumber)))
	}

	if _, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber:  card.CardNumber,
		Amount:      int(transaction.Amount),
		OperationID: fmt.Sprintf("transaction:update:%d:%d:reverse-old", transaction.TransactionID, request.Amount),
		SourceType:  "transaction_update",
		SourceID:    strconv.Itoa(int(transaction.TransactionID)),
	}); err != nil {
		status = "error"
		s.markTransactionAsFailed(ctx, *request.TransactionID, method, span)
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", card.CardNumber))
	}
	_, err = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
		CardNumber:  card.CardNumber,
		Amount:      request.Amount,
		OperationID: fmt.Sprintf("transaction:update:%d:%d:apply-new", transaction.TransactionID, request.Amount),
		SourceType:  "transaction_update",
		SourceID:    strconv.Itoa(int(transaction.TransactionID)),
	})
	if err != nil {
		status = "error"
		_, _ = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber:  card.CardNumber,
			Amount:      int(transaction.Amount),
			OperationID: fmt.Sprintf("transaction:update:%d:%d:restore-old-debit", transaction.TransactionID, request.Amount),
			SourceType:  "transaction_update_compensation",
			SourceID:    strconv.Itoa(int(transaction.TransactionID)),
		})
		s.markTransactionAsFailed(ctx, *request.TransactionID, method, span)
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", card.CardNumber))
	}
	parsedTime := transaction.TransactionTime

	merchantId := int(transaction.MerchantID)
	transactionId := int(transaction.TransactionID)

	updateReq := &requests.UpdateTransactionRequest{
		TransactionID:   &transactionId,
		CardNumber:      transaction.CardNumber,
		Amount:          request.Amount,
		PaymentMethod:   request.PaymentMethod,
		MerchantID:      &merchantId,
		TransactionTime: parsedTime,
	}
	res, err := s.transactionCommandRepository.UpdateTransaction(ctx, updateReq)
	if err != nil {
		status = "error"
		_, _ = s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
			CardNumber:  card.CardNumber,
			Amount:      request.Amount,
			OperationID: fmt.Sprintf("transaction:update:%d:%d:undo-new-debit", transaction.TransactionID, request.Amount),
			SourceType:  "transaction_update_compensation",
			SourceID:    strconv.Itoa(int(transaction.TransactionID)),
		})
		_, _ = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber:  card.CardNumber,
			Amount:      int(transaction.Amount),
			OperationID: fmt.Sprintf("transaction:update:%d:%d:restore-old-debit", transaction.TransactionID, request.Amount),
			SourceType:  "transaction_update_compensation",
			SourceID:    strconv.Itoa(int(transaction.TransactionID)),
		})
		s.markTransactionAsFailed(ctx, *request.TransactionID, method, span)
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", *request.TransactionID))
	}

	if _, err := s.transactionCommandRepository.UpdateTransactionStatus(ctx, &requests.UpdateTransactionStatus{
		TransactionID: int(transaction.TransactionID),
		Status:        "success",
	}); err != nil {
		status = "error"
		_, _ = s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
			CardNumber:  card.CardNumber,
			Amount:      request.Amount,
			OperationID: fmt.Sprintf("transaction:update:%d:%d:undo-new-debit", transaction.TransactionID, request.Amount),
			SourceType:  "transaction_update_compensation",
			SourceID:    strconv.Itoa(int(transaction.TransactionID)),
		})
		_, _ = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber:  card.CardNumber,
			Amount:      int(transaction.Amount),
			OperationID: fmt.Sprintf("transaction:update:%d:%d:restore-old-debit", transaction.TransactionID, request.Amount),
			SourceType:  "transaction_update_compensation",
			SourceID:    strconv.Itoa(int(transaction.TransactionID)),
		})
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", int(transaction.TransactionID)))
	}

	logSuccess("Successfully updated transaction", zap.Int("transaction.id", int(res.TransactionID)))

	return res, nil
}

func (s *transactionCommandService) TrashedTransaction(ctx context.Context, transaction_id int) (*salemongo.Transaction, error) {
	const method = "TrashedTransaction"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("transaction_id", transaction_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting TrashedTransaction process", zap.Int("transaction_id", transaction_id))

	res, err := s.transactionCommandRepository.TrashedTransaction(ctx, transaction_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](
			s.logger,
			err,
			method,
			span,

			zap.Int("transaction_id", transaction_id),
		)
	}

	logSuccess("Successfully trashed transaction", zap.Int("transaction_id", transaction_id))

	return res, nil
}

func (s *transactionCommandService) RestoreTransaction(ctx context.Context, transaction_id int) (*salemongo.Transaction, error) {
	const method = "RestoreTransaction"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("transaction_id", transaction_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting RestoreTransaction process", zap.Int("transaction_id", transaction_id))

	res, err := s.transactionCommandRepository.RestoreTransaction(ctx, transaction_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Transaction](
			s.logger,
			err,
			method,
			span,

			zap.Int("transaction_id", transaction_id),
		)
	}

	logSuccess("Successfully restored transaction", zap.Int("transaction_id", transaction_id))

	return res, nil
}

func (s *transactionCommandService) DeleteTransactionPermanent(ctx context.Context, transaction_id int) (bool, error) {
	const method = "DeleteTransactionPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("transaction_id", transaction_id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting DeleteTransactionPermanent process", zap.Int("transaction_id", transaction_id))

	_, err := s.transactionCommandRepository.DeleteTransactionPermanent(ctx, transaction_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			err,
			method,
			span,

			zap.Int("transaction_id", transaction_id),
		)
	}

	logSuccess("Successfully permanently deleted transaction", zap.Int("transaction_id", transaction_id))

	return true, nil
}

func (s *transactionCommandService) RestoreAllTransaction(ctx context.Context) (bool, error) {
	const method = "RestoreAllTransaction"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Restoring all transactions")

	_, err := s.transactionCommandRepository.RestoreAllTransaction(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("restore all transactions"),
			method,
			span,
		)
	}

	logSuccess("Successfully restored all transactions")
	return true, nil
}

func (s *transactionCommandService) DeleteAllTransactionPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllTransactionPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Permanently deleting all transactions")

	_, err := s.transactionCommandRepository.DeleteAllTransactionPermanent(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("delete all transactions permanently"),
			method,
			span,
		)
	}

	logSuccess("Successfully deleted all transactions permanently")
	return true, nil
}

// compensateTransaction reverses a partial transaction exactly once.
func (s *transactionCommandService) compensateTransaction(ctx context.Context, transactionID, amount int, userCard string, merchantID int, merchantCredited bool, method string, span trace.Span) {
	ok, err := s.transactionCommandRepository.GuardStatus(ctx, transactionID, state.Processing, state.Compensating, "partial settlement; compensation started")
	if err != nil || !ok {
		if err != nil {
			s.logger.Error("compensation: failed to claim transaction", zap.Error(err), zap.Int("transaction_id", transactionID))
		}
		return
	}

	if _, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber: userCard, Amount: amount,
		OperationID: fmt.Sprintf("transaction:%d:user-compensation", transactionID),
		SourceType:  "transaction_compensation", SourceID: strconv.Itoa(transactionID),
	}); err != nil {
		_, _ = s.transactionCommandRepository.GuardStatus(ctx, transactionID, state.Compensating, state.Unknown, "user debit compensation failed: "+err.Error())
		s.logger.Error("compensation: failed to restore user saldo", zap.Error(err), zap.Int("transaction_id", transactionID))
		return
	}

	if !merchantCredited {
		if _, err := s.transactionCommandRepository.GuardStatus(ctx, transactionID, state.Compensating, state.Compensated, ""); err != nil {
			s.logger.Error("compensation: failed to finalize transaction", zap.Error(err), zap.Int("transaction_id", transactionID), zap.String("method", method))
		}
		return
	}

	merchant, err := s.merchantAdapter.FindByMerchantId(ctx, merchantID)
	if err != nil {
		_, _ = s.transactionCommandRepository.GuardStatus(ctx, transactionID, state.Compensating, state.Unknown, "merchant lookup for compensation failed: "+err.Error())
		s.logger.Error("compensation: failed to resolve merchant", zap.Error(err), zap.Int("transaction_id", transactionID))
		return
	}
	merchantCard, err := s.cardAdapter.FindCardByUserId(ctx, int(merchant.UserID))
	if err != nil {
		_, _ = s.transactionCommandRepository.GuardStatus(ctx, transactionID, state.Compensating, state.Unknown, "merchant card lookup for compensation failed: "+err.Error())
		s.logger.Error("compensation: failed to resolve merchant card", zap.Error(err), zap.Int("transaction_id", transactionID))
		return
	}
	if _, err := s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
		CardNumber: merchantCard.CardNumber, Amount: amount,
		OperationID: fmt.Sprintf("transaction:%d:merchant-compensation", transactionID),
		SourceType:  "transaction_compensation", SourceID: strconv.Itoa(transactionID),
	}); err != nil {
		_, _ = s.transactionCommandRepository.GuardStatus(ctx, transactionID, state.Compensating, state.Unknown, "merchant credit compensation failed: "+err.Error())
		s.logger.Error("compensation: failed to reverse merchant saldo", zap.Error(err), zap.Int("transaction_id", transactionID))
		return
	}
	if _, err := s.transactionCommandRepository.GuardStatus(ctx, transactionID, state.Compensating, state.Compensated, ""); err != nil {
		s.logger.Error("compensation: failed to finalize transaction", zap.Error(err), zap.Int("transaction_id", transactionID), zap.String("method", method))
	}
}

func (s *transactionCommandService) enqueueTransactionEvents(ctx context.Context, tx *salemongo.Transaction, updatedTx *salemongo.Transaction, card *salemongo.CardByEmailRow, merchant *salemongo.MerchantAllFieldsRow, merchantCard *salemongo.CardAllFieldsRow, request *requests.CreateTransactionRequest, newUserBalance, newMerchantBalance int) {
	if s.outboxStore == nil {
		return
	}
	txID := strconv.Itoa(int(tx.TransactionID))

	htmlBody, err := email.GenerateEmailHTML(map[string]string{
		"Title": "Transaction Successful", "Message": fmt.Sprintf("Your transaction of %d has been processed successfully.", request.Amount),
		"Button": "View History", "Link": "https://sanedge.example.com/transaction/history",
	})
	if err == nil {
		emailPayload, _ := json.Marshal(map[string]any{"email": card.Email, "subject": "Transaction Successful - SanEdge", "body": htmlBody})
		if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transaction", AggregateID: txID, EventType: "transaction.created", Payload: emailPayload}); iErr != nil {
			s.logger.Error("outbox: failed to enqueue transaction email event", zap.Error(iErr))
		}
	}

	statsEvent := events.TransactionEvent{
		TransactionID: uint64(updatedTx.TransactionID),
		TransactionNo: updatedTx.TransactionNo.String(),
		CardNumber:    card.CardNumber, CardType: card.CardType, CardProvider: card.CardProvider,
		Amount: int64(request.Amount), PaymentMethod: request.PaymentMethod,
		MerchantID: uint64(merchant.MerchantID), MerchantName: merchant.Name,
		Status: "success", CreatedAt: time.Now(),
	}
	statsBytes, _ := json.Marshal(statsEvent)
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transaction", AggregateID: txID, EventType: "transaction.stats", Payload: statsBytes}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue transaction stats event", zap.Error(iErr))
	}

	userSaldoBytes, _ := json.Marshal(events.SaldoEvent{CardNumber: card.CardNumber, TotalBalance: int64(newUserBalance), CreatedAt: time.Now()})
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transaction", AggregateID: card.CardNumber, EventType: "transaction.saldo_user", Payload: userSaldoBytes}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue user saldo event", zap.Error(iErr))
	}

	merchantSaldoBytes, _ := json.Marshal(events.SaldoEvent{CardNumber: merchantCard.CardNumber, TotalBalance: int64(newMerchantBalance), CreatedAt: time.Now()})
	if iErr := s.outboxStore.Insert(ctx, repository.OutboxRecord{AggregateType: "transaction", AggregateID: merchantCard.CardNumber, EventType: "transaction.saldo_merchant", Payload: merchantSaldoBytes}); iErr != nil {
		s.logger.Error("outbox: failed to enqueue merchant saldo event", zap.Error(iErr))
	}
}

func (s *transactionCommandService) markTransactionAsFailed(ctx context.Context, transactionID int, method string, span trace.Span) {
	req := &requests.UpdateTransactionStatus{
		TransactionID: transactionID,
		Status:        "failed",
	}
	async.RunWithTimeout(5*time.Second, func(ctx context.Context) {
		if _, err := s.transactionCommandRepository.UpdateTransactionStatus(ctx, req); err != nil {
			s.logger.Error("compensation: failed to mark transaction as failed", zap.Error(err), zap.Int("transaction_id", transactionID), zap.String("method", method))
		}
	})
}
