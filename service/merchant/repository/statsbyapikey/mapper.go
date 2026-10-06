package merchantstatsapikeyrepository

import "go.mongodb.org/mongo-driver/mongo"

type MerchantStatsByApiKeyRepository interface {
	MerchantStatsAmountByApiKeyRepository
	MerchantStatsMethodByApiKeyRepository
	MerchantStatsTotalAmountByApiKeyRepository
}

type repository struct {
	MerchantStatsAmountByApiKeyRepository
	MerchantStatsMethodByApiKeyRepository
	MerchantStatsTotalAmountByApiKeyRepository
}

func NewMerchantStatsByApiKeyRepository(db *mongo.Database) MerchantStatsByApiKeyRepository {
	return &repository{
		MerchantStatsAmountByApiKeyRepository:      NewMerchantStatsAmountByApiKeyRepository(db),
		MerchantStatsMethodByApiKeyRepository:      NewMerchantStatsMethodByApiKeyRepository(db),
		MerchantStatsTotalAmountByApiKeyRepository: NewMerchantStatsTotalAmountByApiKeyRepository(db),
	}
}
