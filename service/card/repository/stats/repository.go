package repositorystats

import "go.mongodb.org/mongo-driver/mongo"

type CardStatsRepository interface {
	CardStatsBalanceRepository
	CardStatsTopupRepository
	CardStatsTransactionRepository
	CardStatsTransferRepository
	CardStatsWithdrawRepository
}

type repository struct {
	CardStatsBalanceRepository
	CardStatsTopupRepository
	CardStatsTransactionRepository
	CardStatsTransferRepository
	CardStatsWithdrawRepository
}

func NewCardStatsRepository(db *mongo.Database) CardStatsRepository {

	return &repository{
		CardStatsBalanceRepository:     NewCardStatsBalanceRepository(db),
		CardStatsTopupRepository:       NewCardStatsTopupRepository(db),
		CardStatsTransactionRepository: NewCardStatsTransactionRepository(db),
		CardStatsTransferRepository:    NewCardStatsTransferRepository(db),
		CardStatsWithdrawRepository:    NewCardStatsWithdrawRepository(db),
	}
}
