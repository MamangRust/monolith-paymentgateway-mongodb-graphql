package merchantpermission

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type merchantPermission struct {
	kafka         *kafka.Kafka
	logger        logger.LoggerInterface
	requestTopic  string
	responseTopic string
	timeout       time.Duration
	responseChans map[string]chan []byte
	mu            sync.RWMutex
}

func NewMerchantPermission(
	k *kafka.Kafka,
	requestTopic, responseTopic string,
	timeout time.Duration,
	logger logger.LoggerInterface,
) MerchantPermission {
	p := &merchantPermission{
		kafka:         k,
		requestTopic:  requestTopic,
		responseTopic: responseTopic,
		timeout:       timeout,
		responseChans: make(map[string]chan []byte),
		logger:        logger,
	}

	handler := &merchantResponseHandler{validator: p}

	if k != nil {
		go func() {
			err := k.StartConsumers([]string{responseTopic}, "merchant-permission-gateway", handler)
			if err != nil {
				p.logger.Fatal("Failed to start kafka consumer", zap.Error(err))
				panic("failed to start kafka consumer: " + err.Error())
			}
		}()
	}

	return p
}

func (p *merchantPermission) ValidateMerchant(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	p.logger.Debug("Validating merchant API key")

	correlationID := uuid.NewString()
	p.logger.Info("Validating merchant via Kafka", zap.String("correlation_id", correlationID))

	respChan := make(chan []byte, 1)

	p.mu.Lock()
	p.responseChans[correlationID] = respChan
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.responseChans, correlationID)
		close(respChan)
		p.mu.Unlock()
	}()

	if err := p.sendMerchantValidationRequest(apiKey, correlationID); err != nil {
		return nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	select {
	case responseData := <-respChan:
		if responseData == nil {
			p.logger.Error("Received nil response", zap.String("correlation_id", correlationID))
			return nil, errors.New("invalid response received")
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(responseData, &payload); err != nil {
			p.logger.Error("Failed to unmarshal merchant response", zap.Error(err))
			return nil, errors.New("failed to parse merchant response")
		}

		p.logger.Info("Merchant validation success", zap.String("correlation_id", correlationID))
		return payload, nil

	case <-ctxTimeout.Done():
		p.logger.Error("Timeout waiting for Kafka response",
			zap.String("correlation_id", correlationID),
			zap.Duration("timeout", p.timeout))
		return nil, errors.New("timeout waiting for merchant validation")
	}
}

func (p *merchantPermission) sendMerchantValidationRequest(apiKey string, correlationID string) error {
	payload := map[string]interface{}{
		"api_key":        apiKey,
		"correlation_id": correlationID,
		"reply_topic":    p.responseTopic,
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

	p.logger.Info("Kafka message sent for merchant validation",
		zap.String("topic", p.requestTopic),
		zap.String("correlation_id", correlationID))

	return nil
}
