package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"

	mencache "github.com/MamangRust/monolith-payment-gateway-merchant/redis"
	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/email"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/events"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type merchantCommandServiceDeps struct {
	Kafka                     *kafka.Kafka
	Cache                     mencache.MerchantCommandCache
	UserAdapter               adapter.UserAdapter
	MerchantQueryRepository   repository.MerchantQueryRepository
	MerchantCommandRepository repository.MerchantCommandRepository
	Logger                    logger.LoggerInterface
	Observability             observability.TraceLoggerObservability
}

type merchantCommandService struct {
	kafka                     *kafka.Kafka
	cache                     mencache.MerchantCommandCache
	userAdapter               adapter.UserAdapter
	merchantQueryRepository   repository.MerchantQueryRepository
	merchantCommandRepository repository.MerchantCommandRepository
	logger                    logger.LoggerInterface
	observability             observability.TraceLoggerObservability
}

func NewMerchantCommandService(params *merchantCommandServiceDeps) MerchantCommandService {
	return &merchantCommandService{
		kafka:                     params.Kafka,
		cache:                     params.Cache,
		merchantCommandRepository: params.MerchantCommandRepository,
		userAdapter:               params.UserAdapter,
		merchantQueryRepository:   params.MerchantQueryRepository,
		logger:                    params.Logger,
		observability:             params.Observability,
	}
}

func (s *merchantCommandService) CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*salemongo.Merchant, error) {
	const method = "CreateMerchant"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	user, err := s.userAdapter.FindById(ctx, request.UserID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("user_id", request.UserID))
	}

	res, err := s.merchantCommandRepository.CreateMerchant(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("user_id", request.UserID))
	}

	go func() {
		htmlBody, err := email.GenerateEmailHTML(map[string]string{
			"Title":   "Welcome to SanEdge Merchant Portal",
			"Message": "Your merchant account has been created successfully. To continue, please upload the required documents for verification. Once completed, our team will review and activate your account.",
			"Button":  "Upload Documents",
			"Link":    fmt.Sprintf("https://sanedge.example.com/merchant/%d/documents", user.UserID),
		})
		if err != nil {
			s.logger.Error("failed to generate merchant email HTML", zap.Error(err))
			return
		}

		emailPayload := map[string]any{
			"email":   user.Email,
			"subject": "Initial Verification - SanEdge",
			"body":    htmlBody,
		}

		payloadBytes, err := json.Marshal(emailPayload)
		if err != nil {
			s.logger.Error("failed to marshal email payload for new merchant", zap.Error(err), zap.Int("merchant_id", int(res.MerchantID)))
			return
		}

		if s.kafka != nil {
			err = s.kafka.SendMessage("email-service-topic-merchant-create", strconv.Itoa(int(res.MerchantID)), payloadBytes)
			if err != nil {
				s.logger.Error("failed to send merchant creation email via kafka", zap.Error(err), zap.Int("merchant_id", int(res.MerchantID)))
			}

			statsEvent := events.MerchantEvent{
				MerchantID: uint64(res.MerchantID),
				UserID:     uint64(request.UserID),
				Name:       request.Name,
				Email:      user.Email,
				Status:     "inactive",
				CreatedAt:  time.Now(),
			}

			statsPayloadByte, err := json.Marshal(statsEvent)
			if err != nil {
				s.logger.Error("failed to marshal merchant stats event", zap.Error(err), zap.Int("merchant_id", int(res.MerchantID)))
			} else {
				_ = s.kafka.SendMessage("stats-topic-merchant-events", strconv.Itoa(int(res.MerchantID)), statsPayloadByte)
			}
		}
	}()

	logSuccess("Successfully created merchant", zap.Int("merchant_id", int(res.MerchantID)))

	return res, nil
}

func (s *merchantCommandService) UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*salemongo.Merchant, error) {
	const method = "UpdateMerchant"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	res, err := s.merchantCommandRepository.UpdateMerchant(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("merchant_id", *request.MerchantID))
	}

	s.cache.DeleteCachedMerchant(ctx, int(res.MerchantID))

	logSuccess("Successfully updated merchant", zap.Int("merchant_id", int(res.MerchantID)))

	return res, nil
}

func (s *merchantCommandService) UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*salemongo.Merchant, error) {
	const method = "UpdateMerchantStatus"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	merchant, err := s.merchantQueryRepository.FindByMerchantId(ctx, *request.MerchantID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("merchant_id", *request.MerchantID))
	}

	user, err := s.userAdapter.FindById(ctx, int(merchant.UserID))
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("user_id", int(merchant.UserID)))
	}

	res, err := s.merchantCommandRepository.UpdateMerchantStatus(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("merchant_id", *request.MerchantID))
	}

	go func() {
		statusReq := request.Status
		subject := ""
		message := ""
		link := fmt.Sprintf("https://sanedge.example.com/merchant/%d/dashboard", *request.MerchantID)

		switch statusReq {
		case "active":
			subject = "Your Merchant Account is Now Active"
			message = "Congratulations! Your merchant account has been verified and is now <b>active</b>. You can now fully access all features in the SanEdge Merchant Portal."
		case "inactive":
			subject = "Merchant Account Set to Inactive"
			message = "Your merchant account status has been set to <b>inactive</b>. Please contact support if you believe this is a mistake."
		case "rejected":
			subject = "Merchant Account Rejected"
			message = "We're sorry to inform you that your merchant account has been <b>rejected</b>. Please contact support or review your submissions."
		default:
			s.logger.Error("invalid merchant status provided for email notification", zap.String("status", statusReq), zap.Int("merchant_id", *request.MerchantID))
			return
		}

		htmlBody, err := email.GenerateEmailHTML(map[string]string{
			"Title":   subject,
			"Message": message,
			"Button":  "Go to Portal",
			"Link":    link,
		})
		if err != nil {
			s.logger.Error("failed to generate merchant status email HTML", zap.Error(err))
			return
		}

		emailPayload := map[string]any{
			"email":   user.Email,
			"subject": subject,
			"body":    htmlBody,
		}

		payloadBytes, err := json.Marshal(emailPayload)
		if err != nil {
			s.logger.Error("failed to marshal email payload for merchant status update", zap.Error(err), zap.Int("merchant_id", *request.MerchantID))
			return
		}

		if s.kafka != nil {
			err = s.kafka.SendMessage("email-service-topic-merchant-update-status", strconv.Itoa(*request.MerchantID), payloadBytes)
			if err != nil {
				s.logger.Error("failed to send merchant status update email via kafka", zap.Error(err), zap.Int("merchant_id", *request.MerchantID))
			}

			statsEvent := events.MerchantEvent{
				MerchantID: uint64(res.MerchantID),
				UserID:     uint64(merchant.UserID),
				Name:       merchant.Name,
				Email:      user.Email,
				Status:     res.Status,
				CreatedAt:  time.Now(),
			}

			statsPayloadByte, err := json.Marshal(statsEvent)
			if err != nil {
				s.logger.Error("failed to marshal merchant status stats event", zap.Error(err), zap.Int("merchant_id", int(res.MerchantID)))
			} else {
				_ = s.kafka.SendMessage("stats-topic-merchant-events", strconv.Itoa(int(res.MerchantID)), statsPayloadByte)
			}
		}
	}()

	s.cache.DeleteCachedMerchant(ctx, int(res.MerchantID))

	logSuccess("Successfully updated merchant status", zap.Int("merchant_id", int(res.MerchantID)))

	return res, nil
}

func (s *merchantCommandService) TrashedMerchant(ctx context.Context, merchantID int) (*salemongo.Merchant, error) {
	const method = "TrashedMerchant"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("merchant_id", merchantID))
	defer func() { end(status, "grpc") }()

	res, err := s.merchantCommandRepository.TrashedMerchant(ctx, merchantID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("merchant_id", merchantID))
	}

	logSuccess("Successfully trashed merchant", zap.Int("merchant_id", merchantID))

	return res, nil
}

func (s *merchantCommandService) RestoreMerchant(ctx context.Context, merchantID int) (*salemongo.Merchant, error) {
	const method = "RestoreMerchant"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("merchant_id", merchantID))
	defer func() { end(status, "grpc") }()

	res, err := s.merchantCommandRepository.RestoreMerchant(ctx, merchantID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Merchant](s.logger, err, method, span, zap.Int("merchant_id", merchantID))
	}

	logSuccess("Successfully restored merchant", zap.Int("merchant_id", merchantID))

	return res, nil
}

func (s *merchantCommandService) DeleteMerchantPermanent(ctx context.Context, merchantID int) (bool, error) {
	const method = "DeleteMerchantPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("merchant_id", merchantID))
	defer func() { end(status, "grpc") }()

	_, err := s.merchantCommandRepository.DeleteMerchantPermanent(ctx, merchantID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span, zap.Int("merchant_id", merchantID))
	}

	logSuccess("Successfully deleted merchant permanently", zap.Int("merchant_id", merchantID))

	return true, nil
}

func (s *merchantCommandService) RestoreAllMerchant(ctx context.Context) (bool, error) {
	const method = "RestoreAllMerchant"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.merchantCommandRepository.RestoreAllMerchant(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, sharedErrors.ErrFailed("restore all merchants"), method, span)
	}

	logSuccess("Successfully restored all merchants")
	return true, nil
}

func (s *merchantCommandService) DeleteAllMerchantPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllMerchantPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.merchantCommandRepository.DeleteAllMerchantPermanent(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, sharedErrors.ErrFailed("delete all merchants permanently"), method, span)
	}

	logSuccess("Successfully deleted all merchants permanently")
	return true, nil
}
