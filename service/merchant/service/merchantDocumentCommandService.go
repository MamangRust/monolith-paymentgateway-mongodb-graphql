package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	cache "github.com/MamangRust/monolith-payment-gateway-merchant/redis"
	"github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/email"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.uber.org/zap"
)

type merchantDocumentCommandDeps struct {
	Kafka                   *kafka.Kafka
	Cache                   cache.MerchantDocumentCommandCache
	CommandRepository       repository.MerchantDocumentCommandRepository
	MerchantQueryRepository repository.MerchantQueryRepository
	UserAdapter             adapter.UserAdapter
	Logger                  logger.LoggerInterface
	Observability           observability.TraceLoggerObservability
}

type merchantDocumentCommandService struct {
	kafka         *kafka.Kafka
	cache         cache.MerchantDocumentCommandCache
	commandRepo   repository.MerchantDocumentCommandRepository
	merchantRepo  repository.MerchantQueryRepository
	userAdapter   adapter.UserAdapter
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewMerchantDocumentCommandService(params *merchantDocumentCommandDeps) MerchantDocumentCommandService {
	return &merchantDocumentCommandService{
		kafka:         params.Kafka,
		cache:         params.Cache,
		commandRepo:   params.CommandRepository,
		merchantRepo:  params.MerchantQueryRepository,
		userAdapter:   params.UserAdapter,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *merchantDocumentCommandService) CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*salemongo.MerchantDocument, error) {
	const method = "CreateMerchantDocument"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	merchant, err := s.merchantRepo.FindByMerchantId(ctx, request.MerchantID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("merchant_id", request.MerchantID))
	}

	user, err := s.userAdapter.FindById(ctx, int(merchant.UserID))
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("user_id", int(merchant.UserID)))
	}

	merchantDocument, err := s.commandRepo.CreateMerchantDocument(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("merchant_id", request.MerchantID))
	}

	go func() {
		if s.kafka == nil {
			return
		}

		htmlBody, err := email.GenerateEmailHTML(map[string]string{
			"Title":   "Welcome to SanEdge Merchant Portal",
			"Message": "Thank you for registering your merchant account. Your account is currently <b>inactive</b> and under initial review. To proceed, please upload all required documents for verification. Once your documents are submitted, our team will review them and activate your account accordingly.",
			"Button":  "Upload Documents",
			"Link":    fmt.Sprintf("https://sanedge.example.com/merchant/%d/documents", user.UserID),
		})
		if err != nil {
			s.logger.Error("failed to generate merchant document email HTML", zap.Error(err))
			return
		}

		emailPayload := map[string]any{
			"email":   user.Email,
			"subject": "Merchant Verification Pending - Action Required",
			"body":    htmlBody,
		}

		payloadBytes, err := json.Marshal(emailPayload)
		if err != nil {
			s.logger.Error("failed to marshal email payload for new merchant document", zap.Error(err), zap.Int("document_id", int(merchantDocument.DocumentID)))
			return
		}

		err = s.kafka.SendMessage("email-service-topic-merchant-document-create", strconv.Itoa(int(merchantDocument.DocumentID)), payloadBytes)
		if err != nil {
			s.logger.Error("failed to send merchant document creation email via kafka", zap.Error(err), zap.Int("document_id", int(merchantDocument.DocumentID)))
		}
	}()

	logSuccess("Successfully created merchant document", zap.Int("document_id", int(merchantDocument.DocumentID)))

	return merchantDocument, nil
}

func (s *merchantDocumentCommandService) UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*salemongo.MerchantDocument, error) {
	const method = "UpdateMerchantDocument"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	merchantDocument, err := s.commandRepo.UpdateMerchantDocument(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("document_id", *request.DocumentID))
	}

	s.cache.DeleteCachedMerchantDocuments(ctx, int(merchantDocument.DocumentID))

	logSuccess("Successfully updated merchant document", zap.Int("document_id", int(merchantDocument.DocumentID)))

	return merchantDocument, nil
}

func (s *merchantDocumentCommandService) UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*salemongo.MerchantDocument, error) {
	const method = "UpdateMerchantDocumentStatus"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	merchant, err := s.merchantRepo.FindByMerchantId(ctx, request.MerchantID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("merchant_id", request.MerchantID))
	}

	user, err := s.userAdapter.FindById(ctx, int(merchant.UserID))
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("user_id", int(merchant.UserID)))
	}

	merchantDocument, err := s.commandRepo.UpdateMerchantDocumentStatus(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("merchant_id", request.MerchantID))
	}

	go func() {
		if s.kafka == nil {
			return
		}

		statusReq := request.Status
		note := request.Note
		subject := ""
		message := ""
		buttonLabel := ""
		link := fmt.Sprintf("https://sanedge.example.com/merchant/%d/documents", request.MerchantID)

		switch statusReq {
		case "pending":
			subject = "Merchant Document Status: Pending Review"
			message = "Your merchant documents have been submitted and are currently pending review."
			buttonLabel = "View Documents"
		case "approved":
			subject = "Merchant Document Status: Approved"
			message = "Congratulations! Your merchant documents have been approved. Your account is now active and fully functional."
			buttonLabel = "Go to Dashboard"
			link = fmt.Sprintf("https://sanedge.example.com/merchant/%d/dashboard", request.MerchantID)
		case "rejected":
			subject = "Merchant Document Status: Rejected"
			message = "Unfortunately, your merchant documents were rejected. Please review the feedback below and re-upload the necessary documents."
			buttonLabel = "Re-upload Documents"
		default:
			s.logger.Error("invalid merchant document status provided for email notification", zap.String("status", statusReq), zap.Int("merchant_id", request.MerchantID))
			return
		}

		if note != "" {
			message += fmt.Sprintf(`<br><br><b>Reviewer Note:</b><br><i>%s</i>`, note)
		}

		htmlBody, err := email.GenerateEmailHTML(map[string]string{
			"Title":   subject,
			"Message": message,
			"Button":  buttonLabel,
			"Link":    link,
		})
		if err != nil {
			s.logger.Error("failed to generate merchant document status email HTML", zap.Error(err))
			return
		}

		emailPayload := map[string]any{
			"email":   user.Email,
			"subject": subject,
			"body":    htmlBody,
		}

		payloadBytes, err := json.Marshal(emailPayload)
		if err != nil {
			s.logger.Error("failed to marshal email payload for merchant document status update", zap.Error(err), zap.Int("merchant_id", request.MerchantID))
			return
		}

		err = s.kafka.SendMessage("email-service-topic-merchant-document-update-status", strconv.Itoa(request.MerchantID), payloadBytes)
		if err != nil {
			s.logger.Error("failed to send merchant document status update email via kafka", zap.Error(err), zap.Int("merchant_id", request.MerchantID))
		}
	}()

	s.cache.DeleteCachedMerchantDocuments(ctx, int(merchantDocument.DocumentID))

	logSuccess("Successfully updated merchant document status", zap.Int("document_id", int(merchantDocument.DocumentID)))

	return merchantDocument, nil
}

func (s *merchantDocumentCommandService) TrashedMerchantDocument(ctx context.Context, documentID int) (*salemongo.MerchantDocument, error) {
	const method = "TrashedMerchantDocument"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	res, err := s.commandRepo.TrashedMerchantDocument(ctx, documentID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("document_id", documentID))
	}

	s.cache.DeleteCachedMerchantDocuments(ctx, documentID)

	logSuccess("Successfully trashed document", zap.Int("document_id", documentID))

	return res, nil
}

func (s *merchantDocumentCommandService) RestoreMerchantDocument(ctx context.Context, documentID int) (*salemongo.MerchantDocument, error) {
	const method = "RestoreMerchantDocument"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	res, err := s.commandRepo.RestoreMerchantDocument(ctx, documentID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.MerchantDocument](s.logger, err, method, span, zap.Int("document_id", documentID))
	}

	logSuccess("Successfully restored document", zap.Int("document_id", documentID))

	return res, nil
}

func (s *merchantDocumentCommandService) DeleteMerchantDocumentPermanent(ctx context.Context, documentID int) (bool, error) {
	const method = "DeleteMerchantDocumentPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.commandRepo.DeleteMerchantDocumentPermanent(ctx, documentID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span, zap.Int("document_id", documentID))
	}

	logSuccess("Successfully deleted document permanently", zap.Int("document_id", documentID))

	return true, nil
}

func (s *merchantDocumentCommandService) RestoreAllMerchantDocument(ctx context.Context) (bool, error) {
	const method = "RestoreAllMerchantDocument"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.commandRepo.RestoreAllMerchantDocument(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}

	logSuccess("Successfully restored all documents")
	return true, nil
}

func (s *merchantDocumentCommandService) DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllMerchantDocumentPermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() { end(status, "grpc") }()

	_, err := s.commandRepo.DeleteAllMerchantDocumentPermanent(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}

	logSuccess("Successfully deleted all documents permanently")
	return true, nil
}
