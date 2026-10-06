package requests

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
)

// FindAllSaldos represents parameters for searching and paginating saldo records.
// Used to retrieve and filter saldo/balance information with pagination support.
type FindAllSaldos struct {
	Search   string `json:"search" validate:"required"`         // Search term to filter records (matches against card numbers)
	Page     int    `json:"page" validate:"min=1"`              // Page number for pagination (1-based index)
	PageSize int    `json:"page_size" validate:"min=1,max=100"` // Number of items per page (1-100)
}

// MonthTotalSaldoBalance represents parameters for retrieving monthly balance summaries.
// Used to fetch aggregated balance information for specific month/year periods.
type MonthTotalSaldoBalance struct {
	Year  int `json:"year" validate:"required"`  // Year for the balance query (YYYY format)
	Month int `json:"month" validate:"required"` // Month for the balance query (1-12)
}

// CreateSaldoRequest represents the payload for initializing a new saldo record.
// Used when creating a new balance entry for a card.
type CreateSaldoRequest struct {
	CardNumber   string `json:"card_number" validate:"required"`   // Card number associated with the balance
	TotalBalance int    `json:"total_balance" validate:"required"` // Initial balance amount (in smallest currency unit)
}

// UpdateSaldoRequest represents the payload for modifying an existing saldo record.
// Used for comprehensive updates to balance information.
type UpdateSaldoRequest struct {
	SaldoID      *int   `json:"saldo_id"`                          // ID of the saldo record (optional in some flows)
	CardNumber   string `json:"card_number" validate:"required"`   // Card number associated with the balance
	TotalBalance int    `json:"total_balance" validate:"required"` // Updated balance amount (in smallest currency unit)
}

// UpdateSaldoBalance represents the payload for balance adjustment operations.
// Used specifically for updating card balance amounts.
type UpdateSaldoBalance struct {
	CardNumber   string `json:"card_number" validate:"required,min=1"`       // Card number (minimum 1 character)
	TotalBalance int    `json:"total_balance" validate:"required,min=50000"` // New balance (minimum 50,000 in smallest unit)
}

// ApplySaldoAdjustmentRequest represents an append-only admin correction.
type ApplySaldoAdjustmentRequest struct {
	CardNumber  string `json:"card_number" validate:"required,min=1"`
	Delta       int64  `json:"delta" validate:"required"`
	OperationID string `json:"operation_id" validate:"required,min=1,max=128"`
	SourceType  string `json:"source_type" validate:"required,min=1,max=64"`
	SourceID    string `json:"source_id,omitempty" validate:"max=128"`
	Note        string `json:"note" validate:"required,min=1,max=1000"`
}

func (r *ApplySaldoAdjustmentRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}
	if r.Delta == 0 {
		return errors.New("adjustment delta must not be zero")
	}
	return nil
}

// DebitSaldoRequest represents an atomic debit command for a card balance.
type DebitSaldoRequest struct {
	CardNumber  string `json:"card_number" validate:"required,min=1"`
	Amount      int    `json:"amount" validate:"required,gt=0"`
	OperationID string `json:"operation_id,omitempty"`
	SourceType  string `json:"source_type,omitempty"`
	SourceID    string `json:"source_id,omitempty"`
}

// CreditSaldoRequest represents an atomic credit command for a card balance.
type CreditSaldoRequest struct {
	CardNumber  string `json:"card_number" validate:"required,min=1"`
	Amount      int    `json:"amount" validate:"required,gt=0"`
	OperationID string `json:"operation_id,omitempty"`
	SourceType  string `json:"source_type,omitempty"`
	SourceID    string `json:"source_id,omitempty"`
}

// Validate performs validation of DebitSaldoRequest fields.
func (r *DebitSaldoRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs validation of CreditSaldoRequest fields.
func (r *CreditSaldoRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// UpdateSaldoWithdraw represents the payload for withdrawal operations.
// Used when processing balance withdrawals from cards.
type UpdateSaldoWithdraw struct {
	CardNumber     string     `json:"card_number" validate:"required,min=1"`       // Card number (minimum 1 character)
	TotalBalance   int        `json:"total_balance" validate:"required,min=50000"` // Current balance (minimum 50,000)
	WithdrawAmount *int       `json:"withdraw_amount" validate:"omitempty,gte=0"`  // Amount to withdraw (optional, must be ≥ 0)
	WithdrawTime   *time.Time `json:"withdraw_time" validate:"omitempty"`          // Timestamp of withdrawal (optional)
}

// Validate performs basic validation of CreateSaldoRequest fields.
// Ensures required fields are present and properly formatted.
// Returns:
//   - error: if validation fails
func (r *CreateSaldoRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}
	return nil
}

// Validate performs basic validation of UpdateSaldoRequest fields.
// Ensures required fields are present and properly formatted.
// Returns:
//   - error: if validation fails
func (r *UpdateSaldoRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}
	return nil
}

// Validate performs validation of UpdateSaldoBalance fields.
// Ensures balance meets minimum requirements and card number is valid.
// Returns:
//   - error: if validation fails
func (r *UpdateSaldoBalance) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}
	return nil
}

// Validate performs comprehensive validation of UpdateSaldoWithdraw fields.
// Ensures:
// - Withdrawal amount and time are either both provided or both omitted
// - Withdrawal amount doesn't exceed available balance
// - All required fields meet minimum requirements
// Returns:
//   - error: if validation fails with specific validation messages
func (r *UpdateSaldoWithdraw) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	if r.WithdrawAmount != nil && r.WithdrawTime == nil {
		return errors.New("withdraw time must be provided if withdraw amount is provided")
	}
	if r.WithdrawAmount == nil && r.WithdrawTime != nil {
		return errors.New("withdraw amount must be provided if withdraw time is provided")
	}
	if r.WithdrawAmount != nil && *r.WithdrawAmount > r.TotalBalance {
		return errors.New("withdraw amount cannot be greater than total balance")
	}
	return nil
}
