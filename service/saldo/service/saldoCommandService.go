package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/email"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	mencache "github.com/MamangRust/monolith-payment-gateway-saldo/redis"
	"github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"github.com/MamangRust/monolith-payment-gateway-shared/security"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type saldoCommandParams struct {
	Cache                  mencache.SaldoCommandCache
	saldoCommandRepository repository.SaldoCommandRepository
	CardAdapter            adapter.CardAdapter
	Logger                 logger.LoggerInterface
	Observability          observability.TraceLoggerObservability
	Kafka                  *kafka.Kafka
}

type saldoCommandService struct {
	cache                  mencache.SaldoCommandCache
	saldoCommandRepository repository.SaldoCommandRepository
	cardAdapter            adapter.CardAdapter
	logger                 logger.LoggerInterface
	observability          observability.TraceLoggerObservability
	kafka                  *kafka.Kafka
}

func NewSaldoCommandService(params *saldoCommandParams) SaldoCommandService {
	return &saldoCommandService{
		cache:                  params.Cache,
		saldoCommandRepository: params.saldoCommandRepository,
		cardAdapter:            params.CardAdapter,
		logger:                 params.Logger,
		observability:          params.Observability,
		kafka:                  params.Kafka,
	}
}

func (s *saldoCommandService) CreateSaldo(ctx context.Context, request *requests.CreateSaldoRequest) (*repository.SaldoMutationResult, error) {
	const method = "CreateSaldo"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.cardAdapter.FindCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span, zap.String("card_number", request.CardNumber))
	}

	res, err := s.saldoCommandRepository.CreateSaldo(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span, zap.String("card_number", request.CardNumber))
	}

	s.publishSaldoCreatedEmail(ctx, request.CardNumber)

	logSuccess("Successfully created saldo record", zap.String("card_number", request.CardNumber))
	return res, nil
}

func (s *saldoCommandService) publishSaldoCreatedEmail(ctx context.Context, cardNumber string) {
	if s.kafka == nil {
		return
	}
	card, err := s.cardAdapter.FindUserCardByCardNumber(ctx, cardNumber)
	if err != nil {
		s.logger.Warn("saldo email: failed to resolve user email for card", zap.Error(err))
		return
	}
	if card == nil || card.Email == "" {
		return
	}
	htmlBody, err := email.GenerateEmailHTML(map[string]string{
		"Title":   "Saldo Account Created",
		"Message": fmt.Sprintf("Your payment account for card %s has been created successfully.", cardNumber),
		"Button":  "View Balance",
		"Link":    "https://sanedge.example.com/balance",
	})
	if err != nil {
		return
	}
	emailPayload, err := json.Marshal(map[string]any{"email": card.Email, "subject": "Saldo Account Created - SanEdge", "body": htmlBody})
	if err != nil {
		return
	}
	_ = s.kafka.SendMessage("email-service-topic-saldo-create", cardNumber, emailPayload)
}

func (s *saldoCommandService) CreateSaldoIfNotExists(ctx context.Context, request *requests.CreateSaldoRequest) error {
	return s.saldoCommandRepository.CreateSaldoIfNotExists(ctx, request)
}

func (s *saldoCommandService) enqueueSaldoChanged(ctx context.Context, saldoID int64, cardNumber string, totalBalance int64) {
	s.cache.DeleteSaldoCache(ctx, int(saldoID))
	s.cache.DeleteSaldoCacheByCardNumber(ctx, cardNumber)
}

func (s *saldoCommandService) UpdateSaldo(ctx context.Context, request *requests.UpdateSaldoRequest) (*repository.SaldoMutationResult, error) {
	const method = "UpdateSaldo"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.cardAdapter.FindCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}

	res, err := s.saldoCommandRepository.UpdateSaldo(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}

	s.enqueueSaldoChanged(ctx, int64(res.SaldoID), res.CardNumber, res.TotalBalance)
	logSuccess("Successfully updated saldo record", zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	return res, nil
}

func (s *saldoCommandService) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*repository.SaldoMutationResult, error) {
	const method = "UpdateSaldoBalance"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	if request == nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, sharedErrors.NewBadRequestError("saldo request cannot be nil"), method, span)
	}

	_, err := s.cardAdapter.FindCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}

	res, err := s.saldoCommandRepository.UpdateSaldoBalance(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}

	s.enqueueSaldoChanged(ctx, int64(res.SaldoID), res.CardNumber, res.TotalBalance)
	logSuccess("Successfully updated saldo balance", zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	return res, nil
}

func (s *saldoCommandService) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*repository.SaldoMutationResult, error) {
	const method = "DebitSaldo"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	if err := request.Validate(); err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, sharedErrors.NewBadRequestError("invalid debit saldo request"), method, span)
	}

	res, err := s.saldoCommandRepository.DebitSaldo(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}
	s.enqueueSaldoChanged(ctx, int64(res.SaldoID), res.CardNumber, res.TotalBalance)
	logSuccess("Successfully debited saldo", zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	return res, nil
}

func (s *saldoCommandService) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*repository.SaldoMutationResult, error) {
	const method = "CreditSaldo"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	if err := request.Validate(); err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, sharedErrors.NewBadRequestError("invalid credit saldo request"), method, span)
	}

	res, err := s.saldoCommandRepository.CreditSaldo(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}
	s.enqueueSaldoChanged(ctx, int64(res.SaldoID), res.CardNumber, res.TotalBalance)
	logSuccess("Successfully credited saldo", zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	return res, nil
}

func (s *saldoCommandService) ApplySaldoAdjustment(ctx context.Context, request *requests.ApplySaldoAdjustmentRequest) (*repository.SaldoAdjustmentResult, error) {
	const method = "ApplySaldoAdjustment"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	if err := request.Validate(); err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoAdjustmentResult](s.logger, sharedErrors.NewBadRequestError("invalid saldo adjustment request"), method, span)
	}
	res, err := s.saldoCommandRepository.ApplySaldoAdjustment(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoAdjustmentResult](s.logger, err, method, span)
	}
	s.enqueueSaldoChanged(ctx, int64(res.SaldoID), res.CardNumber, res.TotalBalance)
	logSuccess("Successfully applied saldo adjustment", zap.String("card_number", security.MaskCardNumber(request.CardNumber)))
	return res, nil
}

func (s *saldoCommandService) ResolveReconciliation(ctx context.Context, queueID int64, operationID, note string) error {
	return s.saldoCommandRepository.ResolveReconciliation(ctx, queueID, operationID, note)
}

func (s *saldoCommandService) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*repository.SaldoMutationResult, error) {
	const method = "UpdateSaldoWithdraw"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.cardAdapter.FindCardByCardNumber(ctx, request.CardNumber)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}

	res, err := s.saldoCommandRepository.UpdateSaldoWithdraw(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*repository.SaldoMutationResult](s.logger, err, method, span)
	}

	s.enqueueSaldoChanged(ctx, int64(res.SaldoID), res.CardNumber, res.TotalBalance)
	logSuccess("Successfully updated saldo withdraw record")
	return res, nil
}

func (s *saldoCommandService) TrashSaldo(ctx context.Context, saldo_id int) (*salemongo.Saldo, error) {
	const method = "TrashSaldo"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("saldo_id", saldo_id))
	defer func() { end(status, "grpc") }()

	res, err := s.saldoCommandRepository.TrashedSaldo(ctx, saldo_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Saldo](s.logger, err, method, span)
	}
	logSuccess("Successfully trashed saldo", zap.Int("saldo_id", saldo_id))
	return res, nil
}

func (s *saldoCommandService) RestoreSaldo(ctx context.Context, saldo_id int) (*salemongo.Saldo, error) {
	const method = "RestoreSaldo"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("saldo_id", saldo_id))
	defer func() { end(status, "grpc") }()

	res, err := s.saldoCommandRepository.RestoreSaldo(ctx, saldo_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Saldo](s.logger, err, method, span)
	}
	logSuccess("Successfully restored saldo", zap.Int("saldo_id", saldo_id))
	return res, nil
}

func (s *saldoCommandService) DeleteSaldoPermanent(ctx context.Context, saldo_id int) (bool, error) {
	const method = "DeleteSaldoPermanent"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("saldo_id", saldo_id))
	defer func() { end(status, "grpc") }()

	_, err := s.saldoCommandRepository.DeleteSaldoPermanent(ctx, saldo_id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}
	logSuccess("Successfully deleted saldo permanently")
	return true, nil
}

func (s *saldoCommandService) RestoreAllSaldo(ctx context.Context) (bool, error) {
	const method = "RestoreAllSaldo"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.saldoCommandRepository.RestoreAllSaldo(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, sharedErrors.ErrFailed("restore all saldo"), method, span)
	}
	logSuccess("Successfully restored all saldo")
	return true, nil
}

func (s *saldoCommandService) DeleteAllSaldoPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllSaldoPermanent"
	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.saldoCommandRepository.DeleteAllSaldoPermanent(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, sharedErrors.ErrFailed("delete all saldo permanently"), method, span)
	}
	logSuccess("Successfully deleted all saldo permanently")
	return true, nil
}
