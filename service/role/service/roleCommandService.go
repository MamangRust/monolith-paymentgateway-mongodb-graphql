package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	mencache "github.com/MamangRust/monolith-payment-gateway-role/redis"
	"github.com/MamangRust/monolith-payment-gateway-role/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type roleCommandDeps struct {
	Ctx           context.Context
	Cache         mencache.RoleCommandCache
	Repository    repository.RoleCommandRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type roleCommandService struct {
	mencache      mencache.RoleCommandCache
	roleCommand   repository.RoleCommandRepository
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewRoleCommandService(params *roleCommandDeps) RoleCommandService {
	return &roleCommandService{
		mencache:      params.Cache,
		roleCommand:   params.Repository,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *roleCommandService) CreateRole(ctx context.Context, request *requests.CreateRoleRequest) (*salemongo.Role, error) {
	const method = "CreateRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("roleName", request.Name))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting CreateRole process",
		zap.String("roleName", request.Name),
	)

	role, err := s.roleCommand.CreateRole(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Role](
			s.logger,
			err,
			method,
			span,
			zap.String("role_name", request.Name),
		)
	}

	logSuccess("CreateRole process completed",
		zap.String("roleName", request.Name),
		zap.Int("roleID", int(role.RoleID)),
	)

	return role, nil
}

func (s *roleCommandService) UpdateRole(ctx context.Context, request *requests.UpdateRoleRequest) (*salemongo.Role, error) {
	const method = "UpdateRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("roleID", *request.ID),
		attribute.String("newRoleName", request.Name))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting UpdateRole process",
		zap.Int("roleID", *request.ID),
		zap.String("newRoleName", request.Name),
	)

	role, err := s.roleCommand.UpdateRole(ctx, request)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Role](
			s.logger,
			err,
			method,
			span,
			zap.Int("role_id", *request.ID),
			zap.String("new_name", request.Name),
		)
	}

	logSuccess("UpdateRole process completed",
		zap.Int("roleID", *request.ID),
		zap.String("newRoleName", request.Name),
	)

	return role, nil
}

func (s *roleCommandService) TrashedRole(ctx context.Context, id int) (*salemongo.Role, error) {
	const method = "TrashedRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("roleID", id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting TrashedRole process",
		zap.Int("roleID", id),
	)

	role, err := s.roleCommand.TrashedRole(ctx, id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Role](
			s.logger,
			err,
			method,
			span,
			zap.Int("role_id", id),
		)
	}

	logSuccess("TrashedRole process completed",
		zap.Int("roleID", id),
	)

	return role, nil
}

func (s *roleCommandService) RestoreRole(ctx context.Context, id int) (*salemongo.Role, error) {
	const method = "RestoreRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("roleID", id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting RestoreRole process",
		zap.Int("roleID", id),
	)

	role, err := s.roleCommand.RestoreRole(ctx, id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.Role](
			s.logger,
			err,
			method,
			span,
			zap.Int("role_id", id),
		)
	}

	logSuccess("RestoreRole process completed",
		zap.Int("roleID", id),
	)

	return role, nil
}

func (s *roleCommandService) DeleteRolePermanent(ctx context.Context, id int) (bool, error) {
	const method = "DeleteRolePermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("roleID", id))

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Starting DeleteRolePermanent process",
		zap.Int("roleID", id),
	)

	_, err := s.roleCommand.DeleteRolePermanent(ctx, id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			err,
			method,
			span,
			zap.Int("role_id", id),
		)
	}

	logSuccess("DeleteRolePermanent process completed",
		zap.Int("roleID", id),
	)

	return true, nil
}

func (s *roleCommandService) RestoreAllRole(ctx context.Context) (bool, error) {
	const method = "RestoreAllRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Restoring all roles")

	_, err := s.roleCommand.RestoreAllRole(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("restore all roles"),
			method,
			span,
		)
	}

	logSuccess("Successfully restored all roles")
	return true, nil
}

func (s *roleCommandService) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllRolePermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status, "grpc")
	}()

	s.logger.Debug("Permanently deleting all roles")

	_, err := s.roleCommand.DeleteAllRolePermanent(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](
			s.logger,
			sharedErrors.ErrFailed("delete role permanently"),
			method,
			span,
		)
	}

	logSuccess("Successfully deleted all roles permanently")
	return true, nil
}
