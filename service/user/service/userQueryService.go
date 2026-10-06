package service

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errorhandler"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	mencache "github.com/MamangRust/monolith-payment-gateway-user/redis"
	"github.com/MamangRust/monolith-payment-gateway-user/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type userQueryDeps struct {
	Cache         mencache.UserQueryCache
	Repository    repository.UserQueryRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type userQueryService struct {
	cache               mencache.UserQueryCache
	userQueryRepository repository.UserQueryRepository
	logger              logger.LoggerInterface
	observability       observability.TraceLoggerObservability
}

func NewUserQueryService(params *userQueryDeps) UserQueryService {
	return &userQueryService{
		cache:               params.Cache,
		userQueryRepository: params.Repository,
		logger:              params.Logger,
		observability:       params.Observability,
	}
}

func (s *userQueryService) FindAll(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, error) {
	const method = "FindAll"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCachedUsersCache(ctx, req); found {
		logSuccess("Successfully retrieved all user records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	users, err := s.userQueryRepository.FindAllUsers(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.User](
			s.logger,
			sharedErrors.ErrFailed("find all users"),
			method,
			span,
			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	totalCount64, err := s.userQueryRepository.CountAllUsers(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.User](
			s.logger,
			sharedErrors.ErrFailed("count users"),
			method,
			span,
		)
	}
	totalCount := int(totalCount64)

	s.cache.SetCachedUsersCache(ctx, req, users, &totalCount)

	logSuccess("Successfully fetched user",
		zap.Int("totalRecords", totalCount),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize))

	return users, &totalCount, nil
}

func (s *userQueryService) FindByID(ctx context.Context, id int) (*salemongo.User, error) {
	const method = "FindByID"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("user_id", id))

	defer func() {
		end(status, "grpc")
	}()

	if data, found := s.cache.GetCachedUserCache(ctx, id); found {
		logSuccess("Successfully retrieved user record from cache", zap.Int("user.id", id))
		return data, nil
	}

	user, err := s.userQueryRepository.FindById(ctx, id)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.User](
			s.logger,
			sharedErrors.ErrNotFoundResponse("User"),
			method,
			span,
			zap.Int("user_id", id),
		)
	}

	s.cache.SetCachedUserCache(ctx, user)

	logSuccess("Successfully fetched user", zap.Int("user_id", id))

	return user, nil
}

func (s *userQueryService) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, error) {
	const method = "FindByActive"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCachedUserActiveCache(ctx, req); found {
		logSuccess("Successfully retrieved active user records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	users, err := s.userQueryRepository.FindByActive(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.User](
			s.logger,
			sharedErrors.ErrFailed("find active users"),
			method,
			span,
			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	totalCount64, err := s.userQueryRepository.CountActiveUsers(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.User](
			s.logger,
			sharedErrors.ErrFailed("count active users"),
			method,
			span,
		)
	}
	totalCount := int(totalCount64)

	s.cache.SetCachedUserActiveCache(ctx, req, users, &totalCount)

	logSuccess("Successfully fetched active user",
		zap.Int("totalRecords", totalCount),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize))

	return users, &totalCount, nil
}

func (s *userQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*salemongo.User, *int, error) {
	const method = "FindByTrashed"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.String("search", search))

	defer func() {
		end(status, "grpc")
	}()

	if data, total, found := s.cache.GetCachedUserTrashedCache(ctx, req); found {
		return data, total, nil
	}

	users, err := s.userQueryRepository.FindByTrashed(ctx, req)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.User](
			s.logger,
			sharedErrors.ErrFailed("find trashed users"),
			method,
			span,
			zap.Int("page", page),
			zap.Int("pageSize", pageSize),
			zap.String("search", search),
		)
	}

	totalCount64, err := s.userQueryRepository.CountTrashedUsers(ctx, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*salemongo.User](
			s.logger,
			sharedErrors.ErrFailed("count trashed users"),
			method,
			span,
		)
	}
	totalCount := int(totalCount64)

	s.cache.SetCachedUserTrashedCache(ctx, req, users, &totalCount)

	logSuccess("Successfully fetched trashed user",
		zap.Int("totalRecords", totalCount),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize))

	return users, &totalCount, nil
}

func (s *userQueryService) FindByEmail(ctx context.Context, email string) (*salemongo.User, error) {
	const method = "FindByEmail"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("email", email))

	defer func() {
		end(status, "grpc")
	}()

	user, err := s.userQueryRepository.FindByEmail(ctx, email)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.User](
			s.logger,
			sharedErrors.ErrNotFoundResponse("User"),
			method,
			span,
			zap.String("email", email),
		)
	}

	logSuccess("Successfully fetched user by email", zap.String("email", email))

	return user, nil
}

func (s *userQueryService) FindByVerificationCode(ctx context.Context, code string) (*salemongo.User, error) {
	const method = "FindByVerificationCode"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.String("verification_code", code))

	defer func() {
		end(status, "grpc")
	}()

	user, err := s.userQueryRepository.FindByVerificationCode(ctx, code)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*salemongo.User](
			s.logger,
			sharedErrors.ErrNotFoundResponse("User"),
			method,
			span,
			zap.String("verification_code", code),
		)
	}

	logSuccess("Successfully fetched user by verification code", zap.String("verification_code", code))

	return user, nil
}

func (s *userQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
