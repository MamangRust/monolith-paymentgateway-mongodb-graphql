package transactionstatsrepository

import "go.mongodb.org/mongo-driver/mongo"

type TransactionStatsRepository interface {
	TransactionStatsAmountRepository
	TransactionStatsMethodRepository
	TransactionStatsStatusRepository
}

type repository struct {
	TransactionStatsAmountRepository
	TransactionStatsMethodRepository
	TransactionStatsStatusRepository
}

func NewTransactionStatsRepository(db *mongo.Database) TransactionStatsRepository {
	return &repository{
		TransactionStatsAmountRepository: NewTransactionStatsAmountRepository(db),
		TransactionStatsMethodRepository: NewTransactionStatsMethodRepository(db),
		TransactionStatsStatusRepository: NewTransactionStatsStatusRepository(db),
	}
}
