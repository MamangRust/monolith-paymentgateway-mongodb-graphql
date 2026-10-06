package transferstatsbycardrepository

import "go.mongodb.org/mongo-driver/mongo"

type TransferStatsByCardRepository interface {
	TransferStatsByCardAmountSenderRepository
	TransferStatsByCardAmountReceiverRepository
	TransferStatsByCardStatusRepository
}

type repository struct {
	TransferStatsByCardAmountSenderRepository
	TransferStatsByCardAmountReceiverRepository
	TransferStatsByCardStatusRepository
}

func NewTransferStatsByCardRepository(db *mongo.Database) TransferStatsByCardRepository {
	return &repository{
		TransferStatsByCardAmountSenderRepository:   NewTransferStatsByCardAmountSenderRepository(db),
		TransferStatsByCardAmountReceiverRepository: NewTransferStatsByCardAmountReceiverRepository(db),
		TransferStatsByCardStatusRepository:         NewTransferStatsByCardStatusRepository(db),
	}
}
