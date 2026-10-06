package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-role/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type userRoleDeps struct {
	Repository    repository.UserRoleCommandRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type userRoleService struct {
	userRole      repository.UserRoleCommandRepository
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewUserRoleService(params *userRoleDeps) UserRoleService {
	return &userRoleService{
		userRole:      params.Repository,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *userRoleService) AssignRoleToUser(ctx context.Context, userID, roleID int) (*salemongo.UserRole, error) {
	const method = "AssignRoleToUser"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("userID", userID),
		attribute.Int("roleID", roleID))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting AssignRoleToUser process",
		zap.Int("userID", userID),
		zap.Int("roleID", roleID),
	)

	userRole, err := s.userRole.AssignRoleToUser(ctx, userID, roleID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.UserRole](
			s.logger,
			err,
			method,
			span,
			zap.Int("user_id", userID),
			zap.Int("role_id", roleID),
		)
	}

	logSuccess("AssignRoleToUser process completed",
		zap.Int("userID", userID),
		zap.Int("roleID", roleID),
	)

	return userRole, nil
}

func (s *userRoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	const method = "RemoveRoleFromUser"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("userID", userID),
		attribute.Int("roleID", roleID))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting RemoveRoleFromUser process",
		zap.Int("userID", userID),
		zap.Int("roleID", roleID),
	)

	_, err := s.userRole.RemoveRoleFromUser(ctx, userID, roleID)
	if err != nil {
		status = "error"
		_, err := errorhandler.HandleError[bool](
			s.logger,
			err,
			method,
			span,
			zap.Int("user_id", userID),
			zap.Int("role_id", roleID),
		)
		return err
	}

	logSuccess("RemoveRoleFromUser process completed",
		zap.Int("userID", userID),
		zap.Int("roleID", roleID),
	)

	return nil
}
