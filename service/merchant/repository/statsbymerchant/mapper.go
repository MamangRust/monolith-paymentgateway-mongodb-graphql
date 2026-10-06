package merchantstatsmerchantrepository

import "go.mongodb.org/mongo-driver/mongo"

type MerchantStatsByMerchantRepository interface {
	MerchantStatsAmountByMerchantRepository
	MerchantStatsMethodByMerchantRepository
	MerchantStatsTotalAmountByMerchantRepository
}

type repository struct {
	MerchantStatsAmountByMerchantRepository
	MerchantStatsMethodByMerchantRepository
	MerchantStatsTotalAmountByMerchantRepository
}

func NewMerchantStatsByMerchantRepository(db *mongo.Database) MerchantStatsByMerchantRepository {
	return &repository{
		MerchantStatsAmountByMerchantRepository:      NewMerchantStatsAmountByMerchantRepository(db),
		MerchantStatsMethodByMerchantRepository:      NewMerchantStatsMethodByMerchantRepository(db),
		MerchantStatsTotalAmountByMerchantRepository: NewMerchantStatsTotalAmountByMerchantRepository(db),
	}
}
