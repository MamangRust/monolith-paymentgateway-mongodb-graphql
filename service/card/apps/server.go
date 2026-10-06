package apps

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-card/handler"
	cardkafka "github.com/MamangRust/monolith-payment-gateway-card/kafka"
	"github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-card/service"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/card"
	stats "github.com/MamangRust/monolith-payment-gateway-pb/card/stats"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/outbox"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-pkg/server"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type kafkaOutboxPublisher struct {
	k *kafka.Kafka
}

func (p kafkaOutboxPublisher) SendMessage(ctx context.Context, topic, key string, value []byte) error {
	return p.k.SendMessage(topic, key, value)
}

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}
	myKafka := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})
	srv.AddCleanupHook(myKafka.Close)

	outboxService := outbox.NewOutboxService(srv.MongoDB, kafkaOutboxPublisher{k: myKafka}, srv.Logger)

	connUser, err := grpc.NewClient(
		viper.GetString("GRPC_USER_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to user service: %w", err)
	}
	srv.AddCleanupHook(func() error { return connUser.Close() })

	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(srv.MongoDB, pbuser.NewUserQueryServiceClient(connUser),
		repository.GuardOptions{
			User: []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
		})
	svc := service.NewService(&service.Deps{
		Cache:        srv.CacheStore,
		Logger:       srv.Logger,
		Repositories: repos,
		Kafka:        myKafka,
		UserAdapter:  repos.UserAdapter(),
	})
	h := handler.NewHandler(svc)

	// Start Outbox relay in background
	{
		outboxDone := make(chan struct{})
		go func() {
			defer close(outboxDone)
			outboxService.Start(srv.Ctx, 5*time.Second, 100)
		}()
		srv.AddCleanupHook(func() error {
			select {
			case <-outboxDone:
				return nil
			case <-time.After(5 * time.Second):
				return errors.New("timed out waiting for card outbox relay")
			}
		})
	}

	// Start Fraud Scoring consumer in background
	fraudConsumer := cardkafka.NewFraudScoringConsumer(
		repos.CardAuthTransaction,
		srv.Logger,
	)
	go func() {
		srv.Logger.Info("starting fraud scoring consumer")
		if _, err := myKafka.StartConsumersWithContext(srv.Ctx, []string{"card.txn.created"}, "card-fraud-scoring-group", fraudConsumer); err != nil {
			srv.Logger.Error("fraud scoring consumer error: " + err.Error())
		}
	}()

	// Persist card domain events for a durable, queryable audit trail.
	cardEventLogConsumer := cardkafka.NewCardEventLogHandler(srv.MongoDB, srv.Logger)
	go func() {
		srv.Logger.Info("starting card event log consumer")
		if _, err := myKafka.StartConsumersWithContext(srv.Ctx, []string{
			"card.payment.posted",
			"card.statement.generated",
			"card.limit.changed",
			"card.fraud.alert",
		}, "card-event-log-group", cardEventLogConsumer); err != nil {
			srv.Logger.Error("card event log consumer error: " + err.Error())
		}
	}()

	// Start Billing Scheduler periodic ticker in background
	billingScheduler := cardkafka.NewBillingScheduler(svc, srv.Logger)
	go func() {
		srv.Logger.Info("starting billing scheduler")
		billingScheduler.StartPeriodic(srv.Ctx)
	}()

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterCardQueryServiceServer(gs, h)
		pb.RegisterCardCommandServiceServer(gs, h)
		pb.RegisterCardDashboardServiceServer(gs, h)
		stats.RegisterCardStatsBalanceServiceServer(gs, h)
		stats.RegisterCardStatsTransferServiceServer(gs, h)
		stats.RegisterCardStatsWithdrawServiceServer(gs, h)
		stats.RegisterCardStatsTopupServiceServer(gs, h)
		stats.RegisterCardStatsTransactionServiceServer(gs, h)
	}

	return srv, nil
}
