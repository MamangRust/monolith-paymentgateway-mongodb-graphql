package saldostatsrepository

import (
	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type SaldoStatsRepository interface {
	SaldoStatsBalanceRepository
	SaldoStatsTotalSaldoRepository
}

type repository struct {
	SaldoStatsBalanceRepository
	SaldoStatsTotalSaldoRepository
}

func NewSaldoStatsRepository(db *mongo.Database) SaldoStatsRepository {
	return &repository{
		SaldoStatsBalanceRepository:    NewSaldoStatsBalanceRepository(db),
		SaldoStatsTotalSaldoRepository: NewSaldoStatsTotalBalanceRepository(db),
	}
}

func toTotalCount(res []*salemongo.MonthlySaldoBalanceRow) int {
	if len(res) > 0 {
		return len(res)
	}
	return 0
}
