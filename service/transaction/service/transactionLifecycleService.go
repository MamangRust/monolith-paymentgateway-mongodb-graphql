package service

import (
	"context"
	"fmt"
	"strconv"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/status"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/validation"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/security"
	"github.com/MamangRust/monolith-payment-gateway-shared/state"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// AuthorizeTransaction places a hold on the customer's funds without settling
// them to the merchant. Settlement happens later through CaptureTransaction.
func (s *transactionCommandService) AuthorizeTransaction(ctx context.Context, apiKey string, request *requests.CreateTransactionRequest) (*salemongo.Transaction, error) {
	const method = "AuthorizeTransaction"

	ctx, span, end, statusFlag, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("apikey", security.MaskAPIKey(apiKey)))
	defer func() { end(statusFlag, "grpc") }()

	if request == nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, sharedErrors.NewBadRequestError("transaction authorization request cannot be nil"), method, span)
	}

	if err := validation.ValidateAmount(request.Amount); err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	merchant, err := s.merchantAdapter.FindByApiKey(ctx, apiKey)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("api_key", security.MaskAPIKey(apiKey)))
	}

	card, err := s.cardAdapter.FindUserCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	if _, err := s.saldoAdapter.FindByCardNumber(ctx, card.CardNumber); err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(card.CardNumber)))
	}

	merchantID := int(merchant.MerchantID)
	request.MerchantID = &merchantID

	transaction, err := s.transactionCommandRepository.CreateTransaction(ctx, request)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	}

	authorized, err := s.transactionCommandRepository.TransitionStatus(ctx, int(transaction.TransactionID), state.Pending, status.Authorized, "")
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", int(transaction.TransactionID)))
	}

	logSuccess("Successfully authorized transaction", zap.Int("transaction.id", int(authorized.TransactionID)))

	return authorized, nil
}

// CaptureTransaction settles a previously authorized transaction to the
// merchant by crediting the merchant card saldo.
func (s *transactionCommandService) CaptureTransaction(ctx context.Context, apiKey string, transactionID int) (*salemongo.Transaction, error) {
	const method = "CaptureTransaction"

	ctx, span, end, statusFlag, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("transaction_id", transactionID))
	defer func() { end(statusFlag, "grpc") }()

	transaction, err := s.transactionQueryRepository.FindById(ctx, transactionID)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	if err := status.ValidateTransition(status.DomainTransaction, transaction.Status, status.Captured); err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("current_status", transaction.Status))
	}

	merchant, err := s.merchantAdapter.FindByApiKey(ctx, apiKey)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}
	if transaction.MerchantID != merchant.MerchantID {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, sharedErrors.NewForbiddenError("merchant does not own transaction"), method, span, zap.Int("transaction_id", transactionID))
	}

	merchantCard, err := s.cardAdapter.FindCardByUserId(ctx, int(merchant.UserID))
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("merchant_id", int(merchant.MerchantID)))
	}

	if _, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber:  merchantCard.CardNumber,
		Amount:      int(transaction.Amount),
		OperationID: fmt.Sprintf("transaction:%d:capture-credit", transactionID),
		SourceType:  "transaction_capture",
		SourceID:    strconv.Itoa(transactionID),
	}); err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	captured, err := s.transactionCommandRepository.TransitionStatus(ctx, transactionID, status.Authorized, status.Captured, "")
	if err != nil {
		_, _ = s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
			CardNumber:  merchantCard.CardNumber,
			Amount:      int(transaction.Amount),
			OperationID: fmt.Sprintf("transaction:%d:capture-compensation", transactionID),
			SourceType:  "transaction_compensation",
			SourceID:    strconv.Itoa(transactionID),
		})
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	logSuccess("Successfully captured transaction", zap.Int("transaction.id", transactionID))

	return captured, nil
}

// VoidTransaction releases a hold created by AuthorizeTransaction without any
// settlement, so no balances move.
func (s *transactionCommandService) VoidTransaction(ctx context.Context, apiKey string, transactionID int) (*salemongo.Transaction, error) {
	const method = "VoidTransaction"

	ctx, span, end, statusFlag, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("transaction_id", transactionID))
	defer func() { end(statusFlag, "grpc") }()

	transaction, err := s.transactionQueryRepository.FindById(ctx, transactionID)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	if err := status.ValidateTransition(status.DomainTransaction, transaction.Status, status.Voided); err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("current_status", transaction.Status))
	}

	merchant, err := s.merchantAdapter.FindByApiKey(ctx, apiKey)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}
	if transaction.MerchantID != merchant.MerchantID {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, sharedErrors.NewForbiddenError("merchant does not own transaction"), method, span, zap.Int("transaction_id", transactionID))
	}

	voided, err := s.transactionCommandRepository.TransitionStatus(ctx, transactionID, transaction.Status, status.Voided, "")
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	logSuccess("Successfully voided transaction", zap.Int("transaction.id", transactionID))

	return voided, nil
}

// RefundTransaction returns a captured amount to the customer and reverses the
// merchant settlement.
func (s *transactionCommandService) RefundTransaction(ctx context.Context, apiKey string, transactionID int) (*salemongo.Transaction, error) {
	const method = "RefundTransaction"

	ctx, span, end, statusFlag, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("transaction_id", transactionID))
	defer func() { end(statusFlag, "grpc") }()

	transaction, err := s.transactionQueryRepository.FindById(ctx, transactionID)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	if err := status.ValidateTransition(status.DomainTransaction, transaction.Status, status.Refunded); err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("current_status", transaction.Status))
	}

	merchant, err := s.merchantAdapter.FindByApiKey(ctx, apiKey)
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}
	if transaction.MerchantID != merchant.MerchantID {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, sharedErrors.NewForbiddenError("merchant does not own transaction"), method, span, zap.Int("transaction_id", transactionID))
	}

	merchantCard, err := s.cardAdapter.FindCardByUserId(ctx, int(merchant.UserID))
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("merchant_id", int(merchant.MerchantID)))
	}

	if _, err := s.saldoAdapter.DebitSaldo(ctx, &requests.DebitSaldoRequest{
		CardNumber:  merchantCard.CardNumber,
		Amount:      int(transaction.Amount),
		OperationID: fmt.Sprintf("transaction:%d:refund-merchant-debit", transactionID),
		SourceType:  "transaction_refund",
		SourceID:    strconv.Itoa(transactionID),
	}); err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.String("merchant_card_number", security.MaskCardNumber(merchantCard.CardNumber)))
	}

	if _, err := s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
		CardNumber:  transaction.CardNumber,
		Amount:      int(transaction.Amount),
		OperationID: fmt.Sprintf("transaction:%d:refund-user-credit", transactionID),
		SourceType:  "transaction_refund",
		SourceID:    strconv.Itoa(transactionID),
	}); err != nil {
		// Return the debited merchant amount so the ledger stays balanced.
		_, _ = s.saldoAdapter.CreditSaldo(ctx, &requests.CreditSaldoRequest{
			CardNumber:  merchantCard.CardNumber,
			Amount:      int(transaction.Amount),
			OperationID: fmt.Sprintf("transaction:%d:refund-compensation", transactionID),
			SourceType:  "transaction_compensation",
			SourceID:    strconv.Itoa(transactionID),
		})
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	refunded, err := s.transactionCommandRepository.TransitionStatus(ctx, transactionID, status.Captured, status.Refunded, "")
	if err != nil {
		statusFlag = "error"
		return errorhandler.HandleError[*salemongo.Transaction](s.logger, err, method, span, zap.Int("transaction_id", transactionID))
	}

	logSuccess("Successfully refunded transaction", zap.Int("transaction.id", transactionID))

	return refunded, nil
}
