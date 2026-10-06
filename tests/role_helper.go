package tests

import (
	"context"
	"time"

	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	role_handler "github.com/MamangRust/monolith-payment-gateway-role/handler"
	role_repo "github.com/MamangRust/monolith-payment-gateway-role/repository"
	role_service "github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DefaultRoleName is the role user-service assigns to every newly created user.
const DefaultRoleName = "ROLE_ADMIN"

// RoleConn starts the role service in-process against the test MongoDB/Redis and
// returns a gRPC connection to it. The connection and server are torn down by
// TestSuite.Teardown.
func (ts *TestSuite) RoleConn(log logger.LoggerInterface) (*grpc.ClientConn, error) {
	if ts.roleConn != nil {
		return ts.roleConn, nil
	}

	redisClient := ts.RedisClient()
	metrics, err := observability.NewCacheMetrics("test-role")
	if err != nil {
		return nil, err
	}
	cacheStore := cache.NewCacheStore(redisClient, log, metrics)

	server := grpc.NewServer()
	registerRoleService(server, ts.db, cacheStore, log)

	addr, err := RunGRPCServer(server)
	if err != nil {
		server.Stop()
		return nil, err
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		server.Stop()
		return nil, err
	}

	ts.roleServer = server
	ts.roleConn = conn
	return conn, nil
}

// registerRoleService wires the role query/command gRPC services backed by db
// onto an existing server, so callers that host several services on one
// listener (dependency harness, test suites) can share it.
func registerRoleService(server *grpc.Server, db *mongo.Database, cacheStore *cache.CacheStore, log logger.LoggerInterface) {
	svc := role_service.NewService(&role_service.Deps{
		Cache:        cacheStore,
		Repositories: role_repo.NewRepositories(db),
		Logger:       log,
	})
	h := role_handler.NewHandler(svc)
	pbrole.RegisterRoleQueryServiceServer(server, h.RoleQuery)
	pbrole.RegisterRoleCommandServiceServer(server, h.RoleCommand)
	pbuserrole.RegisterUserRoleServiceServer(server, h.UserRole)
}

// UserRoleAdapters starts the in-process role service, seeds the default
// ROLE_ADMIN role and returns the raw role query and user_role gRPC clients.
// User repositories build the guarded role/user_role adapters from these
// clients internally. Any suite that builds a user service in-process must use
// this so CreateUser can resolve and assign the default role.
func (ts *TestSuite) UserRoleAdapters(log logger.LoggerInterface) (pbrole.RoleQueryServiceClient, pbuserrole.UserRoleServiceClient, error) {
	conn, err := ts.RoleConn(log)
	if err != nil {
		return nil, nil, err
	}

	if _, err := ts.SeedRole(ts.Ctx, DefaultRoleName); err != nil {
		return nil, nil, err
	}

	return pbrole.NewRoleQueryServiceClient(conn), pbuserrole.NewUserRoleServiceClient(conn), nil
}

// SeedRole upserts a role by name so user-service CreateUser can look up the
// default ROLE_ADMIN. Returns the role id.
func (ts *TestSuite) SeedRole(ctx context.Context, name string) (int, error) {
	return seedRole(ctx, ts.db, name)
}

// seedRole upserts a role by name and returns its id.
func seedRole(ctx context.Context, db *mongo.Database, name string) (int, error) {
	roleID, err := salemongo.NextID(ctx, db, "roles")
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	_, err = db.Collection("roles").UpdateOne(ctx,
		bson.M{"role_name": name},
		bson.M{"$setOnInsert": salemongo.Role{
			RoleID:    roleID,
			RoleName:  name,
			CreatedAt: &now,
			UpdatedAt: &now,
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return 0, err
	}
	return int(roleID), nil
}
