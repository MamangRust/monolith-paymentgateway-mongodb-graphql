package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexes creates the indexes required by the payment gateway services.
func CreateIndexes(ctx context.Context, db *mongo.Database) error {
	indexes := map[string][]mongo.IndexModel{
		"users": {
			{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"roles": {
			{Keys: bson.D{{Key: "role_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "role_name", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"user_roles": {
			{Keys: bson.D{{Key: "user_role_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "role_id", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"refresh_tokens": {
			{Keys: bson.D{{Key: "refresh_token_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "token", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"reset_tokens": {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "token", Value: 1}}},
		},
		"cards": {
			{Keys: bson.D{{Key: "card_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "card_number", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"merchants": {
			{Keys: bson.D{{Key: "merchant_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "merchant_no", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "api_key", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"merchant_documents": {
			{Keys: bson.D{{Key: "document_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "merchant_id", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"saldos": {
			{Keys: bson.D{{Key: "saldo_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"topups": {
			{Keys: bson.D{{Key: "topup_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "topup_no", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "idempotency_key", Value: 1}}, Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.D{{Key: "idempotency_key", Value: bson.D{{Key: "$gt", Value: ""}}}})},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"transactions": {
			{Keys: bson.D{{Key: "transaction_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "transaction_no", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "idempotency_key", Value: 1}}, Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.D{{Key: "idempotency_key", Value: bson.D{{Key: "$gt", Value: ""}}}})},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
			{Keys: bson.D{{Key: "merchant_id", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"transfers": {
			{Keys: bson.D{{Key: "transfer_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "transfer_no", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "idempotency_key", Value: 1}}, Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.D{{Key: "idempotency_key", Value: bson.D{{Key: "$gt", Value: ""}}}})},
			{Keys: bson.D{{Key: "transfer_from", Value: 1}}},
			{Keys: bson.D{{Key: "transfer_to", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"withdraws": {
			{Keys: bson.D{{Key: "withdraw_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "withdraw_no", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "idempotency_key", Value: 1}}, Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.D{{Key: "idempotency_key", Value: bson.D{{Key: "$gt", Value: ""}}}})},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
			{Keys: bson.D{{Key: "deleted_at", Value: 1}}},
		},
		"billing_cycles": {
			{Keys: bson.D{{Key: "billing_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			// One statement per card and billing period: re-running the billing
			// scheduler for the same period must not create a second cycle (the
			// repository skips duplicate-key inserts).
			{Keys: bson.D{{Key: "card_number", Value: 1}, {Key: "cycle_start", Value: 1}, {Key: "cycle_end", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
		},
		"card_auth_transactions": {
			{Keys: bson.D{{Key: "auth_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "txn_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
		},
		"card_payments": {
			{Keys: bson.D{{Key: "payment_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "payment_uuid", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
		},
		"card_rewards": {
			{Keys: bson.D{{Key: "reward_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
		},
		"balance_ledger": {
			{Keys: bson.D{{Key: "entry_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "operation_id", Value: 1}}},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
		},
		"reconciliation_queue": {
			{Keys: bson.D{{Key: "queue_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "saldo_id", Value: 1}}},
			{Keys: bson.D{{Key: "status", Value: 1}}},
		},
		"outbox_events": {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "event_key", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "status", Value: 1}, {Key: "next_attempt_at", Value: 1}}},
		},
		"consumer_inbox": {
			{Keys: bson.D{{Key: "consumer_name", Value: 1}, {Key: "event_key", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"card_event_logs": {
			{Keys: bson.D{{Key: "event_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "card_number", Value: 1}}},
		},
		"idempotency_records": {
			{Keys: bson.D{{Key: "scope", Value: 1}, {Key: "idempotency_key", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		},
	}

	for coll, models := range indexes {
		if len(models) == 0 {
			continue
		}
		if _, err := db.Collection(coll).Indexes().CreateMany(ctx, models); err != nil {
			return err
		}
	}
	return nil
}
