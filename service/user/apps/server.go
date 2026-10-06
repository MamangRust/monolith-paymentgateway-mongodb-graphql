package apps

import (
	"fmt"
	"time"

	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/user"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-pkg/server"
	"github.com/MamangRust/monolith-payment-gateway-user/handler"
	"github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/MamangRust/monolith-payment-gateway-user/service"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	// Role service client. User creates assign the default ROLE_ADMIN through the
	// shared role adapter, so the user service no longer reads the roles
	// collection or writes user_roles directly. Fail during bootstrap rather
	// than constructing an adapter around a nil connection.
	connRole, err := grpc.NewClient(
		viper.GetString("GRPC_ROLE_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to role service: %w", err)
	}
	srv.AddCleanupHook(func() error {
		return connRole.Close()
	})

	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, srv.Logger)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(srv.MongoDB,
		pbrole.NewRoleQueryServiceClient(connRole),
		pbuserrole.NewUserRoleServiceClient(connRole),
		repository.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		})
	svc := service.NewService(&service.Deps{
		Cache:        srv.CacheStore,
		Logger:       srv.Logger,
		Repositories: repos,
		Hash:         hash.NewHashingPassword(),
	})
	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterUserQueryServiceServer(gs, h)
		pb.RegisterUserCommandServiceServer(gs, h)
	}

	return srv, nil
}
