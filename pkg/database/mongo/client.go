package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

func firstNonEmpty(keys ...string) string {
	for _, key := range keys {
		if v := viper.GetString(key); v != "" {
			return v
		}
	}
	return ""
}

func resolveURI() string {
	if uri := firstNonEmpty("MONGO_URI", "DB_URL"); uri != "" {
		if strings.HasPrefix(uri, "mongodb://") || strings.HasPrefix(uri, "mongodb+srv://") {
			return uri
		}
	}
	host := firstNonEmpty("MONGO_HOST", "DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := firstNonEmpty("MONGO_PORT", "DB_PORT")
	if port == "" {
		port = "27017"
	}
	user := firstNonEmpty("MONGO_USERNAME", "MONGO_USER", "DB_USERNAME", "DB_USER")
	password := firstNonEmpty("MONGO_PASSWORD", "DB_PASSWORD")
	if user != "" && password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", user, password, host, port)
	}
	return fmt.Sprintf("mongodb://%s:%s", host, port)
}

func resolveDatabaseName() string {
	name := firstNonEmpty("MONGO_DATABASE", "MONGO_DB_NAME", "DB_NAME")
	if name == "" {
		return "payment_gateway"
	}
	return strings.ToLower(name)
}

func NewMongoClient(log logger.LoggerInterface) (*mongo.Client, error) {
	uri := resolveURI()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	maxPoolSize := viper.GetUint64("MONGO_MAX_POOL_SIZE")
	if maxPoolSize == 0 {
		maxPoolSize = 100
	}
	minPoolSize := viper.GetUint64("MONGO_MIN_POOL_SIZE")
	if minPoolSize == 0 {
		minPoolSize = 10
	}
	clientOpts.SetMaxPoolSize(maxPoolSize)
	clientOpts.SetMinPoolSize(minPoolSize)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Error("Failed to ping MongoDB", zap.Error(err))
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	log.Debug("MongoDB connection established", zap.String("database", resolveDatabaseName()))
	return client, nil
}

func NewDatabase(client *mongo.Client) *mongo.Database {
	return client.Database(resolveDatabaseName())
}

func NextID(ctx context.Context, db *mongo.Database, collection string) (int32, error) {
	coll := db.Collection("counters")
	filter := bson.M{"_id": collection}
	update := bson.M{"$inc": bson.M{"seq": int32(1)}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var result struct {
		Seq int32 `bson:"seq"`
	}
	err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}
