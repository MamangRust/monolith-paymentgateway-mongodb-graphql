package withdrawstatsrepository

import "go.mongodb.org/mongo-driver/mongo"

type WithdrawStatsRepository interface {
	WithdrawStatsAmountRepository
	WithdrawStatsStatusRepository
}

type repositories struct {
	WithdrawStatsAmountRepository
	WithdrawStatsStatusRepository
}

func NewWithdrawStatsRepository(db *mongo.Database) WithdrawStatsRepository {
	return &repositories{
		WithdrawStatsAmountRepository: NewWithdrawStatsAmountRepository(db),
		WithdrawStatsStatusRepository: NewWithdrawStatsStatusRepository(db),
	}
}
