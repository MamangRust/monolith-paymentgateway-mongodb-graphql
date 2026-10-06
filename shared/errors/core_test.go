package errors

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func duplicateKeyError() error {
	return mongo.WriteException{
		WriteErrors: []mongo.WriteError{
			{Code: 11000, Message: "E11000 duplicate key error collection: payment_gateway.merchants index: api_key_unique dup key"},
		},
	}
}

func TestErrConstraintOrFailedMapsDuplicateKeyToConflict(t *testing.T) {
	err := ErrConstraintOrFailed(duplicateKeyError(), "Merchant", "create merchant")

	if err.Code != 409 {
		t.Fatalf("expected HTTP 409, got %d", err.Code)
	}
	if err.Type != ErrorTypeConflict {
		t.Fatalf("expected conflict type, got %s", err.Type)
	}
	if err.Message != "Merchant already exists" {
		t.Fatalf("unexpected public message: %q", err.Message)
	}
	if err.Internal == nil {
		t.Fatal("expected original database error to be retained internally")
	}
	if err.Error() == duplicateKeyError().Error() {
		t.Fatal("database detail leaked into public error")
	}
}

func TestErrNoRowsOrFailedMapsNoDocumentsToNotFound(t *testing.T) {
	err := ErrNoRowsOrFailed(mongo.ErrNoDocuments, "Merchant", "update merchant")

	if err.Code != 404 {
		t.Fatalf("expected HTTP 404, got %d", err.Code)
	}
}

func TestErrConstraintOrFailedKeepsNonConstraintErrorsInternal(t *testing.T) {
	err := ErrConstraintOrFailed(errors.New("database unavailable"), "Merchant", "create merchant")

	if err.Code != 500 {
		t.Fatalf("expected HTTP 500, got %d", err.Code)
	}
	if err.Message == "database unavailable" {
		t.Fatal("database detail leaked into public message")
	}
}

func TestAppErrorJSONDoesNotLeakInternalOrCode(t *testing.T) {
	// The client-visible payload must never carry the internal database error
	// (dup-key detail, E11000 code) nor the raw HTTP code field.
	appErr := ErrConstraintOrFailed(duplicateKeyError(), "Merchant", "create merchant")
	appErr.Internal = errors.New("E11000 duplicate key error collection: payment_gateway.merchants index: api_key_unique dup key")

	b, err := json.Marshal(appErr)
	if err != nil {
		t.Fatalf("marshal app error: %v", err)
	}
	s := string(b)

	if strings.Contains(s, "E11000") || strings.Contains(s, "duplicate key") {
		t.Fatalf("internal database detail leaked into JSON response: %s", s)
	}
	if strings.Contains(s, `"code"`) || strings.Contains(s, `"internal"`) {
		t.Fatalf("internal fields serialized into JSON response: %s", s)
	}
	if !strings.Contains(s, `"message":"Merchant already exists"`) {
		t.Fatalf("public message missing from JSON response: %s", s)
	}
	if !strings.Contains(s, `"type":`) {
		t.Fatalf("error type missing from JSON response: %s", s)
	}
}
