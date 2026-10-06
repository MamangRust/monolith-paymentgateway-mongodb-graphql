package transferstatsrepository

import "go.mongodb.org/mongo-driver/mongo"

type TransferStatsRepository interface {
	TransferStatsAmountRepository
	TransferStatsStatusRepository
}

type repositories struct {
	TransferStatsAmountRepository
	TransferStatsStatusRepository
}

func NewTransferStatsRepository(db *mongo.Database) TransferStatsRepository {
	return &repositories{
		TransferStatsAmountRepository: NewTransferStatsAmountRepository(db),
		TransferStatsStatusRepository: NewTransferStatsStatusRepository(db),
	}
}
