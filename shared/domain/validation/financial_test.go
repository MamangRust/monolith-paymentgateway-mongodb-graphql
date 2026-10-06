package validation

import (
	"errors"
	"testing"
)

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  int
		wantErr error
	}{
		{"positive amount", 1000, nil},
		{"zero is rejected", 0, ErrInvalidAmount},
		{"negative is rejected", -5, ErrInvalidAmount},
		{"max amount accepted", MaxAmount, nil},
		{"above max is rejected", MaxAmount + 1, ErrAmountTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateAmount(%d) error = %v, want %v", tt.amount, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTransferParties(t *testing.T) {
	if err := ValidateTransferParties("123", "456"); err != nil {
		t.Errorf("different cards should pass, got %v", err)
	}
	if err := ValidateTransferParties("123", "123"); !errors.Is(err, ErrSameCardTransfer) {
		t.Errorf("same cards should be rejected, got %v", err)
	}
}
