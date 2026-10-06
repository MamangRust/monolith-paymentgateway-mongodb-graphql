package topupstatsrepository

import "go.mongodb.org/mongo-driver/mongo"

type TopupStatsRepository interface {
	TopupStatsAmountRepository
	TopupStatsStatusRepository
	TopupStatsMethodRepository
}

type repository struct {
	TopupStatsAmountRepository
	TopupStatsStatusRepository
	TopupStatsMethodRepository
}

func NewTopupStatsRepository(db *mongo.Database) TopupStatsRepository {
	return &repository{
		TopupStatsAmountRepository: NewTopupStatsAmountRepository(db),
		TopupStatsStatusRepository: NewTopupStatsStatusRepository(db),
		TopupStatsMethodRepository: NewTopupStatsMethodRepository(db),
	}
}
