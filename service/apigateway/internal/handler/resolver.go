package graph

import (
	errorstd "errors"
	"time"

	"fmt"

	authgraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/auth"
	cardgraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/card"
	merchantgraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/merchant"
	merchantdocumentgraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/merchant_document"
	rolegraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/role"
	saldographqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/saldo"
	topupgraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/topup"
	transactiongraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/transaction"
	transfergraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/transfer"
	usergraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/user"
	withdrawgraphqlmapper "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/mapper/withdraw"
	merchantpermission "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/permission/merchant"
	rolepermission "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/permission/role"
	mencache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis"
	auth_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/auth"
	card_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/card"
	merchant_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/merchant"
	merchant_document_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/merchant_document"
	role_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/role"
	saldo_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/saldo"
	topup_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/topup"
	transaction_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/transaction"
	transfer_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/transfer"
	user_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/user"
	withdraw_cache "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/redis/api/withdraw"
	authpb "github.com/MamangRust/monolith-payment-gateway-pb"
	cardpb "github.com/MamangRust/monolith-payment-gateway-pb/card"
	cardstatpb "github.com/MamangRust/monolith-payment-gateway-pb/card/stats"
	merchantpb "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	merchantstatpb "github.com/MamangRust/monolith-payment-gateway-pb/merchant/stats"
	merchantdocumentpb "github.com/MamangRust/monolith-payment-gateway-pb/merchant_document"
	rolepb "github.com/MamangRust/monolith-payment-gateway-pb/role"
	saldopb "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	saldostatspb "github.com/MamangRust/monolith-payment-gateway-pb/saldo/stats"
	topuppb "github.com/MamangRust/monolith-payment-gateway-pb/topup"
	topupstatpb "github.com/MamangRust/monolith-payment-gateway-pb/topup/stats"
	transactionpb "github.com/MamangRust/monolith-payment-gateway-pb/transaction"
	transactionstatpb "github.com/MamangRust/monolith-payment-gateway-pb/transaction/stats"
	transferpb "github.com/MamangRust/monolith-payment-gateway-pb/transfer"
	transferstatpb "github.com/MamangRust/monolith-payment-gateway-pb/transfer/stats"
	userpb "github.com/MamangRust/monolith-payment-gateway-pb/user"
	withdrawpb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw"
	withdrawstatpb "github.com/MamangRust/monolith-payment-gateway-pb/withdraw/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/kafka"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/errors"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthGraphql             AuthHandleGraphql
	RoleGraphql             RoleHandleGraphql
	UserGraphql             UserHandleGraphql
	CardGraphql             CardHandleGraphql
	MerchantGraphql         MerchantHandleGraphql
	MerchantDocumentGraphql MerchantDocumentHandleGraphql
	SaldoGraphql            SaldoHandleGraphql
	TopupGraphql            TopupHandleGraphql
	TransactionGraphql      TransactionHandleGraphql
	TransferGraphql         TransferHandleGraphql
	WithdrawGraphql         WithdrawHandleGraphql
	ResolverHandle          *resolverHandler
}

type UserClient struct {
	UserQueryClient   userpb.UserQueryServiceClient
	UserCommandClient userpb.UserCommandServiceClient
}

type RoleClient struct {
	RoleQueryClient   rolepb.RoleQueryServiceClient
	RoleCommandClient rolepb.RoleCommandServiceClient
}

type CardClient struct {
	CardQueryClient                  cardpb.CardQueryServiceClient
	CardCommandClient                cardpb.CardCommandServiceClient
	CardDashboardClient              cardpb.CardDashboardServiceClient
	CardStatsBalanceClient           cardstatpb.CardStatsBalanceServiceClient
	CardStatsTopupAmountClient       cardstatpb.CardStatsTopupServiceClient
	CardStatsTransactionAmountClient cardstatpb.CardStatsTransactionServiceClient
	CardStatsWithdrawAmountClient    cardstatpb.CardStatsWithdrawServiceClient
	CardStatsTransferAmountClient    cardstatpb.
						CardStatsTransferServiceClient
}

type MerchantClient struct {
	MerchantQuery            merchantpb.MerchantQueryServiceClient
	MerchantCommand          merchantpb.MerchantCommandServiceClient
	MerchantTransaction      merchantpb.MerchantTransactionServiceClient
	MerchantStatsAmount      merchantstatpb.MerchantStatsAmountServiceClient
	MerchantStatsTotalAmount merchantstatpb.MerchantStatsTotalAmountServiceClient
	MerchantStatsMethod      merchantstatpb.MerchantStatsMethodServiceClient
}

type MerchantDocumentClient struct {
	MerchantDocumentQueryClient   merchantdocumentpb.MerchantDocumentQueryServiceClient
	MerchantDocumentCommandClient merchantdocumentpb.MerchantDocumentCommandServiceClient
}

type SaldoClient struct {
	SaldoQueryClient             saldopb.SaldoQueryServiceClient
	SaldoCommandClient           saldopb.SaldoCommandServiceClient
	SaldoStatsBalanceClient      saldostatspb.SaldoStatsBalanceServiceClient
	SaldoStatsTotalBalanceClient saldostatspb.SaldoStatsTotalBalanceClient
}

type TopupClient struct {
	TopupQueryClient   topuppb.TopupQueryServiceClient
	TopupCommandClient topuppb.TopupCommandServiceClient
	TopupStatsAmount   topupstatpb.TopupStatsAmountServiceClient
	TopupStatsMethod   topupstatpb.TopupStatsMethodServiceClient
	TopupStatsStatus   topupstatpb.TopupStatsStatusServiceClient
}

type TransactionClient struct {
	TransactionQueryClient   transactionpb.TransactionQueryServiceClient
	TransactionCommandClient transactionpb.TransactionCommandServiceClient
	TransactionStatsAmount   transactionstatpb.TransactionStatsAmountServiceClient
	TransactionStatsMethod   transactionstatpb.TransactionStatsMethodServiceClient
	TransactionStatsStatus   transactionstatpb.TransactionStatsStatusServiceClient
}

type TransferClient struct {
	TransferQueryClient   transferpb.TransferQueryServiceClient
	TransferCommandClient transferpb.TransferCommandServiceClient
	TransferStatsAmount   transferstatpb.TransferStatsAmountServiceClient
	TransferStatsStatus   transferstatpb.TransferStatsStatusServiceClient
}
type WithdrawClient struct {
	WithdrawQueryClient   withdrawpb.WithdrawQueryServiceClient
	WithdrawCommandClient withdrawpb.WithdrawCommandServiceClient
	WithdrawStatsAmount   withdrawstatpb.WithdrawStatsAmountServiceClient
	WithdrawStatsStatus   withdrawstatpb.WithdrawStatsStatusServiceClient
}

type AuthHandleGraphql struct {
	AuthClient authpb.AuthServiceClient
	Logger     logger.LoggerInterface
	Mapping    authgraphqlmapper.AuthGraphqlMapper
	Cache      auth_cache.AuthMencache
}

type RoleHandleGraphql struct {
	RoleClient RoleClient
	Logger     logger.LoggerInterface
	Mapping    rolegraphqlmapper.RoleGraphqlMapper
	Kafka      *kafka.Kafka
	CacheRole  mencache.RoleCache
	Permission rolepermission.RolePermission
	Cache      role_cache.RoleMencache
}

type UserHandleGraphql struct {
	UserClient UserClient
	Logger     logger.LoggerInterface
	Mapping    usergraphqlmapper.UserGraphqlMapper
	Cache      user_cache.UserMencache
}

type CardHandleGraphql struct {
	CardClient CardClient
	Logger     logger.LoggerInterface
	Mapping    cardgraphqlmapper.CardGraphqlMapper
	Cache      card_cache.CardMencache
}

type MerchantHandleGraphql struct {
	MerchantClient MerchantClient
	Logger         logger.LoggerInterface
	Mapping        merchantgraphqlmapper.MerchantGraphqlMapper
	Cache          merchant_cache.MerchantMencache
}

type MerchantDocumentHandleGraphql struct {
	MerchantClient MerchantDocumentClient
	Logger         logger.LoggerInterface
	Mapping        merchantdocumentgraphqlmapper.MerchantDocumentGraphqlMapper
	Cache          merchant_document_cache.MerchantDocumentMencache
}

type SaldoHandleGraphql struct {
	SaldoClient SaldoClient
	Logger      logger.LoggerInterface
	Mapping     saldographqlmapper.SaldoGraphqlMapper
	Cache       saldo_cache.SaldoMencache
}

type TopupHandleGraphql struct {
	TopupClient TopupClient
	Logger      logger.LoggerInterface
	Mapping     topupgraphqlmapper.TopupGraphqlMapper
	Cache       topup_cache.TopupMencach
}

type TransactionHandleGraphql struct {
	TransactionClient TransactionClient
	Logger            logger.LoggerInterface
	Mapping           transactiongraphqlmapper.TransactionGraphqlMapper
	Permission        merchantpermission.MerchantPermission
	CacheMerchant     mencache.MerchantCache
	Cache             transaction_cache.TransactionMencache
}

type TransferHandleGraphql struct {
	TransferClient TransferClient
	Logger         logger.LoggerInterface
	Mapping        transfergraphqlmapper.TransferGraphqlMapper
	Cache          transfer_cache.TransferMencache
}

type WithdrawHandleGraphql struct {
	WithdrawClient WithdrawClient
	Logger         logger.LoggerInterface
	Mapping        withdrawgraphqlmapper.WithdrawGraphqlMapper
	Cache          withdraw_cache.WithdrawMencache
}

type ServiceConnections struct {
	AuthClient        *grpc.ClientConn
	RoleClient        *grpc.ClientConn
	UserClient        *grpc.ClientConn
	CardClient        *grpc.ClientConn
	MerchantClient    *grpc.ClientConn
	SaldoClient       *grpc.ClientConn
	TopupClient       *grpc.ClientConn
	TransactionClient *grpc.ClientConn
	TransferClient    *grpc.ClientConn
	WithdrawClient    *grpc.ClientConn
}

type Deps struct {
	Clients  *ServiceConnections
	Logger   logger.LoggerInterface
	Kafka    *kafka.Kafka
	Mencache mencache.CacheApiGateway
}

func NewResolver(
	deps *Deps,
) *Resolver {
	observability, _ := observability.NewObservability(
		"graphql-client",
		deps.Logger,
	)

	resolverHandle := NewResolverHandler(observability, deps.Logger)

	store := deps.Mencache.GetStore()
	cacheAuth := auth_cache.NewMencache(store)
	cacheUser := user_cache.NewUserMencache(store)
	cacheRole := role_cache.NewRoleMencache(store)
	cacheSaldo := saldo_cache.NewSaldoMencache(store)
	cacheTopup := topup_cache.NewTopupMencache(store)
	cacheTransaction := transaction_cache.NewTransactionMencache(store)
	cacheTransfer := transfer_cache.NewTransferMencache(store)
	cacheWithdraw := withdraw_cache.NewWithdrawMencache(store)
	cacheMerchant := merchant_cache.NewMerchantMencache(store)
	cacheCard := card_cache.NewCardMencache(store)
	cacheMerchantDocument := merchant_document_cache.NewMerchantDocumentMencache(store)

	return &Resolver{
		ResolverHandle: resolverHandle,
		AuthGraphql: AuthHandleGraphql{
			AuthClient: authpb.NewAuthServiceClient(deps.Clients.AuthClient),
			Logger:     deps.Logger,
			Mapping:    authgraphqlmapper.NewAuthGraphqlMapper(),
			Cache:      cacheAuth,
		},
		RoleGraphql: RoleHandleGraphql{
			RoleClient: RoleClient{
				RoleQueryClient:   rolepb.NewRoleQueryServiceClient(deps.Clients.RoleClient),
				RoleCommandClient: rolepb.NewRoleCommandServiceClient(deps.Clients.RoleClient),
			},
			Kafka:      deps.Kafka,
			Logger:     deps.Logger,
			Mapping:    rolegraphqlmapper.NewRoleGraphqlMapper(),
			Permission: rolepermission.NewRolePermission(deps.Kafka, "request-role", "response-role", 5*time.Second, deps.Logger, deps.Mencache),
			Cache:      cacheRole,
		},
		UserGraphql: UserHandleGraphql{
			UserClient: UserClient{
				UserQueryClient:   userpb.NewUserQueryServiceClient(deps.Clients.UserClient),
				UserCommandClient: userpb.NewUserCommandServiceClient(deps.Clients.UserClient),
			},
			Logger:  deps.Logger,
			Mapping: usergraphqlmapper.NewUserGraphqlMapper(),
			Cache:   cacheUser,
		},
		CardGraphql: CardHandleGraphql{
			CardClient: CardClient{
				CardQueryClient:                  cardpb.NewCardQueryServiceClient(deps.Clients.CardClient),
				CardCommandClient:                cardpb.NewCardCommandServiceClient(deps.Clients.CardClient),
				CardDashboardClient:              cardpb.NewCardDashboardServiceClient(deps.Clients.CardClient),
				CardStatsBalanceClient:           cardstatpb.NewCardStatsBalanceServiceClient(deps.Clients.CardClient),
				CardStatsTopupAmountClient:       cardstatpb.NewCardStatsTopupServiceClient(deps.Clients.CardClient),
				CardStatsTransactionAmountClient: cardstatpb.NewCardStatsTransactionServiceClient(deps.Clients.CardClient),
				CardStatsWithdrawAmountClient:    cardstatpb.NewCardStatsWithdrawServiceClient(deps.Clients.CardClient),
				CardStatsTransferAmountClient:    cardstatpb.NewCardStatsTransferServiceClient(deps.Clients.CardClient),
			},
			Logger:  deps.Logger,
			Mapping: cardgraphqlmapper.NewCardResponseMapper(),
			Cache:   cacheCard,
		},
		MerchantGraphql: MerchantHandleGraphql{
			MerchantClient: MerchantClient{
				MerchantQuery:            merchantpb.NewMerchantQueryServiceClient(deps.Clients.MerchantClient),
				MerchantCommand:          merchantpb.NewMerchantCommandServiceClient(deps.Clients.MerchantClient),
				MerchantTransaction:      merchantpb.NewMerchantTransactionServiceClient(deps.Clients.MerchantClient),
				MerchantStatsAmount:      merchantstatpb.NewMerchantStatsAmountServiceClient(deps.Clients.MerchantClient),
				MerchantStatsTotalAmount: merchantstatpb.NewMerchantStatsTotalAmountServiceClient(deps.Clients.MerchantClient),
				MerchantStatsMethod:      merchantstatpb.NewMerchantStatsMethodServiceClient(deps.Clients.MerchantClient),
			},
			Logger:  deps.Logger,
			Mapping: merchantgraphqlmapper.NewMerchantResponseMapper(),
			Cache:   cacheMerchant,
		},
		MerchantDocumentGraphql: MerchantDocumentHandleGraphql{
			MerchantClient: MerchantDocumentClient{
				MerchantDocumentQueryClient:   merchantdocumentpb.NewMerchantDocumentQueryServiceClient(deps.Clients.MerchantClient),
				MerchantDocumentCommandClient: merchantdocumentpb.NewMerchantDocumentCommandServiceClient(deps.Clients.MerchantClient),
			},
			Logger:  deps.Logger,
			Mapping: merchantdocumentgraphqlmapper.NewMerchantDocumentGraphqlMapper(),
			Cache:   cacheMerchantDocument,
		},
		SaldoGraphql: SaldoHandleGraphql{
			SaldoClient: SaldoClient{
				SaldoQueryClient:             saldopb.NewSaldoQueryServiceClient(deps.Clients.SaldoClient),
				SaldoCommandClient:           saldopb.NewSaldoCommandServiceClient(deps.Clients.SaldoClient),
				SaldoStatsBalanceClient:      saldostatspb.NewSaldoStatsBalanceServiceClient(deps.Clients.SaldoClient),
				SaldoStatsTotalBalanceClient: saldostatspb.NewSaldoStatsTotalBalanceClient(deps.Clients.SaldoClient),
			},
			Logger:  deps.Logger,
			Mapping: saldographqlmapper.NewSaldoGraphqlMapper(),
			Cache:   cacheSaldo,
		},
		TopupGraphql: TopupHandleGraphql{
			TopupClient: TopupClient{
				TopupQueryClient:   topuppb.NewTopupQueryServiceClient(deps.Clients.TopupClient),
				TopupCommandClient: topuppb.NewTopupCommandServiceClient(deps.Clients.TopupClient),
				TopupStatsAmount:   topupstatpb.NewTopupStatsAmountServiceClient(deps.Clients.TopupClient),
				TopupStatsMethod:   topupstatpb.NewTopupStatsMethodServiceClient(deps.Clients.TopupClient),
				TopupStatsStatus:   topupstatpb.NewTopupStatsStatusServiceClient(deps.Clients.TopupClient),
			},
			Logger:  deps.Logger,
			Mapping: topupgraphqlmapper.NewTopupGraphqlMapper(),
			Cache:   cacheTopup,
		},
		TransactionGraphql: TransactionHandleGraphql{
			TransactionClient: TransactionClient{
				TransactionQueryClient:   transactionpb.NewTransactionQueryServiceClient(deps.Clients.TransactionClient),
				TransactionCommandClient: transactionpb.NewTransactionCommandServiceClient(deps.Clients.TransactionClient),
				TransactionStatsAmount:   transactionstatpb.NewTransactionStatsAmountServiceClient(deps.Clients.TransactionClient),
				TransactionStatsMethod:   transactionstatpb.NewTransactionStatsMethodServiceClient(deps.Clients.TransactionClient),
				TransactionStatsStatus:   transactionstatpb.NewTransactionStatsStatusServiceClient(deps.Clients.TransactionClient),
			},
			Logger:     deps.Logger,
			Mapping:    transactiongraphqlmapper.NewTransactionGraphqlMapper(),
			Permission: merchantpermission.NewMerchantPermission(deps.Kafka, "request-transaction", "response-transaction", 5*time.Second, deps.Logger),
			Cache:      cacheTransaction,
		},
		TransferGraphql: TransferHandleGraphql{
			TransferClient: TransferClient{
				TransferQueryClient:   transferpb.NewTransferQueryServiceClient(deps.Clients.TransferClient),
				TransferCommandClient: transferpb.NewTransferCommandServiceClient(deps.Clients.TransferClient),
				TransferStatsAmount:   transferstatpb.NewTransferStatsAmountServiceClient(deps.Clients.TransferClient),
				TransferStatsStatus:   transferstatpb.NewTransferStatsStatusServiceClient(deps.Clients.TransferClient),
			},
			Logger:  deps.Logger,
			Mapping: transfergraphqlmapper.NewTransferGraphqlMapper(),
			Cache:   cacheTransfer,
		},
		WithdrawGraphql: WithdrawHandleGraphql{
			WithdrawClient: WithdrawClient{
				WithdrawQueryClient:   withdrawpb.NewWithdrawQueryServiceClient(deps.Clients.WithdrawClient),
				WithdrawCommandClient: withdrawpb.NewWithdrawCommandServiceClient(deps.Clients.WithdrawClient),
				WithdrawStatsAmount:   withdrawstatpb.NewWithdrawStatsAmountServiceClient(deps.Clients.WithdrawClient),
				WithdrawStatsStatus:   withdrawstatpb.NewWithdrawStatsStatusServiceClient(deps.Clients.WithdrawClient),
			},
			Logger:  deps.Logger,
			Mapping: withdrawgraphqlmapper.NewWithdrawGraphqlMapper(),
			Cache:   cacheWithdraw,
		},
	}
}

func (h *Resolver) handleGraphQLError(err error, operation string) *errors.AppError {
	if err == nil {
		return nil
	}

	var appErr *errors.AppError
	if errorstd.As(err, &appErr) {
		return appErr
	}

	return errors.NewInternalError(err).WithMessage("Failed to " + operation)
}

func (r *Resolver) parseValidationErrors(err error) []sharedErrors.ValidationError {
	var validationErrs []sharedErrors.ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrs = append(validationErrs, sharedErrors.ValidationError{
				Field:   fe.Field(),
				Message: r.getValidationMessage(fe),
			})
		}
		return validationErrs
	}

	return []sharedErrors.ValidationError{
		{
			Field:   "general",
			Message: err.Error(),
		},
	}
}

func (r *Resolver) getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s", fe.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("Validation failed on '%s' tag", fe.Tag())
	}
}
