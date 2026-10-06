package state

import "testing"

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		{Pending, Processing, true},
		{Pending, Failed, true},
		{Pending, Success, false},
		{Processing, Success, true},
		{Processing, Compensating, true},
		{Processing, Failed, true},
		{Processing, Unknown, true},
		{Processing, Pending, false},
		{Compensating, Compensated, true},
		{Compensating, Unknown, true},
		{Compensating, Success, false},
		{Unknown, Compensating, true},
		{Unknown, Compensated, true},
		{Unknown, Success, false},
		{Success, Unknown, false},
		{Success, Failed, false},
		{Failed, Success, false},
		{Compensated, Unknown, false},
		{"", Success, false},
	}
	for _, tt := range tests {
		if got := CanTransition(tt.from, tt.to); got != tt.want {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestCheckTransition(t *testing.T) {
	if err := CheckTransition(Pending, Processing); err != nil {
		t.Errorf("expected legal transition, got %v", err)
	}
	if err := CheckTransition(Processing, Pending); err == nil {
		t.Error("expected error for illegal transition, got nil")
	}
}

func TestIsTerminal(t *testing.T) {
	for _, s := range []string{Success, Failed, Compensated} {
		if !IsTerminal(s) {
			t.Errorf("IsTerminal(%q) = false, want true", s)
		}
	}
	for _, s := range []string{Pending, Processing, Compensating, Unknown} {
		if IsTerminal(s) {
			t.Errorf("IsTerminal(%q) = true, want false", s)
		}
	}
}

func TestIsRecoverable(t *testing.T) {
	for _, s := range []string{Processing, Compensating, Unknown} {
		if !IsRecoverable(s) {
			t.Errorf("IsRecoverable(%q) = false, want true", s)
		}
	}
	for _, s := range []string{Pending, Success, Failed, Compensated} {
		if IsRecoverable(s) {
			t.Errorf("IsRecoverable(%q) = true, want false", s)
		}
	}
}

func TestIsValid(t *testing.T) {
	for _, s := range []string{Pending, Processing, Success, Failed, Compensating, Compensated, Unknown} {
		if !IsValid(s) {
			t.Errorf("IsValid(%q) = false, want true", s)
		}
	}
	if IsValid("bogus") {
		t.Error("IsValid(bogus) = true, want false")
	}
}
