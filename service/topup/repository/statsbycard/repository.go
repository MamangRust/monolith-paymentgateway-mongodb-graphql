package topupstatsbycardrepository

import "go.mongodb.org/mongo-driver/mongo"

type TopupStatsByCardRepository interface {
	TopupStatsByCardAmountRepository
	TopupStatsByCardMethodRepository
	TopupStatsByCardStatusRepository
}

type repository struct {
	TopupStatsByCardAmountRepository
	TopupStatsByCardMethodRepository
	TopupStatsByCardStatusRepository
}

func NewTopupStatsByCardRepository(db *mongo.Database) TopupStatsByCardRepository {
	return &repository{
		TopupStatsByCardAmountRepository: NewTopupStatsByCardAmountRepository(db),
		TopupStatsByCardMethodRepository: NewTopupStatsByCardMethodRepository(db),
		TopupStatsByCardStatusRepository: NewTopupStatsByCardStatusRepository(db),
	}
}
