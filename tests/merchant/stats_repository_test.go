package merchant_test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/google/uuid"

	stats_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/stats"
	apikey_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbyapikey"
	merchant_repo "github.com/MamangRust/monolith-payment-gateway-merchant/repository/statsbymerchant"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/stretchr/testify/suite"
)

type MerchantStatsRepositoryTestSuite struct {
	suite.Suite
	ts           *tests.TestSuite
	repo         stats_repo.MerchantStatsRepository
	apikeyRepo   apikey_repo.MerchantStatsByApiKeyRepository
	merchantRepo merchant_repo.MerchantStatsByMerchantRepository
	merchantID1  int32
	merchantID2  int32
	apiKey1      string
	apiKey2      string
	testYear     int
}

func (s *MerchantStatsRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.repo = stats_repo.NewMerchantStatsRepository(mongoDB)
	s.apikeyRepo = apikey_repo.NewMerchantStatsByApiKeyRepository(mongoDB)
	s.merchantRepo = merchant_repo.NewMerchantStatsByMerchantRepository(mongoDB)

	s.testYear = time.Now().Year()

	// Seed Data. The seeded collections carry unique indexes on their numeric
	// id and on their *_no uuid, so fixtures must set both explicitly.
	ctx := context.Background()
	_, err = mongoDB.Collection("users").InsertOne(ctx, bson.M{"user_id": 1, "firstname": "Merchant", "lastname": "Stats", "email": "merchant_stats@example.com", "password": "pass", "verification_code": "123", "is_verified": true})
	s.Require().NoError(err)

	s.apiKey1 = "merchant-key-1"
	s.apiKey2 = "merchant-key-2"

	s.merchantID1 = 1
	_, err = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": s.merchantID1, "merchant_no": uuid.New(), "name": "Merchant 1", "api_key": s.apiKey1, "status": "active"})
	s.Require().NoError(err)

	s.merchantID2 = 2
	_, err = mongoDB.Collection("merchants").InsertOne(ctx, bson.M{"merchant_id": s.merchantID2, "merchant_no": uuid.New(), "name": "Merchant 2", "api_key": s.apiKey2, "status": "active"})
	s.Require().NoError(err)

	// Seed cards for transactions
	cardNumber1 := "1111222233334444"
	cardNumber2 := "5555666677778888"
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 1, "card_number": cardNumber1, "card_type": "debit", "cvv": "123", "card_provider": "visa", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("cards").InsertOne(ctx, bson.M{"card_id": 2, "card_number": cardNumber2, "card_type": "debit", "cvv": "123", "card_provider": "mastercard", "expire_date": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)})
	s.Require().NoError(err)

	// Seed Transactions
	// Merchant 1: Jan (500, visa), Feb (300, mastercard)
	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 1, "transaction_no": uuid.New(), "card_number": cardNumber1, "amount": 500, "payment_method": "visa", "merchant_id": s.merchantID1, "transaction_time": time.Date(s.testYear, 1, 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 2, "transaction_no": uuid.New(), "card_number": cardNumber2, "amount": 300, "payment_method": "mastercard", "merchant_id": s.merchantID1, "transaction_time": time.Date(s.testYear, 2, 10, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)

	// Merchant 2: Jan (1000, visa)
	_, err = mongoDB.Collection("transactions").InsertOne(ctx, bson.M{"transaction_id": 3, "transaction_no": uuid.New(), "card_number": cardNumber1, "amount": 1000, "payment_method": "visa", "merchant_id": s.merchantID2, "transaction_time": time.Date(s.testYear, 1, 15, 10, 0, 0, 0, time.UTC), "status": "success"})
	s.Require().NoError(err)
}

func (s *MerchantStatsRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *MerchantStatsRepositoryTestSuite) TestGlobalStats() {
	ctx := context.Background()

	// Monthly Amount
	res, err := s.repo.GetMonthlyAmountMerchant(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(res)
	// Jan: 500 (M1) + 1000 (M2) = 1500
	s.Equal(int32(1500), res[0].TotalAmount)
	// Feb: 300
	s.Equal(int32(300), res[1].TotalAmount)

	// Monthly Method
	methRes, err := s.repo.GetMonthlyPaymentMethodsMerchant(ctx, s.testYear)
	s.NoError(err)
	s.NotEmpty(methRes)
}

func (s *MerchantStatsRepositoryTestSuite) TestMerchantStats() {
	ctx := context.Background()

	// Merchant 1 Monthly
	res1, err := s.merchantRepo.GetMonthlyAmountByMerchants(ctx, &requests.MonthYearAmountMerchant{Year: s.testYear, MerchantID: int(s.merchantID1)})
	s.NoError(err)
	s.NotEmpty(res1)
	s.Equal(int32(500), res1[0].TotalAmount)
	s.Equal(int32(300), res1[1].TotalAmount)

	// Merchant 2 Monthly
	res2, err := s.merchantRepo.GetMonthlyAmountByMerchants(ctx, &requests.MonthYearAmountMerchant{Year: s.testYear, MerchantID: int(s.merchantID2)})
	s.NoError(err)
	s.Equal(int32(1000), res2[0].TotalAmount)
	s.Equal(int32(0), res2[1].TotalAmount)
}

func (s *MerchantStatsRepositoryTestSuite) TestApiKeyStats() {
	ctx := context.Background()

	// API Key 1 Monthly
	res1, err := s.apikeyRepo.GetMonthlyAmountByApikey(ctx, &requests.MonthYearAmountApiKey{Year: s.testYear, Apikey: s.apiKey1})
	s.NoError(err)
	s.NotEmpty(res1)
	s.Equal(int32(500), res1[0].TotalAmount)

	// API Key 2 Monthly
	res2, err := s.apikeyRepo.GetMonthlyAmountByApikey(ctx, &requests.MonthYearAmountApiKey{Year: s.testYear, Apikey: s.apiKey2})
	s.NoError(err)
	s.Equal(int32(1000), res2[0].TotalAmount)
}

func TestMerchantStatsRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantStatsRepositoryTestSuite))
}
