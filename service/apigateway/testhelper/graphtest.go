package testhelper

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	graph "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/handler"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/middlewares"
	mencache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis"
	"github.com/MamangRust/monolith-payment-gateway-pkg/auth"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/vektah/gqlparser/v2/ast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceConnections mirrors graph.ServiceConnections for test use.
type ServiceConnections = graph.ServiceConnections

// CreateDummyConn creates a lazy gRPC connection that will never actually connect.
func CreateDummyConn() *grpc.ClientConn {
	conn, err := grpc.NewClient("localhost:1", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}
	return conn
}

// NewResolver creates a Resolver with the provided service connections.
func NewResolver(conns *ServiceConnections, log logger.LoggerInterface) *graph.Resolver {
	myMencache := mencache.NewCacheApiGateway(&mencache.Deps{
		Redis:  nil,
		Logger: log,
	})

	return graph.NewResolver(&graph.Deps{
		Clients:  conns,
		Logger:   log,
		Kafka:    nil,
		Mencache: myMencache,
	})
}

// NewResolverWithRedis creates a Resolver with the provided service connections and Redis client.
func NewResolverWithRedis(conns *ServiceConnections, log logger.LoggerInterface, redisClient *redis.Client) *graph.Resolver {
	myMencache := mencache.NewCacheApiGateway(&mencache.Deps{
		Redis:  redisClient,
		Logger: log,
	})

	return graph.NewResolver(&graph.Deps{
		Clients:  conns,
		Logger:   log,
		Kafka:    nil,
		Mencache: myMencache,
	})
}

// NewTestHandler wires a full GraphQL HTTP handler from a map of in-process
// gRPC connections (keyed by service name) plus a Redis client, mirroring how
// external integration tests boot the API gateway. Services that are not
// present in the map simply get a lazy client that is never used.
func NewTestHandler(conns map[string]*grpc.ClientConn, log logger.LoggerInterface, redisClient *redis.Client) http.Handler {
	get := func(key string) *grpc.ClientConn {
		if c, ok := conns[key]; ok && c != nil {
			return c
		}
		return nil
	}

	sc := &graph.ServiceConnections{
		AuthClient:        get("auth"),
		RoleClient:        get("role"),
		UserClient:        get("user"),
		CardClient:        get("card"),
		MerchantClient:    get("merchant"),
		SaldoClient:       get("saldo"),
		TopupClient:       get("topup"),
		TransactionClient: get("transaction"),
		TransferClient:    get("transfer"),
		WithdrawClient:    get("withdraw"),
	}

	return NewGraphQLHTTPHandler(NewResolverWithRedis(sc, log, redisClient))
}

// NewAuthenticatedTestHandler behaves like NewTestHandler but wraps the
// GraphQL handler with the apigateway production auth middleware, so protected
// operations (e.g. getMe) resolve the caller from the Authorization: Bearer
// header. Public operations (registerUser/loginUser/refreshToken) still pass
// through without a token.
func NewAuthenticatedTestHandler(conns map[string]*grpc.ClientConn, log logger.LoggerInterface, redisClient *redis.Client, tm auth.TokenManager) http.Handler {
	handler := NewTestHandler(conns, log, redisClient)
	return middlewares.AuthMiddleware(tm, log)(handler)
}

// NewGraphQLHTTPHandler creates an http.Handler from a gqlgen Resolver.
func NewGraphQLHTTPHandler(resolver *graph.Resolver) http.Handler {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
	}))

	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})

	return srv
}

// SeedMerchantCache writes a merchant ID-to-API-key mapping into Redis
// so the permission validation can find it without Kafka.
func SeedMerchantCache(redisClient *redis.Client, merchantID string, apiKey string) error {
	key := "merchant_api_key:" + merchantID
	return redisClient.Set(context.Background(), key, apiKey, 0).Err()
}

// SeedRoleCache writes a user role mapping into Redis
// so the permission validation can find it without Kafka.
func SeedRoleCache(redisClient *redis.Client, userID string, roles []string) error {
	key := "user_roles:" + userID
	return redisClient.Set(context.Background(), key, roles, 0).Err()
}
