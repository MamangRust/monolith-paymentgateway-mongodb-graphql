package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// CardEventLogHandler stores card domain events durably for audit and replay.
type CardEventLogHandler struct {
	db     *mongo.Database
	logger logger.LoggerInterface
}

func NewCardEventLogHandler(db *mongo.Database, logger logger.LoggerInterface) *CardEventLogHandler {
	return &CardEventLogHandler{db: db, logger: logger}
}

func (h *CardEventLogHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info("card event log consumer setup")
	return nil
}

func (h *CardEventLogHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("card event log consumer cleanup")
	return nil
}

func (h *CardEventLogHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	coll := h.db.Collection("card_event_logs")

	for msg := range claim.Messages() {
		event, err := parseCardEvent(msg.Topic, msg.Key, msg.Value)
		if err != nil {
			h.logger.Error("invalid card event audit payload", zap.Error(err), zap.String("topic", msg.Topic))
			session.MarkMessage(msg, "")
			continue
		}
		if h.db == nil {
			return fmt.Errorf("card event audit database is nil")
		}
		insertCtx, cancel := context.WithTimeout(session.Context(), 10*time.Second)

		doc := &salemongo.CardEventLog{
			Topic:       event.Topic,
			EventType:   event.EventType,
			CardNumber:  nullableString(event.CardNumber),
			ReferenceID: nullableString(event.ReferenceID),
			Payload:     event.Payload,
			ReceivedAt:  time.Now(),
		}
		// Insert and ignore duplicate key errors (idempotent via unique index)
		_, err = coll.InsertOne(insertCtx, doc)
		cancel()
		if err != nil {
			// Duplicate key errors are expected on redelivery
			if mongo.IsDuplicateKeyError(err) {
				session.MarkMessage(msg, "")
				continue
			}
			return fmt.Errorf("insert card event log: %w", err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

type cardEvent struct {
	Topic       string
	EventType   string
	CardNumber  string
	ReferenceID string
	Payload     []byte
}

func parseCardEvent(topic string, key, payload []byte) (cardEvent, error) {
	if strings.TrimSpace(topic) == "" || !json.Valid(payload) {
		return cardEvent{}, fmt.Errorf("topic or payload is invalid")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil || fields == nil {
		return cardEvent{}, fmt.Errorf("payload must be a JSON object")
	}

	event := cardEvent{
		Topic:     topic,
		EventType: strings.TrimPrefix(topic, "card."),
		Payload:   append([]byte(nil), payload...),
	}
	event.CardNumber = rawString(fields, "card_number")
	switch topic {
	case "card.payment.posted":
		event.ReferenceID = rawString(fields, "payment_id")
	case "card.fraud.alert":
		event.ReferenceID = rawString(fields, "txn_id")
	case "card.statement.generated":
		event.ReferenceID = rawString(fields, "billing_id")
		if event.ReferenceID == "" {
			event.ReferenceID = rawString(fields, "statement_id")
		}
		if event.ReferenceID == "" {
			event.ReferenceID = rawString(fields, "generated_at")
		}
		if event.ReferenceID == "" {
			event.ReferenceID = rawString(fields, "billing_cycle_day")
		}
	case "card.limit.changed":
		event.ReferenceID = event.CardNumber
	}
	if event.ReferenceID == "" {
		event.ReferenceID = string(key)
	}
	return event, nil
}

func rawString(fields map[string]json.RawMessage, name string) string {
	value, ok := fields[name]
	if !ok {
		return ""
	}
	var stringValue string
	if json.Unmarshal(value, &stringValue) == nil {
		return stringValue
	}
	var numberValue json.Number
	if json.Unmarshal(value, &numberValue) == nil {
		return numberValue.String()
	}
	return ""
}

var _ sarama.ConsumerGroupHandler = (*CardEventLogHandler)(nil)
