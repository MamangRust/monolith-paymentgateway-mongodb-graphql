package merchantstatsrepository

import "go.mongodb.org/mongo-driver/mongo"

type MerchantStatsRepository interface {
	MerchantStatsAmountRepository
	MerchantStatsMethodRepository
	MerchantStatsTotalAmountRepository
}

type repository struct {
	MerchantStatsAmountRepository
	MerchantStatsMethodRepository
	MerchantStatsTotalAmountRepository
}

func NewMerchantStatsRepository(db *mongo.Database) MerchantStatsRepository {
	return &repository{
		MerchantStatsAmountRepository:      NewMerchantStatsAmountRepository(db),
		MerchantStatsMethodRepository:      NewMerchantStatsMethodRepository(db),
		MerchantStatsTotalAmountRepository: NewMerchantStatsTotalAmountRepository(db),
	}
}
