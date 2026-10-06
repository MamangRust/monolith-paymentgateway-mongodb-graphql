package apps

import (
	"context"
	"fmt"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw"
	pbstats "github.com/MamangRust/monolith-payment-gateway-pb/withdraw/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/outbox"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-pkg/server"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/handler"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/service"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// kafkaOutboxPublisher adapts the payment-gateway *kafka.Kafka (whose
// SendMessage takes no context) to the outbox.OutboxPublisher contract.
type kafkaOutboxPublisher struct {
	k *kafka.Kafka
}

func (p kafkaOutboxPublisher) SendMessage(_ context.Context, topic, key string, value []byte) error {
	return p.k.SendMessage(topic, key, value)
}

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	// gRPC clients for cross-service communication. Fail during bootstrap rather
	// than constructing adapters around a nil connection and panicking on the
	// first request.
	connSaldo, err := grpc.NewClient(
		viper.GetString("GRPC_SALDO_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to saldo service: %w", err)
	}

	connCard, err := grpc.NewClient(
		viper.GetString("GRPC_CARD_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		_ = connSaldo.Close()
		return nil, fmt.Errorf("connect to card service: %w", err)
	}

	srv.AddCleanupHook(func() error {
		return connSaldo.Close()
	})
	srv.AddCleanupHook(func() error {
		return connCard.Close()
	})

	saldoClientQuery := pbsaldo.NewSaldoQueryServiceClient(connSaldo)
	saldoClientCmd := pbsaldo.NewSaldoCommandServiceClient(connSaldo)
	cardClientQuery := pbcard.NewCardQueryServiceClient(connCard)
	cardClientCmd := pbcard.NewCardCommandServiceClient(connCard)

	guardSaldo := resilience.NewDependencyGuard("saldo", 5, 30, 100, 3*time.Second, srv.Logger)
	guardCard := resilience.NewDependencyGuard("card", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(srv.MongoDB,
		cardClientQuery, cardClientCmd,
		saldoClientQuery, saldoClientCmd,
		repository.GuardOptions{
			Card:  []adapter.GuardOption{adapter.WithDependencyGuard(guardCard)},
			Saldo: []adapter.GuardOption{adapter.WithDependencyGuard(guardSaldo)},
		})

	dailyWithdrawLimit := viper.GetInt64("WITHDRAW_DAILY_LIMIT")
	if dailyWithdrawLimit <= 0 {
		dailyWithdrawLimit = 10_000_000
	}
	myKafka := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})
	srv.AddCleanupHook(myKafka.Close)

	outboxService := outbox.NewOutboxService(srv.MongoDB, kafkaOutboxPublisher{k: myKafka}, srv.Logger)

	// Start the outbox relay so events committed with the business writes are
	// published to Kafka with durable retry semantics.
	go outboxService.Start(srv.Ctx, outbox.OutboxRelayInterval, outbox.OutboxRelayBatchSize)

	svc := service.NewService(&service.Deps{
		Kafka:        myKafka,
		Repositories: repos, CardAdapter: repos.CardAdapter(),
		SaldoAdapter:         repos.SaldoAdapter(),
		Logger:               srv.Logger,
		Cache:                srv.CacheStore,
		DailyWithdrawalLimit: dailyWithdrawLimit,
		DB:                   srv.MongoDB,
	})
	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterWithdrawQueryServiceServer(gs, h)
		pb.RegisterWithdrawCommandServiceServer(gs, h)
		pbstats.RegisterWithdrawStatsAmountServiceServer(gs, h)
		pbstats.RegisterWithdrawStatsStatusServiceServer(gs, h)
	}

	return srv, nil
}
