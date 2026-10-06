package mongo

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// IsDuplicateKeyError checks if the error is a MongoDB duplicate key error.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, we := range writeErr.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

// Pipeline lets services build aggregation stages.
type Pipeline = mongo.Pipeline

// BalanceLedger is an append-only record of every balance mutation.
type BalanceLedger struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	EntryID       int64              `bson:"entry_id"`
	OperationID   string             `bson:"operation_id"`
	CardNumber    string             `bson:"card_number"`
	Direction     string             `bson:"direction"`
	Amount        int64              `bson:"amount"`
	Delta         int64              `bson:"delta"`
	BalanceBefore int64              `bson:"balance_before"`
	BalanceAfter  int64              `bson:"balance_after"`
	SourceType    string             `bson:"source_type"`
	SourceID      *string            `bson:"source_id,omitempty"`
	CreatedAt     *time.Time         `bson:"created_at"`
}

// ReconciliationQueue holds balance mismatches discovered by the reconciler.
type ReconciliationQueue struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty"`
	QueueID               int64              `bson:"queue_id"`
	SaldoID               int32              `bson:"saldo_id"`
	CardNumber            string             `bson:"card_number"`
	CurrentBalance        int64              `bson:"current_balance"`
	LedgerBalance         int64              `bson:"ledger_balance"`
	Difference            int64              `bson:"difference"`
	LedgerEntries         int64              `bson:"ledger_entries"`
	Status                string             `bson:"status"`
	MismatchCount         int64              `bson:"mismatch_count"`
	FirstSeenAt           *time.Time         `bson:"first_seen_at"`
	LastSeenAt            *time.Time         `bson:"last_seen_at"`
	ResolvedAt            *time.Time         `bson:"resolved_at,omitempty"`
	ResolutionOperationID *string            `bson:"resolution_operation_id,omitempty"`
	ResolutionNote        *string            `bson:"resolution_note,omitempty"`
}

// ---------------------------------------------------------------------------
// Core entities (bson-tagged, collection-backed)
// ---------------------------------------------------------------------------

// User maps to the "users" collection.
type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	UserID           int32              `bson:"user_id"`
	Firstname        string             `bson:"firstname"`
	Lastname         string             `bson:"lastname"`
	Email            string             `bson:"email"`
	Password         string             `bson:"password"`
	VerificationCode string             `bson:"verification_code"`
	IsVerified       *bool              `bson:"is_verified"`
	CreatedAt        *time.Time         `bson:"created_at"`
	UpdatedAt        *time.Time         `bson:"updated_at"`
	DeletedAt        *time.Time         `bson:"deleted_at,omitempty"`
}

func (User) CollectionName() string { return "users" }

// Role maps to the "roles" collection.
type Role struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	RoleID    int32              `bson:"role_id"`
	RoleName  string             `bson:"role_name"`
	CreatedAt *time.Time         `bson:"created_at"`
	UpdatedAt *time.Time         `bson:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty"`
}

func (Role) CollectionName() string { return "roles" }

// UserRole maps to the "user_roles" collection.
type UserRole struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserRoleID int32              `bson:"user_role_id"`
	UserID     int32              `bson:"user_id"`
	RoleID     int32              `bson:"role_id"`
	CreatedAt  *time.Time         `bson:"created_at"`
	UpdatedAt  *time.Time         `bson:"updated_at"`
	DeletedAt  *time.Time         `bson:"deleted_at,omitempty"`
}

func (UserRole) CollectionName() string { return "user_roles" }

// Card maps to the "cards" collection.
type Card struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	CardID             int32              `bson:"card_id"`
	UserID             int32              `bson:"user_id"`
	CardNumber         string             `bson:"card_number"`
	CardType           string             `bson:"card_type"`
	ExpireDate         time.Time          `bson:"expire_date"`
	Cvv                string             `bson:"cvv"`
	CardProvider       string             `bson:"card_provider"`
	Status             string             `bson:"status"`
	CreditLimit        int32              `bson:"credit_limit"`
	OutstandingBalance int32              `bson:"outstanding_balance"`
	RewardPoints       int32              `bson:"reward_points"`
	EventVersion       int64              `bson:"event_version"`
	CreatedAt          *time.Time         `bson:"created_at"`
	UpdatedAt          *time.Time         `bson:"updated_at"`
	DeletedAt          *time.Time         `bson:"deleted_at,omitempty"`
}

func (Card) CollectionName() string { return "cards" }

// Merchant maps to the "merchants" collection.
type Merchant struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	MerchantID   int32              `bson:"merchant_id"`
	MerchantNo   uuid.UUID          `bson:"merchant_no"`
	Name         string             `bson:"name"`
	ApiKey       string             `bson:"api_key"`
	UserID       int32              `bson:"user_id"`
	Status       string             `bson:"status"`
	EventVersion int64              `bson:"event_version"`
	CreatedAt    *time.Time         `bson:"created_at"`
	UpdatedAt    *time.Time         `bson:"updated_at"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty"`
}

func (Merchant) CollectionName() string { return "merchants" }

// MerchantDocument maps to the "merchant_documents" collection.
type MerchantDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	DocumentID   int32              `bson:"document_id"`
	MerchantID   int32              `bson:"merchant_id"`
	DocumentType string             `bson:"document_type"`
	DocumentUrl  string             `bson:"document_url"`
	Status       string             `bson:"status"`
	Note         *string            `bson:"note"`
	UploadedAt   time.Time          `bson:"uploaded_at"`
	EventVersion int64              `bson:"event_version"`
	CreatedAt    *time.Time         `bson:"created_at"`
	UpdatedAt    *time.Time         `bson:"updated_at"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty"`
}

func (MerchantDocument) CollectionName() string { return "merchant_documents" }

// Saldo maps to the "saldos" collection.
type Saldo struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	SaldoID        int32              `bson:"saldo_id"`
	CardNumber     string             `bson:"card_number"`
	TotalBalance   int64              `bson:"total_balance"`
	WithdrawAmount *int64             `bson:"withdraw_amount"`
	WithdrawTime   *time.Time         `bson:"withdraw_time"`
	CreatedAt      *time.Time         `bson:"created_at"`
	UpdatedAt      *time.Time         `bson:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty"`
}

func (Saldo) CollectionName() string { return "saldos" }

// Topup maps to the "topups" collection.
type Topup struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	TopupID        int32              `bson:"topup_id"`
	TopupNo        uuid.UUID          `bson:"topup_no"`
	CardNumber     string             `bson:"card_number"`
	TopupAmount    int64              `bson:"topup_amount"`
	TopupMethod    string             `bson:"topup_method"`
	TopupTime      time.Time          `bson:"topup_time"`
	Status         string             `bson:"status"`
	IdempotencyKey string             `bson:"idempotency_key"`
	CreatedAt      *time.Time         `bson:"created_at"`
	UpdatedAt      *time.Time         `bson:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty"`
	OperationID    string             `bson:"operation_id"`
}

func (Topup) CollectionName() string { return "topups" }

// Transaction maps to the "transactions" collection.
type Transaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	TransactionID   int32              `bson:"transaction_id"`
	TransactionNo   uuid.UUID          `bson:"transaction_no"`
	CardNumber      string             `bson:"card_number"`
	Amount          int64              `bson:"amount"`
	PaymentMethod   string             `bson:"payment_method"`
	MerchantID      int32              `bson:"merchant_id"`
	TransactionTime time.Time          `bson:"transaction_time"`
	Status          string             `bson:"status"`
	FraudScore      int32              `bson:"fraud_score"`
	IdempotencyKey  string             `bson:"idempotency_key"`
	CreatedAt       *time.Time         `bson:"created_at"`
	UpdatedAt       *time.Time         `bson:"updated_at"`
	DeletedAt       *time.Time         `bson:"deleted_at,omitempty"`
	OperationID     string             `bson:"operation_id"`
}

func (Transaction) CollectionName() string { return "transactions" }

// Transfer maps to the "transfers" collection.
type Transfer struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	TransferID     int32              `bson:"transfer_id"`
	TransferNo     uuid.UUID          `bson:"transfer_no"`
	TransferFrom   string             `bson:"transfer_from"`
	TransferTo     string             `bson:"transfer_to"`
	TransferAmount int64              `bson:"transfer_amount"`
	TransferTime   time.Time          `bson:"transfer_time"`
	Status         string             `bson:"status"`
	IdempotencyKey string             `bson:"idempotency_key"`
	CreatedAt      *time.Time         `bson:"created_at"`
	UpdatedAt      *time.Time         `bson:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty"`
	OperationID    string             `bson:"operation_id"`
}

func (Transfer) CollectionName() string { return "transfers" }

// Withdraw maps to the "withdraws" collection.
type Withdraw struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	WithdrawID     int32              `bson:"withdraw_id"`
	WithdrawNo     uuid.UUID          `bson:"withdraw_no"`
	CardNumber     string             `bson:"card_number"`
	WithdrawAmount int64              `bson:"withdraw_amount"`
	WithdrawTime   time.Time          `bson:"withdraw_time"`
	Status         string             `bson:"status"`
	IdempotencyKey string             `bson:"idempotency_key"`
	CreatedAt      *time.Time         `bson:"created_at"`
	UpdatedAt      *time.Time         `bson:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty"`
	OperationID    string             `bson:"operation_id"`
}

func (Withdraw) CollectionName() string { return "withdraws" }

// RefreshToken maps to the "refresh_tokens" collection.
type RefreshToken struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	RefreshTokenID int32              `bson:"refresh_token_id"`
	UserID         int32              `bson:"user_id"`
	Token          string             `bson:"token"`
	Expiration     time.Time          `bson:"expiration"`
	CreatedAt      *time.Time         `bson:"created_at"`
	UpdatedAt      *time.Time         `bson:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty"`
}

func (RefreshToken) CollectionName() string { return "refresh_tokens" }

// ResetToken maps to the "reset_tokens" collection.
type ResetToken struct {
	ID         int32     `bson:"id"`
	UserID     int64     `bson:"user_id"`
	Token      string    `bson:"token"`
	ExpiryDate time.Time `bson:"expiry_date"`
}

func (ResetToken) CollectionName() string { return "reset_tokens" }

// BillingCycle maps to the "billing_cycles" collection.
type BillingCycle struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	BillingID  int32              `bson:"billing_id"`
	CardNumber string             `bson:"card_number"`
	CycleStart time.Time          `bson:"cycle_start"`
	CycleEnd   time.Time          `bson:"cycle_end"`
	AmountDue  int64              `bson:"amount_due"`
	DueDate    time.Time          `bson:"due_date"`
	Status     string             `bson:"status"`
	CreatedAt  *time.Time         `bson:"created_at"`
	UpdatedAt  *time.Time         `bson:"updated_at"`
}

func (BillingCycle) CollectionName() string { return "billing_cycles" }

// CardAuthTransaction maps to the "card_auth_transactions" collection.
type CardAuthTransaction struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	AuthID         int32              `bson:"auth_id"`
	TxnID          string             `bson:"txn_id"`
	CardNumber     string             `bson:"card_number"`
	MerchantID     int32              `bson:"merchant_id"`
	Amount         int64              `bson:"amount"`
	Currency       string             `bson:"currency"`
	Mcc            string             `bson:"mcc"`
	PosEntryMode   string             `bson:"pos_entry_mode"`
	Status         string             `bson:"status"`
	IdempotencyKey string             `bson:"idempotency_key"`
	RiskScore      int32              `bson:"risk_score"`
	CreatedAt      *time.Time         `bson:"created_at"`
	UpdatedAt      *time.Time         `bson:"updated_at"`
}

func (CardAuthTransaction) CollectionName() string { return "card_auth_transactions" }

// CardPayment maps to the "card_payments" collection.
type CardPayment struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	PaymentID      int32              `bson:"payment_id"`
	PaymentUuid    string             `bson:"payment_uuid"`
	CardNumber     string             `bson:"card_number"`
	BillingID      *int32             `bson:"billing_id"`
	Amount         int64              `bson:"amount"`
	PaymentChannel string             `bson:"payment_channel"`
	ReferenceID    string             `bson:"reference_id"`
	Status         string             `bson:"status"`
	CreatedAt      *time.Time         `bson:"created_at"`
	UpdatedAt      *time.Time         `bson:"updated_at"`
}

func (CardPayment) CollectionName() string { return "card_payments" }

// CardReward maps to the "card_rewards" collection.
type CardReward struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	RewardID     int32              `bson:"reward_id"`
	CardNumber   string             `bson:"card_number"`
	TxnID        string             `bson:"txn_id"`
	Amount       int64              `bson:"amount"`
	Mcc          string             `bson:"mcc"`
	PointsEarned int32              `bson:"points_earned"`
	ExpiresAt    time.Time          `bson:"expires_at"`
	Redeemed     bool               `bson:"redeemed"`
	CreatedAt    *time.Time         `bson:"created_at"`
}

func (CardReward) CollectionName() string { return "card_rewards" }

// OutboxEvent maps to the "outbox_events" collection.
type OutboxEvent struct {
	ID            int64      `bson:"id"`
	EventKey      string     `bson:"event_key"`
	Topic         string     `bson:"topic"`
	MessageKey    string     `bson:"message_key"`
	Payload       []byte     `bson:"payload"`
	Status        string     `bson:"status"`
	Attempts      int32      `bson:"attempts"`
	NextAttemptAt time.Time  `bson:"next_attempt_at"`
	LastError     string     `bson:"last_error"`
	ClaimVersion  int64      `bson:"claim_version"`
	CreatedAt     *time.Time `bson:"created_at"`
	UpdatedAt     *time.Time `bson:"updated_at"`
	PublishedAt   *time.Time `bson:"published_at"`
	DeletedAt     *time.Time `bson:"deleted_at,omitempty"`
}

func (OutboxEvent) CollectionName() string { return "outbox_events" }

// ConsumerInbox maps to the "consumer_inbox" collection.
type ConsumerInbox struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	ConsumerName       string             `bson:"consumer_name"`
	EventKey           string             `bson:"event_key"`
	Topic              string             `bson:"topic"`
	PartitionID        int32              `bson:"partition_id"`
	MessageOffset      int64              `bson:"message_offset"`
	Status             string             `bson:"status"`
	Attempts           int32              `bson:"attempts"`
	LeaseUntil         time.Time          `bson:"lease_until"`
	LastError          string             `bson:"last_error"`
	ProcessedAt        *time.Time         `bson:"processed_at"`
	ReservationVersion int64              `bson:"reservation_version"`
	CreatedAt          *time.Time         `bson:"created_at"`
	UpdatedAt          *time.Time         `bson:"updated_at"`
	DeletedAt          *time.Time         `bson:"deleted_at,omitempty"`
}

func (ConsumerInbox) CollectionName() string { return "consumer_inbox" }

// CardEventLog maps to the "card_event_logs" collection.
type CardEventLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	EventID     int64              `bson:"event_id"`
	Topic       string             `bson:"topic"`
	EventType   string             `bson:"event_type"`
	CardNumber  *string            `bson:"card_number"`
	ReferenceID *string            `bson:"reference_id"`
	Payload     []byte             `bson:"payload"`
	ReceivedAt  time.Time          `bson:"received_at"`
}

func (CardEventLog) CollectionName() string { return "card_event_logs" }

// ---------------------------------------------------------------------------
// Query result rows
// ---------------------------------------------------------------------------

// ResetTokenRow is a projection used by the reset_token_row queries.
type ResetTokenRow struct {
	UserID     int64     `bson:"user_id"`
	Token      string    `bson:"token"`
	ExpiryDate time.Time `bson:"expiry_date"`
}

// CardCreateRow is a projection used by the card_create_row queries.
type CardCreateRow struct {
	CardID       int32     `bson:"card_id"`
	UserID       int32     `bson:"user_id"`
	CardNumber   string    `bson:"card_number"`
	CardType     string    `bson:"card_type"`
	ExpireDate   time.Time `bson:"expire_date"`
	Cvv          string    `bson:"cvv"`
	CardProvider string    `bson:"card_provider"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// CardUpdateRow is a projection used by the card_update_row queries.
type CardUpdateRow struct {
	CardID       int32     `bson:"card_id"`
	UserID       int32     `bson:"user_id"`
	CardNumber   string    `bson:"card_number"`
	CardType     string    `bson:"card_type"`
	ExpireDate   time.Time `bson:"expire_date"`
	Cvv          string    `bson:"cvv"`
	CardProvider string    `bson:"card_provider"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// CardTrashRow is a projection used by the card_trash_row queries.
type CardTrashRow struct {
	CardID             int32      `bson:"card_id"`
	UserID             int32      `bson:"user_id"`
	CardNumber         string     `bson:"card_number"`
	CardType           string     `bson:"card_type"`
	ExpireDate         time.Time  `bson:"expire_date"`
	Cvv                string     `bson:"cvv"`
	CardProvider       string     `bson:"card_provider"`
	Status             string     `bson:"status"`
	CreditLimit        int32      `bson:"credit_limit"`
	OutstandingBalance int32      `bson:"outstanding_balance"`
	RewardPoints       int32      `bson:"reward_points"`
	CreatedAt          time.Time  `bson:"created_at"`
	UpdatedAt          time.Time  `bson:"updated_at"`
	DeletedAt          *time.Time `bson:"deleted_at"`
}

// CardRestoreRow is a projection used by the card_restore_row queries.
type CardRestoreRow struct {
	CardID       int32      `bson:"card_id"`
	UserID       int32      `bson:"user_id"`
	CardNumber   string     `bson:"card_number"`
	CardType     string     `bson:"card_type"`
	ExpireDate   time.Time  `bson:"expire_date"`
	Cvv          string     `bson:"cvv"`
	CardProvider string     `bson:"card_provider"`
	CreatedAt    time.Time  `bson:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at"`
}

// CardAllFieldsRow is a projection used by the card_all_fields_row queries.
type CardAllFieldsRow struct {
	CardID             int32     `bson:"card_id"`
	UserID             int32     `bson:"user_id"`
	CardNumber         string    `bson:"card_number"`
	CardType           string    `bson:"card_type"`
	ExpireDate         time.Time `bson:"expire_date"`
	Cvv                string    `bson:"cvv"`
	CardProvider       string    `bson:"card_provider"`
	Status             string    `bson:"status"`
	CreditLimit        int32     `bson:"credit_limit"`
	OutstandingBalance int32     `bson:"outstanding_balance"`
	RewardPoints       int32     `bson:"reward_points"`
	CreatedAt          time.Time `bson:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at"`
}

// CardListRow is a projection used by the card_list_row queries.
type CardListRow struct {
	CardID       int32     `bson:"card_id"`
	UserID       int32     `bson:"user_id"`
	CardNumber   string    `bson:"card_number"`
	CardType     string    `bson:"card_type"`
	ExpireDate   time.Time `bson:"expire_date"`
	Cvv          string    `bson:"cvv"`
	CardProvider string    `bson:"card_provider"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
	TotalCount   int64     `bson:"total_count"`
}

// CardListWithDeletedRow is a projection used by the card_list_with_deleted_row queries.
type CardListWithDeletedRow struct {
	CardID       int32      `bson:"card_id"`
	UserID       int32      `bson:"user_id"`
	CardNumber   string     `bson:"card_number"`
	CardType     string     `bson:"card_type"`
	ExpireDate   time.Time  `bson:"expire_date"`
	Cvv          string     `bson:"cvv"`
	CardProvider string     `bson:"card_provider"`
	CreatedAt    time.Time  `bson:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at"`
	TotalCount   int64      `bson:"total_count"`
}

// CardByEmailRow is a projection used by the card_by_email_row queries.
type CardByEmailRow struct {
	CardID       int32     `bson:"card_id"`
	Email        string    `bson:"email"`
	UserID       int32     `bson:"user_id"`
	CardNumber   string    `bson:"card_number"`
	CardType     string    `bson:"card_type"`
	ExpireDate   time.Time `bson:"expire_date"`
	Cvv          string    `bson:"cvv"`
	CardProvider string    `bson:"card_provider"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// GetAuthTxnByCardNumberRow is a projection used by the get_auth_txn_by_card_number_row queries.
type GetAuthTxnByCardNumberRow struct {
	AuthID         int32     `bson:"auth_id"`
	TxnID          string    `bson:"txn_id"`
	CardNumber     string    `bson:"card_number"`
	MerchantID     int32     `bson:"merchant_id"`
	Amount         int64     `bson:"amount"`
	Currency       string    `bson:"currency"`
	Mcc            string    `bson:"mcc"`
	PosEntryMode   string    `bson:"pos_entry_mode"`
	Status         string    `bson:"status"`
	IdempotencyKey string    `bson:"idempotency_key"`
	RiskScore      int32     `bson:"risk_score"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// CardPaymentListRow is a projection used by the card_payment_list_row queries.
type CardPaymentListRow struct {
	PaymentID      int32     `bson:"payment_id"`
	PaymentUuid    string    `bson:"payment_uuid"`
	CardNumber     string    `bson:"card_number"`
	BillingID      *int32    `bson:"billing_id"`
	Amount         int64     `bson:"amount"`
	PaymentChannel string    `bson:"payment_channel"`
	ReferenceID    string    `bson:"reference_id"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
	TotalCount     int64     `bson:"total_count"`
}

// RedeemableRewardRow is a projection used by the redeemable_reward_row queries.
type RedeemableRewardRow struct {
	RewardID     int32 `bson:"reward_id"`
	PointsEarned int32 `bson:"points_earned"`
}

// MonthlyAmountRow is a projection used by the monthly_amount_row queries.
type MonthlyAmountRow struct {
	Month  string `bson:"month"`
	Amount int32  `bson:"amount"`
}

// YearlyAmountRow is a projection used by the yearly_amount_row queries.
type YearlyAmountRow struct {
	Year   string `bson:"year"`
	Amount int64  `bson:"amount"`
}

// MonthlyBalanceRow is a projection used by the monthly_balance_row queries.
type MonthlyBalanceRow struct {
	Month        string `bson:"month"`
	TotalBalance int32  `bson:"total_balance"`
}

// YearlyBalanceRow is a projection used by the yearly_balance_row queries.
type YearlyBalanceRow struct {
	Year         string `bson:"year"`
	TotalBalance int64  `bson:"total_balance"`
}

// MonthlyTopupRow is a projection used by the monthly_topup_row queries.
type MonthlyTopupRow struct {
	Month            string `bson:"month"`
	TotalTopupAmount int32  `bson:"total_topup_amount"`
}

// YearlyTopupRow is a projection used by the yearly_topup_row queries.
type YearlyTopupRow struct {
	Year             string `bson:"year"`
	TotalTopupAmount int64  `bson:"total_topup_amount"`
}

// MonthlyTransactionRow is a projection used by the monthly_transaction_row queries.
type MonthlyTransactionRow struct {
	Month                  string `bson:"month"`
	TotalTransactionAmount int32  `bson:"total_transaction_amount"`
}

// YearlyTransactionRow is a projection used by the yearly_transaction_row queries.
type YearlyTransactionRow struct {
	Year                   string `bson:"year"`
	TotalTransactionAmount int64  `bson:"total_transaction_amount"`
}

// MonthlyTransferSentRow is a projection used by the monthly_transfer_sent_row queries.
type MonthlyTransferSentRow struct {
	Month           string `bson:"month"`
	TotalSentAmount int32  `bson:"total_sent_amount"`
}

// YearlyTransferSentRow is a projection used by the yearly_transfer_sent_row queries.
type YearlyTransferSentRow struct {
	Year            string `bson:"year"`
	TotalSentAmount int64  `bson:"total_sent_amount"`
}

// MonthlyTransferReceivedRow is a projection used by the monthly_transfer_received_row queries.
type MonthlyTransferReceivedRow struct {
	Month               string `bson:"month"`
	TotalReceivedAmount int32  `bson:"total_received_amount"`
}

// YearlyTransferReceivedRow is a projection used by the yearly_transfer_received_row queries.
type YearlyTransferReceivedRow struct {
	Year                string `bson:"year"`
	TotalReceivedAmount int64  `bson:"total_received_amount"`
}

// MonthlyWithdrawRow is a projection used by the monthly_withdraw_row queries.
type MonthlyWithdrawRow struct {
	Month               string `bson:"month"`
	TotalWithdrawAmount int32  `bson:"total_withdraw_amount"`
}

// YearlyWithdrawRow is a projection used by the yearly_withdraw_row queries.
type YearlyWithdrawRow struct {
	Year                string `bson:"year"`
	TotalWithdrawAmount int64  `bson:"total_withdraw_amount"`
}

// GetUserEmailByCardNumberRow is a projection used by the get_user_email_by_card_number_row queries.
type GetUserEmailByCardNumberRow struct {
	CardID       int32     `bson:"card_id"`
	Email        string    `bson:"email"`
	UserID       int32     `bson:"user_id"`
	CardNumber   string    `bson:"card_number"`
	CardType     string    `bson:"card_type"`
	ExpireDate   time.Time `bson:"expire_date"`
	Cvv          string    `bson:"cvv"`
	CardProvider string    `bson:"card_provider"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// UpdateCardRow is a projection used by the update_card_row queries.
type UpdateCardRow struct {
	CardID       int32     `bson:"card_id"`
	UserID       int32     `bson:"user_id"`
	CardNumber   string    `bson:"card_number"`
	CardType     string    `bson:"card_type"`
	ExpireDate   time.Time `bson:"expire_date"`
	Cvv          string    `bson:"cvv"`
	CardProvider string    `bson:"card_provider"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// AddRewardPointsRow is a projection used by the add_reward_points_row queries.
type AddRewardPointsRow struct {
	CardID             int32     `bson:"card_id"`
	UserID             int32     `bson:"user_id"`
	CardNumber         string    `bson:"card_number"`
	CardType           string    `bson:"card_type"`
	ExpireDate         time.Time `bson:"expire_date"`
	Cvv                string    `bson:"cvv"`
	CardProvider       string    `bson:"card_provider"`
	Status             string    `bson:"status"`
	CreditLimit        int32     `bson:"credit_limit"`
	OutstandingBalance int32     `bson:"outstanding_balance"`
	RewardPoints       int32     `bson:"reward_points"`
}

// UpdateOutstandingBalanceRow is a projection used by the update_outstanding_balance_row queries.
type UpdateOutstandingBalanceRow struct {
	CardID             int32     `bson:"card_id"`
	UserID             int32     `bson:"user_id"`
	CardNumber         string    `bson:"card_number"`
	CardType           string    `bson:"card_type"`
	ExpireDate         time.Time `bson:"expire_date"`
	Cvv                string    `bson:"cvv"`
	CardProvider       string    `bson:"card_provider"`
	Status             string    `bson:"status"`
	CreditLimit        int32     `bson:"credit_limit"`
	OutstandingBalance int32     `bson:"outstanding_balance"`
	RewardPoints       int32     `bson:"reward_points"`
}

// UpdateCardStatusRow is a projection used by the update_card_status_row queries.
type UpdateCardStatusRow struct {
	CardID             int32     `bson:"card_id"`
	UserID             int32     `bson:"user_id"`
	CardNumber         string    `bson:"card_number"`
	CardType           string    `bson:"card_type"`
	ExpireDate         time.Time `bson:"expire_date"`
	Cvv                string    `bson:"cvv"`
	CardProvider       string    `bson:"card_provider"`
	Status             string    `bson:"status"`
	CreditLimit        int32     `bson:"credit_limit"`
	OutstandingBalance int32     `bson:"outstanding_balance"`
	RewardPoints       int32     `bson:"reward_points"`
}

// MerchantListRow is a projection used by the merchant_list_row queries.
type MerchantListRow struct {
	MerchantID int32     `bson:"merchant_id"`
	MerchantNo uuid.UUID `bson:"merchant_no"`
	Name       string    `bson:"name"`
	ApiKey     string    `bson:"api_key"`
	UserID     int32     `bson:"user_id"`
	Status     string    `bson:"status"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
	TotalCount int64     `bson:"total_count"`
}

// MerchantListWithDeletedRow is a projection used by the merchant_list_with_deleted_row queries.
type MerchantListWithDeletedRow struct {
	MerchantID int32      `bson:"merchant_id"`
	MerchantNo uuid.UUID  `bson:"merchant_no"`
	Name       string     `bson:"name"`
	ApiKey     string     `bson:"api_key"`
	UserID     int32      `bson:"user_id"`
	Status     string     `bson:"status"`
	CreatedAt  time.Time  `bson:"created_at"`
	UpdatedAt  time.Time  `bson:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at"`
	TotalCount int64      `bson:"total_count"`
}

// MerchantAllFieldsRow is a projection used by the merchant_all_fields_row queries.
type MerchantAllFieldsRow struct {
	MerchantID int32     `bson:"merchant_id"`
	MerchantNo uuid.UUID `bson:"merchant_no"`
	Name       string    `bson:"name"`
	ApiKey     string    `bson:"api_key"`
	UserID     int32     `bson:"user_id"`
	Status     string    `bson:"status"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
}

// MerchantListByUserRow is a projection used by the merchant_list_by_user_row queries.
type MerchantListByUserRow struct {
	MerchantID int32     `bson:"merchant_id"`
	MerchantNo uuid.UUID `bson:"merchant_no"`
	Name       string    `bson:"name"`
	ApiKey     string    `bson:"api_key"`
	UserID     int32     `bson:"user_id"`
	Status     string    `bson:"status"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
}

// MerchantDocumentListRow is a projection used by the merchant_document_list_row queries.
type MerchantDocumentListRow struct {
	DocumentID   int32     `bson:"document_id"`
	MerchantID   int32     `bson:"merchant_id"`
	DocumentType string    `bson:"document_type"`
	DocumentUrl  string    `bson:"document_url"`
	Status       string    `bson:"status"`
	Note         *string   `bson:"note"`
	UploadedAt   time.Time `bson:"uploaded_at"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
	TotalCount   int64     `bson:"total_count"`
}

// MerchantDocumentListWithDeletedRow is a projection used by the merchant_document_list_with_deleted_row queries.
type MerchantDocumentListWithDeletedRow struct {
	DocumentID   int32      `bson:"document_id"`
	MerchantID   int32      `bson:"merchant_id"`
	DocumentType string     `bson:"document_type"`
	DocumentUrl  string     `bson:"document_url"`
	Status       string     `bson:"status"`
	Note         *string    `bson:"note"`
	UploadedAt   time.Time  `bson:"uploaded_at"`
	CreatedAt    time.Time  `bson:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at"`
	TotalCount   int64      `bson:"total_count"`
}

// MerchantDocumentAllFieldsRow is a projection used by the merchant_document_all_fields_row queries.
type MerchantDocumentAllFieldsRow struct {
	DocumentID   int32     `bson:"document_id"`
	MerchantID   int32     `bson:"merchant_id"`
	DocumentType string    `bson:"document_type"`
	DocumentUrl  string    `bson:"document_url"`
	Status       string    `bson:"status"`
	Note         *string   `bson:"note"`
	UploadedAt   time.Time `bson:"uploaded_at"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// MerchantDocumentCreateRow is a projection used by the merchant_document_create_row queries.
type MerchantDocumentCreateRow struct {
	DocumentID   int32     `bson:"document_id"`
	MerchantID   int32     `bson:"merchant_id"`
	DocumentType string    `bson:"document_type"`
	DocumentUrl  string    `bson:"document_url"`
	Status       string    `bson:"status"`
	Note         *string   `bson:"note"`
	UploadedAt   time.Time `bson:"uploaded_at"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// MerchantDocumentUpdateRow is a projection used by the merchant_document_update_row queries.
type MerchantDocumentUpdateRow struct {
	DocumentID   int32     `bson:"document_id"`
	MerchantID   int32     `bson:"merchant_id"`
	DocumentType string    `bson:"document_type"`
	DocumentUrl  string    `bson:"document_url"`
	Status       string    `bson:"status"`
	Note         *string   `bson:"note"`
	UploadedAt   time.Time `bson:"uploaded_at"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

// MerchantTransactionRow is a projection used by the merchant_transaction_row queries.
type MerchantTransactionRow struct {
	TransactionID   int32      `bson:"transaction_id"`
	CardNumber      string     `bson:"card_number"`
	Amount          int32      `bson:"amount"`
	PaymentMethod   string     `bson:"payment_method"`
	MerchantID      int32      `bson:"merchant_id"`
	MerchantName    string     `bson:"merchant_name"`
	TransactionTime time.Time  `bson:"transaction_time"`
	CreatedAt       time.Time  `bson:"created_at"`
	UpdatedAt       time.Time  `bson:"updated_at"`
	DeletedAt       *time.Time `bson:"deleted_at"`
	TotalCount      int64      `bson:"total_count"`
}

// MerchantApiKeyRow is a projection used by the merchant_api_key_row queries.
type MerchantApiKeyRow struct {
	MerchantID int32  `bson:"merchant_id"`
	Name       string `bson:"name"`
	ApiKey     string `bson:"api_key"`
	UserID     int32  `bson:"user_id"`
	Status     string `bson:"status"`
}

// MerchantMonthlyPaymentMethodRow is a projection used by the merchant_monthly_payment_method_row queries.
type MerchantMonthlyPaymentMethodRow struct {
	Month         string `bson:"month"`
	PaymentMethod string `bson:"payment_method"`
	TotalAmount   int32  `bson:"total_amount"`
}

// MerchantYearlyPaymentMethodRow is a projection used by the merchant_yearly_payment_method_row queries.
type MerchantYearlyPaymentMethodRow struct {
	Year          string `bson:"year"`
	PaymentMethod string `bson:"payment_method"`
	TotalAmount   int64  `bson:"total_amount"`
}

// MerchantMonthlyAmountRow is a projection used by the merchant_monthly_amount_row queries.
type MerchantMonthlyAmountRow struct {
	Month       string `bson:"month"`
	TotalAmount int32  `bson:"total_amount"`
}

// MerchantYearlyAmountRow is a projection used by the merchant_yearly_amount_row queries.
type MerchantYearlyAmountRow struct {
	Year        string `bson:"year"`
	TotalAmount int64  `bson:"total_amount"`
}

// MerchantMonthlyTotalAmountRow is a projection used by the merchant_monthly_total_amount_row queries.
type MerchantMonthlyTotalAmountRow struct {
	Year        string `bson:"year"`
	Month       string `bson:"month"`
	TotalAmount int32  `bson:"total_amount"`
}

// MerchantYearlyTotalAmountRow is a projection used by the merchant_yearly_total_amount_row queries.
type MerchantYearlyTotalAmountRow struct {
	Year        string `bson:"year"`
	TotalAmount int32  `bson:"total_amount"`
}

// RoleRow is a projection used by the role_row queries.
type RoleRow struct {
	RoleID     int32     `bson:"role_id"`
	RoleName   string    `bson:"role_name"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
	TotalCount int64     `bson:"total_count"`
}

// RoleActiveRow is a projection used by the role_active_row queries.
type RoleActiveRow struct {
	RoleID     int32      `bson:"role_id"`
	RoleName   string     `bson:"role_name"`
	CreatedAt  time.Time  `bson:"created_at"`
	UpdatedAt  time.Time  `bson:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at"`
	TotalCount int64      `bson:"total_count"`
}

// RoleTrashedRow is a projection used by the role_trashed_row queries.
type RoleTrashedRow struct {
	RoleID     int32      `bson:"role_id"`
	RoleName   string     `bson:"role_name"`
	CreatedAt  time.Time  `bson:"created_at"`
	UpdatedAt  time.Time  `bson:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at"`
	TotalCount int64      `bson:"total_count"`
}

// SaldoRow is a projection used by the saldo_row queries.
type SaldoRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	TotalCount     int64      `bson:"total_count"`
}

// SaldoActiveRow is a projection used by the saldo_active_row queries.
type SaldoActiveRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	DeletedAt      *time.Time `bson:"deleted_at"`
	TotalCount     int64      `bson:"total_count"`
}

// SaldoTrashedRow is a projection used by the saldo_trashed_row queries.
type SaldoTrashedRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	DeletedAt      *time.Time `bson:"deleted_at"`
	TotalCount     int64      `bson:"total_count"`
}

// SaldoByIDRow is a projection used by the saldo_by_id_row queries.
type SaldoByIDRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
}

// CreateSaldoRow is a projection used by the create_saldo_row queries.
type CreateSaldoRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
}

// UpdateSaldoRow is a projection used by the update_saldo_row queries.
type UpdateSaldoRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
}

// UpdateSaldoBalanceRow is a projection used by the update_saldo_balance_row queries.
type UpdateSaldoBalanceRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
}

// UpdateSaldoWithdrawRow is a projection used by the update_saldo_withdraw_row queries.
type UpdateSaldoWithdrawRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
}

// MonthlySaldoBalanceRow is a projection used by the monthly_saldo_balance_row queries.
type MonthlySaldoBalanceRow struct {
	Month        string `bson:"month"`
	TotalBalance int32  `bson:"total_balance"`
}

// YearlySaldoBalanceRow is a projection used by the yearly_saldo_balance_row queries.
type YearlySaldoBalanceRow struct {
	Year         string `bson:"year"`
	TotalBalance int64  `bson:"total_balance"`
}

// MonthlyTotalSaldoBalanceRow is a projection used by the monthly_total_saldo_balance_row queries.
type MonthlyTotalSaldoBalanceRow struct {
	Year         string `bson:"year"`
	Month        string `bson:"month"`
	TotalBalance int32  `bson:"total_balance"`
}

// YearlyTotalSaldoBalancesRow is a projection used by the yearly_total_saldo_balances_row queries.
type YearlyTotalSaldoBalancesRow struct {
	Year         string `bson:"year"`
	TotalBalance int32  `bson:"total_balance"`
}

// SaldoCardByCardNumberRow is a projection used by the saldo_card_by_card_number_row queries.
type SaldoCardByCardNumberRow struct {
	CardID             int32      `bson:"card_id"`
	UserID             int32      `bson:"user_id"`
	CardNumber         string     `bson:"card_number"`
	CardType           string     `bson:"card_type"`
	ExpireDate         time.Time  `bson:"expire_date"`
	Cvv                string     `bson:"cvv"`
	CardProvider       string     `bson:"card_provider"`
	Status             string     `bson:"status"`
	CreditLimit        int32      `bson:"credit_limit"`
	OutstandingBalance int32      `bson:"outstanding_balance"`
	RewardPoints       int32      `bson:"reward_points"`
	EventVersion       int64      `bson:"event_version"`
	CreatedAt          time.Time  `bson:"created_at"`
	UpdatedAt          time.Time  `bson:"updated_at"`
	DeletedAt          *time.Time `bson:"deleted_at"`
}

// UpdateSaldoBalanceDeltaRow is a projection used by the update_saldo_balance_delta_row queries.
type UpdateSaldoBalanceDeltaRow struct {
	SaldoID        int32      `bson:"saldo_id"`
	CardNumber     string     `bson:"card_number"`
	TotalBalance   int32      `bson:"total_balance"`
	WithdrawAmount *int32     `bson:"withdraw_amount"`
	WithdrawTime   *time.Time `bson:"withdraw_time"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
}

// TopupListRow is a projection used by the topup_list_row queries.
type TopupListRow struct {
	TopupID     int32     `bson:"topup_id"`
	TopupNo     uuid.UUID `bson:"topup_no"`
	CardNumber  string    `bson:"card_number"`
	TopupAmount int32     `bson:"topup_amount"`
	TopupMethod string    `bson:"topup_method"`
	TopupTime   time.Time `bson:"topup_time"`
	Status      string    `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	TotalCount  int64     `bson:"total_count"`
}

// TopupListWithDeletedRow is a projection used by the topup_list_with_deleted_row queries.
type TopupListWithDeletedRow struct {
	TopupID     int32      `bson:"topup_id"`
	TopupNo     uuid.UUID  `bson:"topup_no"`
	CardNumber  string     `bson:"card_number"`
	TopupAmount int32      `bson:"topup_amount"`
	TopupMethod string     `bson:"topup_method"`
	TopupTime   time.Time  `bson:"topup_time"`
	Status      string     `bson:"status"`
	CreatedAt   time.Time  `bson:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at"`
	DeletedAt   *time.Time `bson:"deleted_at"`
	TotalCount  int64      `bson:"total_count"`
}

// TopupAllFieldsRow is a projection used by the topup_all_fields_row queries.
type TopupAllFieldsRow struct {
	TopupID     int32     `bson:"topup_id"`
	TopupNo     uuid.UUID `bson:"topup_no"`
	CardNumber  string    `bson:"card_number"`
	TopupAmount int32     `bson:"topup_amount"`
	TopupMethod string    `bson:"topup_method"`
	TopupTime   time.Time `bson:"topup_time"`
	Status      string    `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

// TopupTrashRestoreRow is a projection used by the topup_trash_restore_row queries.
type TopupTrashRestoreRow struct {
	TopupID     int32      `bson:"topup_id"`
	TopupNo     uuid.UUID  `bson:"topup_no"`
	CardNumber  string     `bson:"card_number"`
	TopupAmount int32      `bson:"topup_amount"`
	TopupMethod string     `bson:"topup_method"`
	TopupTime   time.Time  `bson:"topup_time"`
	Status      string     `bson:"status"`
	CreatedAt   time.Time  `bson:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at"`
	DeletedAt   *time.Time `bson:"deleted_at"`
}

// TopupByIdempotencyKeyRow is a projection used by the topup_by_idempotency_key_row queries.
type TopupByIdempotencyKeyRow struct {
	TopupID     int32     `bson:"topup_id"`
	TopupNo     uuid.UUID `bson:"topup_no"`
	CardNumber  string    `bson:"card_number"`
	TopupAmount int32     `bson:"topup_amount"`
	TopupMethod string    `bson:"topup_method"`
	TopupTime   time.Time `bson:"topup_time"`
	Status      string    `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

// TopupMonthlyAmountRow is a projection used by the topup_monthly_amount_row queries.
type TopupMonthlyAmountRow struct {
	Month       string `bson:"month"`
	TotalAmount int32  `bson:"total_amount"`
}

// TopupYearlyAmountRow is a projection used by the topup_yearly_amount_row queries.
type TopupYearlyAmountRow struct {
	Year        string `bson:"year"`
	TotalAmount int64  `bson:"total_amount"`
}

// TopupMonthlyMethodRow is a projection used by the topup_monthly_method_row queries.
type TopupMonthlyMethodRow struct {
	Month       string `bson:"month"`
	TopupMethod string `bson:"topup_method"`
	TotalTopups int32  `bson:"total_topups"`
	TotalAmount int32  `bson:"total_amount"`
}

// TopupYearlyMethodRow is a projection used by the topup_yearly_method_row queries.
type TopupYearlyMethodRow struct {
	Year        string `bson:"year"`
	TopupMethod string `bson:"topup_method"`
	TotalTopups int64  `bson:"total_topups"`
	TotalAmount int64  `bson:"total_amount"`
}

// TopupMonthlyStatusRow is a projection used by the topup_monthly_status_row queries.
type TopupMonthlyStatusRow struct {
	Year         string `bson:"year"`
	Month        string `bson:"month"`
	TotalSuccess int64  `bson:"total_success"`
	TotalFailed  int64  `bson:"total_failed"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TopupYearlyStatusRow is a projection used by the topup_yearly_status_row queries.
type TopupYearlyStatusRow struct {
	Year         string `bson:"year"`
	TotalSuccess int32  `bson:"total_success"`
	TotalFailed  int32  `bson:"total_failed"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TopupMonthlyMethodByCardRow is a projection used by the topup_monthly_method_by_card_row queries.
type TopupMonthlyMethodByCardRow struct {
	Month       string `bson:"month"`
	TopupMethod string `bson:"topup_method"`
	TotalTopups int32  `bson:"total_topups"`
	TotalAmount int32  `bson:"total_amount"`
}

// TopupYearlyMethodByCardRow is a projection used by the topup_yearly_method_by_card_row queries.
type TopupYearlyMethodByCardRow struct {
	Year        string `bson:"year"`
	TopupMethod string `bson:"topup_method"`
	TotalTopups int64  `bson:"total_topups"`
	TotalAmount int64  `bson:"total_amount"`
}

// TransactionListRow is a projection used by the transaction_list_row queries.
type TransactionListRow struct {
	TransactionID   int32     `bson:"transaction_id"`
	TransactionNo   uuid.UUID `bson:"transaction_no"`
	CardNumber      string    `bson:"card_number"`
	Amount          int32     `bson:"amount"`
	PaymentMethod   string    `bson:"payment_method"`
	MerchantID      int32     `bson:"merchant_id"`
	TransactionTime time.Time `bson:"transaction_time"`
	Status          string    `bson:"status"`
	CreatedAt       time.Time `bson:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at"`
	TotalCount      int64     `bson:"total_count"`
}

// TransactionListWithDeletedRow is a projection used by the transaction_list_with_deleted_row queries.
type TransactionListWithDeletedRow struct {
	TransactionID   int32      `bson:"transaction_id"`
	TransactionNo   uuid.UUID  `bson:"transaction_no"`
	CardNumber      string     `bson:"card_number"`
	Amount          int32      `bson:"amount"`
	PaymentMethod   string     `bson:"payment_method"`
	MerchantID      int32      `bson:"merchant_id"`
	TransactionTime time.Time  `bson:"transaction_time"`
	Status          string     `bson:"status"`
	CreatedAt       time.Time  `bson:"created_at"`
	UpdatedAt       time.Time  `bson:"updated_at"`
	DeletedAt       *time.Time `bson:"deleted_at"`
	TotalCount      int64      `bson:"total_count"`
}

// TransactionAllFieldsRow is a projection used by the transaction_all_fields_row queries.
type TransactionAllFieldsRow struct {
	TransactionID      int32     `bson:"transaction_id"`
	TransactionNo      uuid.UUID `bson:"transaction_no"`
	CardNumber         string    `bson:"card_number"`
	Amount             int32     `bson:"amount"`
	PaymentMethod      string    `bson:"payment_method"`
	MerchantID         int32     `bson:"merchant_id"`
	TransactionTime    time.Time `bson:"transaction_time"`
	Status             string    `bson:"status"`
	IdempotencyKey     string    `bson:"idempotency_key"`
	CardNumber2        string    `bson:"card_number_2"`
	MerchantCardNumber string    `bson:"merchant_card_number"`
	CreatedAt          time.Time `bson:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at"`
}

// TransactionTrashRestoreRow is a projection used by the transaction_trash_restore_row queries.
type TransactionTrashRestoreRow struct {
	TransactionID   int32      `bson:"transaction_id"`
	TransactionNo   uuid.UUID  `bson:"transaction_no"`
	CardNumber      string     `bson:"card_number"`
	Amount          int32      `bson:"amount"`
	PaymentMethod   string     `bson:"payment_method"`
	MerchantID      int32      `bson:"merchant_id"`
	TransactionTime time.Time  `bson:"transaction_time"`
	Status          string     `bson:"status"`
	CreatedAt       time.Time  `bson:"created_at"`
	UpdatedAt       time.Time  `bson:"updated_at"`
	DeletedAt       *time.Time `bson:"deleted_at"`
}

// TransactionByIdempotencyKeyRow is a projection used by the transaction_by_idempotency_key_row queries.
type TransactionByIdempotencyKeyRow struct {
	TransactionID   int32     `bson:"transaction_id"`
	TransactionNo   uuid.UUID `bson:"transaction_no"`
	CardNumber      string    `bson:"card_number"`
	Amount          int32     `bson:"amount"`
	PaymentMethod   string    `bson:"payment_method"`
	MerchantID      int32     `bson:"merchant_id"`
	TransactionTime time.Time `bson:"transaction_time"`
	Status          string    `bson:"status"`
	CreatedAt       time.Time `bson:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at"`
}

// TransactionMonthlyStatusSuccessRow is a projection used by the transaction_monthly_status_success_row queries.
type TransactionMonthlyStatusSuccessRow struct {
	Year         string `bson:"year"`
	Month        string `bson:"month"`
	TotalSuccess int64  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TransactionYearlyStatusSuccessRow is a projection used by the transaction_yearly_status_success_row queries.
type TransactionYearlyStatusSuccessRow struct {
	Year         string `bson:"year"`
	TotalSuccess int32  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TransactionMonthlyStatusFailedRow is a projection used by the transaction_monthly_status_failed_row queries.
type TransactionMonthlyStatusFailedRow struct {
	Year        string `bson:"year"`
	Month       string `bson:"month"`
	TotalFailed int64  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// TransactionYearlyStatusFailedRow is a projection used by the transaction_yearly_status_failed_row queries.
type TransactionYearlyStatusFailedRow struct {
	Year        string `bson:"year"`
	TotalFailed int32  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// TransactionMonthlyPaymentMethodRow is a projection used by the transaction_monthly_payment_method_row queries.
type TransactionMonthlyPaymentMethodRow struct {
	Month         string `bson:"month"`
	PaymentMethod string `bson:"payment_method"`
	TotalCount    int32  `bson:"total_count"`
	TotalAmount   int32  `bson:"total_amount"`
}

// TransactionYearlyPaymentMethodRow is a projection used by the transaction_yearly_payment_method_row queries.
type TransactionYearlyPaymentMethodRow struct {
	Year          string `bson:"year"`
	PaymentMethod string `bson:"payment_method"`
	TotalCount    int64  `bson:"total_count"`
	TotalAmount   int64  `bson:"total_amount"`
}

// TransactionMonthlyAmountRow is a projection used by the transaction_monthly_amount_row queries.
type TransactionMonthlyAmountRow struct {
	Month       string `bson:"month"`
	TotalAmount int32  `bson:"total_amount"`
}

// TransactionYearlyAmountRow is a projection used by the transaction_yearly_amount_row queries.
type TransactionYearlyAmountRow struct {
	Year        string `bson:"year"`
	TotalAmount int64  `bson:"total_amount"`
}

// TransactionMonthlyStatusSuccessByCardRow is a projection used by the transaction_monthly_status_success_by_card_row queries.
type TransactionMonthlyStatusSuccessByCardRow struct {
	Year         string `bson:"year"`
	Month        string `bson:"month"`
	TotalSuccess int64  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TransactionYearlyStatusSuccessByCardRow is a projection used by the transaction_yearly_status_success_by_card_row queries.
type TransactionYearlyStatusSuccessByCardRow struct {
	Year         string `bson:"year"`
	TotalSuccess int32  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TransactionMonthlyStatusFailedByCardRow is a projection used by the transaction_monthly_status_failed_by_card_row queries.
type TransactionMonthlyStatusFailedByCardRow struct {
	Year        string `bson:"year"`
	Month       string `bson:"month"`
	TotalFailed int64  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// TransactionYearlyStatusFailedByCardRow is a projection used by the transaction_yearly_status_failed_by_card_row queries.
type TransactionYearlyStatusFailedByCardRow struct {
	Year        string `bson:"year"`
	TotalFailed int32  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// TransactionMonthlyPaymentMethodByCardRow is a projection used by the transaction_monthly_payment_method_by_card_row queries.
type TransactionMonthlyPaymentMethodByCardRow struct {
	Month         string `bson:"month"`
	PaymentMethod string `bson:"payment_method"`
	TotalCount    int32  `bson:"total_count"`
	TotalAmount   int32  `bson:"total_amount"`
}

// TransactionYearlyPaymentMethodByCardRow is a projection used by the transaction_yearly_payment_method_by_card_row queries.
type TransactionYearlyPaymentMethodByCardRow struct {
	Year          string `bson:"year"`
	PaymentMethod string `bson:"payment_method"`
	TotalCount    int64  `bson:"total_count"`
	TotalAmount   int64  `bson:"total_amount"`
}

// TransactionMonthlyAmountByCardRow is a projection used by the transaction_monthly_amount_by_card_row queries.
type TransactionMonthlyAmountByCardRow struct {
	Month       string `bson:"month"`
	TotalAmount int32  `bson:"total_amount"`
}

// TransactionYearlyAmountByCardRow is a projection used by the transaction_yearly_amount_by_card_row queries.
type TransactionYearlyAmountByCardRow struct {
	Year        string `bson:"year"`
	TotalAmount int64  `bson:"total_amount"`
}

// TransferListRow is a projection used by the transfer_list_row queries.
type TransferListRow struct {
	TransferID     int32     `bson:"transfer_id"`
	TransferNo     uuid.UUID `bson:"transfer_no"`
	TransferFrom   string    `bson:"transfer_from"`
	TransferTo     string    `bson:"transfer_to"`
	TransferAmount int32     `bson:"transfer_amount"`
	TransferTime   time.Time `bson:"transfer_time"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
	TotalCount     int64     `bson:"total_count"`
}

// TransferListWithDeletedRow is a projection used by the transfer_list_with_deleted_row queries.
type TransferListWithDeletedRow struct {
	TransferID     int32      `bson:"transfer_id"`
	TransferNo     uuid.UUID  `bson:"transfer_no"`
	TransferFrom   string     `bson:"transfer_from"`
	TransferTo     string     `bson:"transfer_to"`
	TransferAmount int32      `bson:"transfer_amount"`
	TransferTime   time.Time  `bson:"transfer_time"`
	Status         string     `bson:"status"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	DeletedAt      *time.Time `bson:"deleted_at"`
	TotalCount     int64      `bson:"total_count"`
}

// TransferAllFieldsRow is a projection used by the transfer_all_fields_row queries.
type TransferAllFieldsRow struct {
	TransferID         int32     `bson:"transfer_id"`
	TransferNo         uuid.UUID `bson:"transfer_no"`
	TransferFrom       string    `bson:"transfer_from"`
	TransferTo         string    `bson:"transfer_to"`
	TransferAmount     int32     `bson:"transfer_amount"`
	TransferTime       time.Time `bson:"transfer_time"`
	Status             string    `bson:"status"`
	TransferID2        int32     `bson:"transfer_id_2"`
	MerchantCardNumber string    `bson:"merchant_card_number"`
	CreatedAt          time.Time `bson:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at"`
}

// TransferTrashRestoreRow is a projection used by the transfer_trash_restore_row queries.
type TransferTrashRestoreRow struct {
	TransferID     int32      `bson:"transfer_id"`
	TransferNo     uuid.UUID  `bson:"transfer_no"`
	TransferFrom   string     `bson:"transfer_from"`
	TransferTo     string     `bson:"transfer_to"`
	TransferAmount int32      `bson:"transfer_amount"`
	TransferTime   time.Time  `bson:"transfer_time"`
	Status         string     `bson:"status"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	DeletedAt      *time.Time `bson:"deleted_at"`
}

// TransferByIdempotencyKeyRow is a projection used by the transfer_by_idempotency_key_row queries.
type TransferByIdempotencyKeyRow struct {
	TransferID     int32     `bson:"transfer_id"`
	TransferNo     uuid.UUID `bson:"transfer_no"`
	TransferFrom   string    `bson:"transfer_from"`
	TransferTo     string    `bson:"transfer_to"`
	TransferAmount int32     `bson:"transfer_amount"`
	TransferTime   time.Time `bson:"transfer_time"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// TransferBySourceCardRow is a projection used by the transfer_by_source_card_row queries.
type TransferBySourceCardRow struct {
	TransferID     int32     `bson:"transfer_id"`
	TransferNo     uuid.UUID `bson:"transfer_no"`
	TransferFrom   string    `bson:"transfer_from"`
	TransferTo     string    `bson:"transfer_to"`
	TransferAmount int32     `bson:"transfer_amount"`
	TransferTime   time.Time `bson:"transfer_time"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// TransferMonthlyAmountRow is a projection used by the transfer_monthly_amount_row queries.
type TransferMonthlyAmountRow struct {
	Month               string `bson:"month"`
	TotalTransferAmount int32  `bson:"total_transfer_amount"`
}

// TransferYearlyAmountRow is a projection used by the transfer_yearly_amount_row queries.
type TransferYearlyAmountRow struct {
	Year                string `bson:"year"`
	TotalTransferAmount int64  `bson:"total_transfer_amount"`
}

// TransferMonthlyStatusSuccessRow is a projection used by the transfer_monthly_status_success_row queries.
type TransferMonthlyStatusSuccessRow struct {
	Year         string `bson:"year"`
	Month        string `bson:"month"`
	TotalSuccess int64  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TransferYearlyStatusSuccessRow is a projection used by the transfer_yearly_status_success_row queries.
type TransferYearlyStatusSuccessRow struct {
	Year         string `bson:"year"`
	TotalSuccess int32  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// TransferMonthlyStatusFailedRow is a projection used by the transfer_monthly_status_failed_row queries.
type TransferMonthlyStatusFailedRow struct {
	Year        string `bson:"year"`
	Month       string `bson:"month"`
	TotalFailed int64  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// TransferYearlyStatusFailedRow is a projection used by the transfer_yearly_status_failed_row queries.
type TransferYearlyStatusFailedRow struct {
	Year        string `bson:"year"`
	TotalFailed int32  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// UserRow is a projection used by the user_row queries.
type UserRow struct {
	UserID     int32     `bson:"user_id"`
	Firstname  string    `bson:"firstname"`
	Lastname   string    `bson:"lastname"`
	Email      string    `bson:"email"`
	Password   string    `bson:"password"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
	TotalCount int64     `bson:"total_count"`
}

// UserActiveRow is a projection used by the user_active_row queries.
type UserActiveRow struct {
	UserID     int32      `bson:"user_id"`
	Firstname  string     `bson:"firstname"`
	Lastname   string     `bson:"lastname"`
	Email      string     `bson:"email"`
	Password   string     `bson:"password"`
	CreatedAt  time.Time  `bson:"created_at"`
	UpdatedAt  time.Time  `bson:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at"`
	TotalCount int64      `bson:"total_count"`
}

// UserTrashedRow is a projection used by the user_trashed_row queries.
type UserTrashedRow struct {
	UserID     int32      `bson:"user_id"`
	Firstname  string     `bson:"firstname"`
	Lastname   string     `bson:"lastname"`
	Email      string     `bson:"email"`
	Password   string     `bson:"password"`
	CreatedAt  time.Time  `bson:"created_at"`
	UpdatedAt  time.Time  `bson:"updated_at"`
	DeletedAt  *time.Time `bson:"deleted_at"`
	TotalCount int64      `bson:"total_count"`
}

// UserByIDRow is a projection used by the user_by_id_row queries.
type UserByIDRow struct {
	UserID    int32     `bson:"user_id"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// UserByEmailRow is a projection used by the user_by_email_row queries.
type UserByEmailRow struct {
	UserID int32  `bson:"user_id"`
	Email  string `bson:"email"`
}

// UserByEmailWithPasswordRow is a projection used by the user_by_email_with_password_row queries.
type UserByEmailWithPasswordRow struct {
	UserID   int32  `bson:"user_id"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
}

// UserByEmailAndVerifiedRow is a projection used by the user_by_email_and_verified_row queries.
type UserByEmailAndVerifiedRow struct {
	UserID    int32     `bson:"user_id"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// UserByVerificationCodeRow is a projection used by the user_by_verification_code_row queries.
type UserByVerificationCodeRow struct {
	UserID    int32     `bson:"user_id"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// CreateUserRow is a projection used by the create_user_row queries.
type CreateUserRow struct {
	UserID    int32     `bson:"user_id"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// UpdateUserRow is a projection used by the update_user_row queries.
type UpdateUserRow struct {
	UserID    int32     `bson:"user_id"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// TrashUserRow is a projection used by the trash_user_row queries.
type TrashUserRow struct {
	UserID    int32      `bson:"user_id"`
	Firstname string     `bson:"firstname"`
	Lastname  string     `bson:"lastname"`
	Email     string     `bson:"email"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at"`
}

// RestoreUserRow is a projection used by the restore_user_row queries.
type RestoreUserRow struct {
	UserID    int32      `bson:"user_id"`
	Firstname string     `bson:"firstname"`
	Lastname  string     `bson:"lastname"`
	Email     string     `bson:"email"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at"`
}

// UserIsVerifiedRow is a projection used by the user_is_verified_row queries.
type UserIsVerifiedRow struct {
	UserID    int32     `bson:"user_id"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// UserPasswordRow is a projection used by the user_password_row queries.
type UserPasswordRow struct {
	UserID    int32     `bson:"user_id"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// WithdrawListRow is a projection used by the withdraw_list_row queries.
type WithdrawListRow struct {
	WithdrawID     int32     `bson:"withdraw_id"`
	WithdrawNo     uuid.UUID `bson:"withdraw_no"`
	CardNumber     string    `bson:"card_number"`
	WithdrawAmount int32     `bson:"withdraw_amount"`
	WithdrawTime   time.Time `bson:"withdraw_time"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
	TotalCount     int64     `bson:"total_count"`
}

// WithdrawListWithDeletedRow is a projection used by the withdraw_list_with_deleted_row queries.
type WithdrawListWithDeletedRow struct {
	WithdrawID     int32      `bson:"withdraw_id"`
	WithdrawNo     uuid.UUID  `bson:"withdraw_no"`
	CardNumber     string     `bson:"card_number"`
	WithdrawAmount int32      `bson:"withdraw_amount"`
	WithdrawTime   time.Time  `bson:"withdraw_time"`
	Status         string     `bson:"status"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	DeletedAt      *time.Time `bson:"deleted_at"`
	TotalCount     int64      `bson:"total_count"`
}

// WithdrawAllFieldsRow is a projection used by the withdraw_all_fields_row queries.
type WithdrawAllFieldsRow struct {
	WithdrawID     int32     `bson:"withdraw_id"`
	WithdrawNo     uuid.UUID `bson:"withdraw_no"`
	CardNumber     string    `bson:"card_number"`
	WithdrawAmount int32     `bson:"withdraw_amount"`
	WithdrawTime   time.Time `bson:"withdraw_time"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// WithdrawTrashRestoreRow is a projection used by the withdraw_trash_restore_row queries.
type WithdrawTrashRestoreRow struct {
	WithdrawID     int32      `bson:"withdraw_id"`
	WithdrawNo     uuid.UUID  `bson:"withdraw_no"`
	CardNumber     string     `bson:"card_number"`
	WithdrawAmount int32      `bson:"withdraw_amount"`
	WithdrawTime   time.Time  `bson:"withdraw_time"`
	Status         string     `bson:"status"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	DeletedAt      *time.Time `bson:"deleted_at"`
}

// WithdrawByIdempotencyKeyRow is a projection used by the withdraw_by_idempotency_key_row queries.
type WithdrawByIdempotencyKeyRow struct {
	WithdrawID     int32     `bson:"withdraw_id"`
	WithdrawNo     uuid.UUID `bson:"withdraw_no"`
	CardNumber     string    `bson:"card_number"`
	WithdrawAmount int32     `bson:"withdraw_amount"`
	WithdrawTime   time.Time `bson:"withdraw_time"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// WithdrawByCardNumberRow is a projection used by the withdraw_by_card_number_row queries.
type WithdrawByCardNumberRow struct {
	WithdrawID     int32     `bson:"withdraw_id"`
	WithdrawNo     uuid.UUID `bson:"withdraw_no"`
	CardNumber     string    `bson:"card_number"`
	WithdrawAmount int32     `bson:"withdraw_amount"`
	WithdrawTime   time.Time `bson:"withdraw_time"`
	Status         string    `bson:"status"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
	TotalCount     int64     `bson:"total_count"`
}

// WithdrawMonthlyAmountRow is a projection used by the withdraw_monthly_amount_row queries.
type WithdrawMonthlyAmountRow struct {
	Month               string `bson:"month"`
	TotalWithdrawAmount int32  `bson:"total_withdraw_amount"`
}

// WithdrawYearlyAmountRow is a projection used by the withdraw_yearly_amount_row queries.
type WithdrawYearlyAmountRow struct {
	Year                string `bson:"year"`
	TotalWithdrawAmount int64  `bson:"total_withdraw_amount"`
}

// WithdrawMonthlyStatusSuccessRow is a projection used by the withdraw_monthly_status_success_row queries.
type WithdrawMonthlyStatusSuccessRow struct {
	Year         string `bson:"year"`
	Month        string `bson:"month"`
	TotalSuccess int64  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// WithdrawYearlyStatusSuccessRow is a projection used by the withdraw_yearly_status_success_row queries.
type WithdrawYearlyStatusSuccessRow struct {
	Year         string `bson:"year"`
	TotalSuccess int32  `bson:"total_success"`
	TotalAmount  int32  `bson:"total_amount"`
}

// WithdrawMonthlyStatusFailedRow is a projection used by the withdraw_monthly_status_failed_row queries.
type WithdrawMonthlyStatusFailedRow struct {
	Year        string `bson:"year"`
	Month       string `bson:"month"`
	TotalFailed int64  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// WithdrawYearlyStatusFailedRow is a projection used by the withdraw_yearly_status_failed_row queries.
type WithdrawYearlyStatusFailedRow struct {
	Year        string `bson:"year"`
	TotalFailed int32  `bson:"total_failed"`
	TotalAmount int32  `bson:"total_amount"`
}

// ---------------------------------------------------------------------------
// Aliases kept for compatibility with the per-card stat projections
// ---------------------------------------------------------------------------

type TransferByDestinationCardRow = TransferBySourceCardRow

type TransferMonthlyStatusSuccessByCardRow = TransferMonthlyStatusSuccessRow

type TransferYearlyStatusSuccessByCardRow = TransferYearlyStatusSuccessRow

type TransferMonthlyStatusFailedByCardRow = TransferMonthlyStatusFailedRow

type TransferYearlyStatusFailedByCardRow = TransferYearlyStatusFailedRow

type TransferMonthlyAmountBySenderCardRow = TransferMonthlyAmountRow

type TransferYearlyAmountBySenderCardRow = TransferYearlyAmountRow

type TransferMonthlyAmountByReceiverCardRow = TransferMonthlyAmountRow

type TransferYearlyAmountByReceiverCardRow = TransferYearlyAmountRow

type WithdrawMonthlyStatusSuccessByCardRow = WithdrawMonthlyStatusSuccessRow

type WithdrawYearlyStatusSuccessByCardRow = WithdrawYearlyStatusSuccessRow

type WithdrawMonthlyStatusFailedByCardRow = WithdrawMonthlyStatusFailedRow

type WithdrawYearlyStatusFailedByCardRow = WithdrawYearlyStatusFailedRow

type WithdrawMonthlyAmountByCardRow = WithdrawMonthlyAmountRow

type WithdrawYearlyAmountByCardRow = WithdrawYearlyAmountRow
