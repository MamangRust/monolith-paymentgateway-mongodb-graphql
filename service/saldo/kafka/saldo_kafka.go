package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-saldo/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"go.uber.org/zap"
)

// saldoKafkaHandler is a struct that implements the sarama.ConsumerGroupHandler interface
type saldoKafkaHandler struct {
	logger       logger.LoggerInterface
	saldoService service.SaldoCommandService
	ctx          context.Context
}

// NewSaldoKafkaHandler creates a new Kafka consumer group handler for processing saldo-related Kafka messages.
//
// It takes a saldo command service and a logger as parameters.
// The handler is responsible for consuming messages from Kafka topics related to saldo operations.
// It implements the sarama.ConsumerGroupHandler interface to manage consumer group lifecycle events.
func NewSaldoKafkaHandler(saldoService service.SaldoCommandService, logger logger.LoggerInterface, ctx context.Context) sarama.ConsumerGroupHandler {
	return &saldoKafkaHandler{
		saldoService: saldoService,
		logger:       logger,
		ctx:          ctx,
	}
}

// Setup is called when the consumer group is first initialized.
func (s *saldoKafkaHandler) Setup(session sarama.ConsumerGroupSession) error {
	s.logger.Info("saldo kafka handler setup")
	return nil
}

// Cleanup is called when the consumer group is closed.
func (s *saldoKafkaHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	s.logger.Info("saldo kafka handler cleanup")
	return nil
}

// ConsumeClaim processes incoming Kafka messages from the specified consumer group claim.
// It unmarshals each message into a payload map and creates a new saldo request.
// If a valid saldo request is found, it calls the CreateSaldo method of the saldoService
// with the created request. If the CreateSaldo method returns an error, it logs the
// error and returns an error with the message "card service error: <error message>".
// Each message is marked as processed in the consumer group session.
func (s *saldoKafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	s.logger.Info("saldo kafka handler consume claim")

	for msg := range claim.Messages() {
		ctx, cancel := context.WithTimeout(s.ctx, 20*time.Second)
		defer cancel()

		var payload map[string]interface{}

		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			return err
		}

		s.logger.Info("hello world", zap.Any("payload", payload))

		cardNumber, ok := payload["card_number"].(string)
		if !ok {
			s.logger.Error("payload card_number missing or not string", zap.Any("payload", payload))
			continue
		}

		totalBalanceFloat, ok := payload["total_balance"].(float64)
		if !ok {
			s.logger.Error("payload total_balance missing or not float64", zap.Any("payload", payload))
			continue
		}

		totalBalance := int(totalBalanceFloat)

		// Atomic idempotency: CreateSaldoIfNotExists uses ON CONFLICT DO NOTHING
		// (guarded by the partial unique index idx_saldos_card_number_active), so
		// it never overwrites an existing balance or creates a duplicate row even
		// when racing with the API create for the same card. The card-created
		// event always carries total_balance=0, which must not clobber a real
		// balance already set via the API.
		if errRes := s.saldoService.CreateSaldoIfNotExists(ctx, &requests.CreateSaldoRequest{
			CardNumber:   cardNumber,
			TotalBalance: int(totalBalance),
		}); errRes != nil {
			s.logger.Error("card service error", zap.Any("error", errRes))

			return fmt.Errorf("card service error: %v", errRes.Error())
		}
	}

	s.logger.Info("saldo kafka handler consume claim success", zap.Bool("success", true))

	return nil
}
