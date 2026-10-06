package transactionbycardrepository

import "go.mongodb.org/mongo-driver/mongo"

type TransactionStatsByCardRepository interface {
	TransactionStatsByCardAmountRepository
	TransactionStatsByCardMethodRepository
	TransactionStatsByCardStatusRepository
}

type repository struct {
	TransactionStatsByCardAmountRepository
	TransactionStatsByCardMethodRepository
	TransactionStatsByCardStatusRepository
}

func NewTransactionStatsByCardRepository(db *mongo.Database) TransactionStatsByCardRepository {
	return &repository{
		TransactionStatsByCardAmountRepository: NewTransactionStatsByCardAmountRepository(db),
		TransactionStatsByCardMethodRepository: NewTransactionStatsByCardMethodRepository(db),
		TransactionStatsByCardStatusRepository: NewTransactionStatsByCardStatusRepository(db),
	}
}
