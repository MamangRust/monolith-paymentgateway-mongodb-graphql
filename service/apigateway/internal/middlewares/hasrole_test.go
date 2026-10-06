package middlewares

import (
	"context"
	"errors"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	mycontext "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/context"
)

type fakeRoleChecker struct {
	calls      int
	lastUserID int
	lastRoles  []string
	err        error
}

func (f *fakeRoleChecker) CheckRole(_ context.Context, userID int, requiredRoles ...string) error {
	f.calls++
	f.lastUserID = userID
	f.lastRoles = requiredRoles
	return f.err
}

func passthroughResolver(called *bool) graphql.Resolver {
	return func(ctx context.Context) (any, error) {
		*called = true
		return "resolved", nil
	}
}

func TestHasRole_SkipsWhenNoUserInContext(t *testing.T) {
	checker := &fakeRoleChecker{}
	called := false

	res, err := HasRole(checker)(context.Background(), nil, passthroughResolver(&called), []string{"Admin"})
	if err != nil {
		t.Fatalf("expected public field to pass through, got error %v", err)
	}
	if !called || res != "resolved" {
		t.Fatal("expected the field resolver to run")
	}
	if checker.calls != 0 {
		t.Fatalf("expected checker not to run, ran %d times", checker.calls)
	}
}

func TestHasRole_AllowsWhenCheckerPermits(t *testing.T) {
	checker := &fakeRoleChecker{}
	called := false
	ctx := mycontext.WithUserID(context.Background(), 42)

	res, err := HasRole(checker)(ctx, nil, passthroughResolver(&called), []string{"Admin", "ROLE_ADMIN"})
	if err != nil {
		t.Fatalf("expected the field to be authorised, got error %v", err)
	}
	if !called || res != "resolved" {
		t.Fatal("expected the field resolver to run")
	}
	if checker.calls != 1 {
		t.Fatalf("expected checker to be called once, ran %d times", checker.calls)
	}
	if checker.lastUserID != 42 {
		t.Fatalf("expected user ID 42, got %d", checker.lastUserID)
	}
	if len(checker.lastRoles) != 2 || checker.lastRoles[0] != "Admin" || checker.lastRoles[1] != "ROLE_ADMIN" {
		t.Fatalf("expected roles [Admin ROLE_ADMIN], got %v", checker.lastRoles)
	}
}

func TestHasRole_DeniesWhenCheckerRejects(t *testing.T) {
	checker := &fakeRoleChecker{err: errors.New("role not permitted")}
	called := false
	ctx := mycontext.WithUserID(context.Background(), 99)

	res, err := HasRole(checker)(ctx, nil, passthroughResolver(&called), []string{"Admin"})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
	if called {
		t.Fatal("expected field resolver NOT to run on deny")
	}
	if res != nil {
		t.Fatalf("expected nil result on deny, got %v", res)
	}
	if !errors.Is(err, checker.err) {
		t.Fatalf("expected wrapped checker error, got %v", err)
	}
}

func TestHasRole_WithoutRolesSkipsChecker(t *testing.T) {
	checker := &fakeRoleChecker{}
	called := false
	ctx := mycontext.WithUserID(context.Background(), 1)

	res, err := HasRole(checker)(ctx, nil, passthroughResolver(&called), []string{})
	if err != nil {
		t.Fatalf("expected no error with empty roles, got %v", err)
	}
	if !called || res != "resolved" {
		t.Fatal("expected the field resolver to run")
	}
	if checker.calls != 0 {
		t.Fatalf("expected checker not to run with empty roles, ran %d times", checker.calls)
	}
}

func TestHasRole_ErrorsWhenCheckerMissing(t *testing.T) {
	ctx := mycontext.WithUserID(context.Background(), 1)

	_, err := HasRole(nil)(ctx, nil, passthroughResolver(func() *bool { var b bool; return &b }()), []string{"Admin"})
	if err == nil {
		t.Fatal("expected ErrRoleCheckerNotConfigured")
	}
	if !errors.Is(err, ErrRoleCheckerNotConfigured) {
		t.Fatalf("expected ErrRoleCheckerNotConfigured, got %v", err)
	}
}
