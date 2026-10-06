package rolepermission

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	mencache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type rolePermission struct {
	kafka         *kafka.Kafka
	logger        logger.LoggerInterface
	requestTopic  string
	responseTopic string
	timeout       time.Duration
	responseChans map[string]chan *response.RoleResponsePayload
	mu            sync.RWMutex
	cache         mencache.RoleCache
}

func NewRolePermission(
	k *kafka.Kafka,
	requestTopic, responseTopic string,
	timeout time.Duration,
	logger logger.LoggerInterface,
	cache mencache.RoleCache,
) RolePermission {
	p := &rolePermission{
		kafka:         k,
		requestTopic:  requestTopic,
		responseTopic: responseTopic,
		timeout:       timeout,
		cache:         cache,
		responseChans: make(map[string]chan *response.RoleResponsePayload),
		logger:        logger,
	}

	handler := &roleResponseHandler{validator: p}

	if k != nil {
		go func() {
			err := k.StartConsumers([]string{responseTopic}, "role-permission-gateway", handler)
			if err != nil {
				p.logger.Fatal("Failed to start kafka consumer", zap.Error(err))
				panic("failed to start kafka consumer: " + err.Error())
			}
		}()
	}

	return p
}

func (p *rolePermission) ValidateRole(ctx context.Context, userID int) ([]string, error) {
	p.logger.Debug("Validating user role", zap.Int("user_id", userID))

	if roles, found := p.cache.GetRoleCache(ctx, strconv.Itoa(userID)); found {
		p.logger.Debug("Role found in cache", zap.Int("user_id", userID), zap.Strings("roles", roles))
		return roles, nil
	}

	correlationID := uuid.NewString()
	p.logger.Info("Validating user role via Kafka",
		zap.Int("user_id", userID),
		zap.String("correlation_id", correlationID))

	respChan := make(chan *response.RoleResponsePayload, 1)

	p.mu.Lock()
	p.responseChans[correlationID] = respChan
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.responseChans, correlationID)
		close(respChan)
		p.mu.Unlock()
	}()

	if err := p.sendValidationRequest(userID, correlationID); err != nil {
		return nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	select {
	case roleResponse := <-respChan:
		if roleResponse == nil {
			p.logger.Error("Received nil response", zap.String("correlation_id", correlationID))
			return nil, errors.New("invalid response received")
		}

		if !roleResponse.Valid || len(roleResponse.RoleNames) == 0 {
			p.logger.Debug("Role validation failed",
				zap.Int("user_id", userID),
				zap.String("correlation_id", correlationID),
				zap.Bool("valid", roleResponse.Valid),
				zap.Int("role_count", len(roleResponse.RoleNames)))
			return nil, errors.New("role validation failed")
		}

		p.logger.Info("Role validation success",
			zap.Int("user_id", userID),
			zap.Strings("roles", roleResponse.RoleNames),
			zap.String("correlation_id", correlationID))

		// Cache the result
		p.cache.SetRoleCache(ctx, strconv.Itoa(userID), roleResponse.RoleNames)

		return roleResponse.RoleNames, nil

	case <-ctxTimeout.Done():
		p.logger.Error("Timeout waiting for Kafka response",
			zap.String("correlation_id", correlationID),
			zap.Duration("timeout", p.timeout))
		return nil, errors.New("timeout waiting for role validation")
	}
}

func (p *rolePermission) CheckRole(ctx context.Context, userID int, requiredRoles ...string) error {
	roles, err := p.ValidateRole(ctx, userID)
	if err != nil {
		return err
	}

	for _, userRole := range roles {
		for _, required := range requiredRoles {
			if userRole == required {
				return nil
			}
		}
	}

	p.logger.Debug("User does not have required role",
		zap.Int("user_id", userID),
		zap.Strings("user_roles", roles),
		zap.Strings("required_roles", requiredRoles))

	return errors.New("role not permitted")
}

func (p *rolePermission) sendValidationRequest(userID int, correlationID string) error {
	payload := requests.RoleRequestPayload{
		UserID:        userID,
		CorrelationID: correlationID,
		ReplyTopic:    p.responseTopic,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error("Failed to encode payload", zap.Error(err), zap.String("correlation_id", correlationID))
		return errors.New("failed to encode payload")
	}

	err = p.kafka.SendMessage(p.requestTopic, correlationID, data)
	if err != nil {
		p.logger.Error("Failed to send Kafka message", zap.Error(err), zap.String("correlation_id", correlationID))
		return errors.New("failed to send Kafka message")
	}

	p.logger.Info("Kafka message sent for role validation",
		zap.String("topic", p.requestTopic),
		zap.String("correlation_id", correlationID))

	return nil
}
