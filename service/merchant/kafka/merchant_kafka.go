package myhandlerkafka

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/MamangRust/monolith-payment-gateway-merchant/service"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/response"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.uber.org/zap"
)

// merchantKafkaHandler is a struct that implements the sarama.ConsumerGroupHandler interface
type merchantKafkaHandler struct {
	logger          logger.LoggerInterface
	merchantService service.MerchantQueryService
	kafka           *kafka.Kafka
}

// NewMerchantKafkaHandler creates a new Kafka consumer group handler for processing merchant API key validation responses.
//
// It takes a merchant query service, a Kafka producer, and a logger as parameters.
// The handler is used to process incoming Kafka messages from the merchant API key validation response topic.
// It implements the sarama.ConsumerGroupHandler interface to handle consumer group lifecycle events.
func NewMerchantKafkaHandler(merchantService service.MerchantQueryService, kafka *kafka.Kafka, logger logger.LoggerInterface) sarama.ConsumerGroupHandler {
	return &merchantKafkaHandler{
		merchantService: merchantService,
		kafka:           kafka,
		logger:          logger,
	}
}

// Setup is called at the beginning of a new Kafka consumer group session.
//
// It can be used to initialize resources or state before message consumption begins.
// In this implementation, it performs no setup and returns nil.
func (m *merchantKafkaHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called at the end of a Kafka consumer group session.
//
// It can be used to release resources allocated during Setup or message consumption.
// In this implementation, it performs no cleanup and returns nil.
func (m *merchantKafkaHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes incoming Kafka messages from the specified consumer group claim.
// It unmarshals each message into a payload map and retrieves the correlation ID.
// If a valid correlation ID is found, it checks if the API key is valid by calling
// the FindByApiKey method of the merchantService. If the API key is valid, it sends
// a valid response to the corresponding Kafka topic. Each message is marked as processed
// in the consumer group session.
func (m *merchantKafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// NOTE: create a FRESH context per message. A consumer-group claim stays
		// open for minutes/hours, so a context created once at ConsumeClaim start
		// is already expired by the time the message is processed, which made
		// every FindByApiKey (Redis GET + DB query) fail instantly with
		// "context deadline exceeded" -> Valid:false -> HTTP 401 at the gateway.
		ctx, cancel := context.WithTimeout(session.Context(), 10*time.Second)

		var payload requests.MerchantRequestPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			cancel()
			m.logger.Error("Failed to unmarshal merchant request", zap.Error(err))
			// This message cannot be processed successfully; acknowledge it so a
			// poison message does not block the partition forever.
			session.MarkMessage(msg, "malformed merchant API-key request")
			continue
		}

		resp := response.MerchantResponsePayload{
			CorrelationID: payload.CorrelationID,
			Valid:         false,
		}

		merchant, err := m.merchantService.FindByApiKey(ctx, payload.ApiKey)
		cancel()
		if err != nil {
			// A missing merchant is a normal invalid-key result. Infrastructure
			// failures (Redis, DB, or gRPC) must be retried instead of being
			// acknowledged as an invalid key and permanently turning into 401.
			var appErr *sharedErrors.AppError
			if !stderrors.As(err, &appErr) || appErr.Type != sharedErrors.ErrorTypeNotFound {
				m.logger.Error("Merchant API-key lookup failed", zap.Error(err))
				return fmt.Errorf("find merchant by API key: %w", err)
			}
		} else if merchant != nil {
			resp.Valid = true
			resp.MerchantID = int64(merchant.MerchantID)
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			m.logger.Error("Failed to marshal merchant response", zap.Error(err))
			return fmt.Errorf("marshal merchant response: %w", err)
		}
		sendErr := m.kafka.SendMessage(payload.ReplyTopic, payload.CorrelationID, respBytes)
		if sendErr != nil {
			// Do not acknowledge the request: the gateway is waiting for this
			// correlation ID, and Sarama must redeliver it after the consumer
			// session retries.
			m.logger.Error("Failed to send Kafka response", zap.Error(sendErr))
			return fmt.Errorf("send merchant response: %w", sendErr)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
