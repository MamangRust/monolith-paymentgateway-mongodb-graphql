package repositorydashboard

import "go.mongodb.org/mongo-driver/mongo"

type CardDashboardRepository interface {
	CardDashboardBalanceRepository
	CardDashboardTopupRepository
	CardDashboardTransactionRepository
	CardDashboardTransferRepository
	CardDashboardWithdrawRepository
}

type repository struct {
	CardDashboardBalanceRepository
	CardDashboardTopupRepository
	CardDashboardTransactionRepository
	CardDashboardTransferRepository
	CardDashboardWithdrawRepository
}

func NewCardDashboardRepository(db *mongo.Database) CardDashboardRepository {
	return &repository{
		CardDashboardBalanceRepository:     NewCardDashboardBalanceRepository(db),
		CardDashboardTopupRepository:       NewCardDashboardTopupRepository(db),
		CardDashboardTransactionRepository: NewCardDashboardTransactionRepository(db),
		CardDashboardTransferRepository:    NewCardDashboardTransferRepository(db),
		CardDashboardWithdrawRepository:    NewCardDashboardWithdrawRepository(db),
	}
}
