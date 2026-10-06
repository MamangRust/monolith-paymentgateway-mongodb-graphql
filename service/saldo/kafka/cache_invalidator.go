package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	mencache "github.com/MamangRust/monolith-payment-gateway-saldo/redis"
	"go.uber.org/zap"
)

// saldoCacheInvalidator implements sarama.ConsumerGroupHandler and deletes the
// saldo query cache keys when a saldo.changed event is observed. Deleting a
// missing key is a no-op, so replaying events (at-least-once delivery, consumer
// restart) is safe.
//
// P1.4: the synchronous DeleteSaldoCache* in the command service is the fast
// path; this consumer is the durable path that survives a crash between the DB
// commit and the synchronous delete.
type saldoCacheInvalidator struct {
	logger  logger.LoggerInterface
	cache   mencache.SaldoCommandCache
	timeout time.Duration
}

// NewSaldoCacheInvalidator creates the durable cache invalidation consumer.
func NewSaldoCacheInvalidator(cache mencache.SaldoCommandCache, logger logger.LoggerInterface) sarama.ConsumerGroupHandler {
	return &saldoCacheInvalidator{
		logger:  logger,
		cache:   cache,
		timeout: 20 * time.Second,
	}
}

func (h *saldoCacheInvalidator) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info("saldo cache invalidator consumer setup")
	return nil
}

func (h *saldoCacheInvalidator) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("saldo cache invalidator consumer cleanup")
	return nil
}

func (h *saldoCacheInvalidator) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		ctx, cancel := context.WithTimeout(context.Background(), h.timeout)

		var payload struct {
			SaldoID      int64  `json:"saldo_id"`
			CardNumber   string `json:"card_number"`
			TotalBalance int64  `json:"total_balance"`
		}
		unmarshalErr := json.Unmarshal(msg.Value, &payload)
		if unmarshalErr == nil && payload.CardNumber != "" {
			h.cache.DeleteSaldoCacheByCardNumber(ctx, payload.CardNumber)
			if payload.SaldoID > 0 {
				h.cache.DeleteSaldoCache(ctx, int(payload.SaldoID))
			}
			h.logger.Debug("saldo cache invalidated from event",
				zap.String("card_number", maskCard(payload.CardNumber)),
				zap.Int64("saldo_id", payload.SaldoID),
			)
		} else if unmarshalErr != nil {
			h.logger.Warn("saldo cache invalidator: failed to parse event payload", zap.Error(unmarshalErr))
		}

		cancel()
		session.MarkMessage(msg, "")
	}
	return nil
}

func maskCard(cardNumber string) string {
	if len(cardNumber) <= 8 {
		return "****"
	}
	return cardNumber[:6] + "****" + cardNumber[len(cardNumber)-4:]
}
