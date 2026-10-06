package events

import "time"

type TransactionEvent struct {
	TransactionID uint64    `json:"transaction_id"`
	TransactionNo string    `json:"transaction_no"`
	CardNumber    string    `json:"card_number"`
	CardType      string    `json:"card_type"`
	CardProvider  string    `json:"card_provider"`
	Amount        int64     `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	MerchantID    uint64    `json:"merchant_id"`
	MerchantName  string    `json:"merchant_name"`
	Status        string    `json:"status"`
	ApiKey        string    `json:"api_key"`
	CreatedAt     time.Time `json:"created_at"`
}

type TopupEvent struct {
	TopupID       uint64    `json:"topup_id"`
	TopupNo       string    `json:"topup_no"`
	CardNumber    string    `json:"card_number"`
	CardType      string    `json:"card_type"`
	CardProvider  string    `json:"card_provider"`
	Amount        int64     `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type TransferEvent struct {
	TransferID      uint64    `json:"transfer_id"`
	TransferNo      string    `json:"transfer_no"`
	SourceCard      string    `json:"source_card"`
	DestinationCard string    `json:"destination_card"`
	Amount          int64     `json:"amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type WithdrawEvent struct {
	WithdrawID uint64    `json:"withdraw_id"`
	WithdrawNo string    `json:"withdraw_no"`
	CardNumber string    `json:"card_number"`
	CardType   string    `json:"card_type"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type SaldoEvent struct {
	CardNumber   string    `json:"card_number"`
	TotalBalance int64     `json:"total_balance"`
	CreatedAt    time.Time `json:"created_at"`
}

type MerchantEvent struct {
	MerchantID uint64    `json:"merchant_id"`
	UserID     uint64    `json:"user_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type CardEvent struct {
	CardID       uint64    `json:"card_id"`
	UserID       uint64    `json:"user_id"`
	CardNumber   string    `json:"card_number"`
	CardType     string    `json:"card_type"`
	CardProvider string    `json:"card_provider"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
