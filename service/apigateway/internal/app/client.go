package apps

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	graph "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/handler"
	"github.com/MamangRust/monolith-payment-gateway-apigateway/internal/middlewares"
	mencache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis"
	"github.com/MamangRust/monolith-payment-gateway-pkg/auth"
	"github.com/MamangRust/monolith-payment-gateway-pkg/dotenv"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	otel_pkg "github.com/MamangRust/monolith-payment-gateway-pkg/otel"
	redisclient "github.com/MamangRust/monolith-payment-gateway-pkg/redis"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceAddresses struct {
	Auth        string
	Role        string
	Card        string
	Merchant    string
	User        string
	Saldo       string
	Topup       string
	Transaction string
	Transfer    string
	Withdraw    string
}

func loadServiceAddresses() *ServiceAddresses {
	return &ServiceAddresses{
		Auth:        getEnvOrDefault("GRPC_AUTH_ADDR", "localhost:50051"),
		Role:        getEnvOrDefault("GRPC_ROLE_ADDR", "localhost:50052"),
		Card:        getEnvOrDefault("GRPC_CARD_ADDR", "localhost:50053"),
		Merchant:    getEnvOrDefault("GRPC_MERCHANT_ADDR", "localhost:50054"),
		User:        getEnvOrDefault("GRPC_USER_ADDR", "localhost:50055"),
		Saldo:       getEnvOrDefault("GRPC_SALDO_ADDR", "localhost:50056"),
		Topup:       getEnvOrDefault("GRPC_TOPUP_ADDR", "localhost:50057"),
		Transaction: getEnvOrDefault("GRPC_TRANSACTION_ADDR", "localhost:50058"),
		Transfer:    getEnvOrDefault("GRPC_TRANSFER_ADDR", "localhost:50059"),
		Withdraw:    getEnvOrDefault("GRPC_WITHDRAW_ADDR", "localhost:50060"),
	}
}

func createServiceConnections(addresses *ServiceAddresses, logger logger.LoggerInterface) (*graph.ServiceConnections, error) {
	var connections graph.ServiceConnections

	conns := map[string]*string{
		"Auth":        &addresses.Auth,
		"Role":        &addresses.Role,
		"Card":        &addresses.Card,
		"Merchant":    &addresses.Merchant,
		"User":        &addresses.User,
		"Saldo":       &addresses.Saldo,
		"Topup":       &addresses.Topup,
		"Transaction": &addresses.Transaction,
		"Transfer":    &addresses.Transfer,
		"Withdraw":    &addresses.Withdraw,
	}

	for name, addr := range conns {
		conn, err := createConnection(*addr, name, logger)
		if err != nil {
			return nil, err
		}

		switch name {
		case "Auth":
			connections.AuthClient = conn
		case "Role":
			connections.RoleClient = conn
		case "Card":
			connections.CardClient = conn
		case "Merchant":
			connections.MerchantClient = conn
		case "User":
			connections.UserClient = conn
		case "Saldo":
			connections.SaldoClient = conn
		case "Topup":
			connections.TopupClient = conn
		case "Transaction":
			connections.TransactionClient = conn
		case "Transfer":
			connections.TransferClient = conn
		case "Withdraw":
			connections.WithdrawClient = conn
		}
	}

	return &connections, nil
}

func createConnection(address, serviceName string, logger logger.LoggerInterface) (*grpc.ClientConn, error) {
	logger.Info(fmt.Sprintf("Connecting to %s service at %s", serviceName, address))

	guard := resilience.NewDependencyGuardInterceptor(logger)

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(guard.UnaryInterceptor()),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to %s service", serviceName), zap.Error(err))
		return nil, err
	}
	return conn, nil
}

func closeConnections(conns *graph.ServiceConnections, log logger.LoggerInterface) {
	for name, conn := range map[string]*grpc.ClientConn{
		"Auth":        conns.AuthClient,
		"Role":        conns.RoleClient,
		"Card":        conns.CardClient,
		"Merchant":    conns.MerchantClient,
		"User":        conns.UserClient,
		"Saldo":       conns.SaldoClient,
		"Topup":       conns.TopupClient,
		"Transaction": conns.TransactionClient,
		"Transfer":    conns.TransferClient,
		"Withdraw":    conns.WithdrawClient,
	} {
		if conn != nil {
			if err := conn.Close(); err != nil {
				log.Error(fmt.Sprintf("Failed to close %s connection", name), zap.Error(err))
			}
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

type Client struct {
	Logger logger.LoggerInterface
}

func RunClient() (*Client, func(), error) {
	flag.Parse()

	addresses := loadServiceAddresses()

	if err := dotenv.Viper(); err != nil {
		fmt.Printf("Warning: Failed to load .env file: %v\n", err)
	}

	ctx := context.Background()

	telemetry := otel_pkg.NewTelemetry(otel_pkg.Config{
		ServiceName:          "apigateway",
		ServiceVersion:       "1.0.0",
		Environment:          "development",
		Endpoint:             getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:             true,
		EnableRuntimeMetrics: true,
	})
	if err := telemetry.Init(ctx); err != nil {
		fmt.Printf("Warning: Failed to initialize telemetry: %v\n", err)
	}

	log, err := logger.NewLogger("apigateway", telemetry.GetLogger())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	log.Debug("Creating gRPC connections...")
	conns, err := createServiceConnections(addresses, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect services: %w", err)
	}

	tokenManager, err := auth.NewManager(viper.GetString("SECRET_KEY"))
	if err != nil {
		log.Fatal("Failed to create token manager", zap.Error(err))
	}

	myKafka := kafka.NewKafka(log, []string{os.Getenv("KAFKA_BROKERS")})

	myredis := redisclient.NewRedisClient(&redisclient.Config{
		Host:         viper.GetString("REDIS_HOST"),
		Port:         viper.GetString("REDIS_PORT"),
		Password:     viper.GetString("REDIS_PASSWORD"),
		DB:           viper.GetInt("REDIS_DB_APIGATEWAY"),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 3,
	})

	if err := myredis.Client.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to ping redis", zap.Error(err))
	}

	mencache := mencache.NewCacheApiGateway(&mencache.Deps{
		Redis:  myredis.Client,
		Logger: log,
	})

	resolver := graph.NewResolver(&graph.Deps{
		Clients:  conns,
		Logger:   log,
		Kafka:    myKafka,
		Mencache: mencache,
	})

	graphqlServer := setupGraphql(tokenManager, resolver, log, myredis.Client)

	go func() {
		log.Info(fmt.Sprintf("🚀 Starting GraphQL server on %s", graphqlServer.Addr))
		if err := graphqlServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("GraphQL server error", zap.Error(err))
		}
	}()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info("Shutting down GraphQL API Gateway...")

		if err := graphqlServer.Shutdown(ctx); err != nil {
			log.Error("GraphQL server shutdown failed", zap.Error(err))
		}

		closeConnections(conns, log)

		if err := telemetry.Shutdown(ctx); err != nil {
			log.Error("Telemetry shutdown failed", zap.Error(err))
		}

		log.Info("Shutdown complete ✅")
	}

	return &Client{
		Logger: log,
	}, shutdown, nil
}

func setupGraphql(token auth.TokenManager, resolver *graph.Resolver, logger logger.LoggerInterface, rdb *redis.Client) *http.Server {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
		Directives: graph.DirectiveRoot{
			HasRole: middlewares.HasRole(resolver.RoleGraphql.Permission),
		},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	port := getEnvOrDefault("CLIENT_PORT", "5000")

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)

	// k8s liveness/readiness probes target these paths; they must stay
	// registered on the GraphQL server port.
	r.Get("/health/live", createLivenessHandler())
	r.Get("/health/ready", createReadinessHandler(logger, rdb))

	r.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(token, logger))
		r.Handle("/query", srv)
	})

	logger.Info("GraphQL Playground running",
		zap.String("url", "http://localhost:"+port),
		zap.String("endpoint", "/query"),
	)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func createLivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	}
}

func createReadinessHandler(logger logger.LoggerInterface, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Error("Readiness check failed: redis unreachable", zap.Error(err))
			respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "not_ready",
				"deps":   map[string]string{"redis": "down"},
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ready",
			"deps":   map[string]string{"redis": "up"},
			"time":   time.Now().UTC(),
		})
	}
}

func respondJSON(w http.ResponseWriter, code int, payload map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
