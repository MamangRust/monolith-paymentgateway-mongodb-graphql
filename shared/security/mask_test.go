package security

import "testing"

func TestMaskCardNumber(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"full PAN", "4532123456789012", "453212******9012"},
		{"empty", "", ""},
		{"short value is fully masked", "1234", "****"},
		{"exactly 10 chars is fully masked", "4532123456", "**********"},
		{"whitespace trimmed", " 4532123456789012 ", "453212******9012"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskCardNumber(tt.in); got != tt.want {
				t.Errorf("MaskCardNumber(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"full key", "sk_live_9f2c1d8e", "************1d8e"},
		{"empty", "", ""},
		{"short key fully masked", "abcd", "****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskAPIKey(tt.in); got != tt.want {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"standard", "john.doe@example.com", "j******e@example.com"},
		{"no domain", "notanemail", "**********"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskEmail(tt.in); got != tt.want {
				t.Errorf("MaskEmail(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
