package repositorydashboard

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// sumAmount totals one numeric field over the (not soft-deleted) documents that
// match the extra filter.
func sumAmount(ctx context.Context, db *mongo.Database, collection, field string, filter bson.M) (*int64, error) {
	match := bson.M{"deleted_at": nil}
	for key, value := range filter {
		match[key] = value
	}

	cursor, err := db.Collection(collection).Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$" + field},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		Total int64 `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	var total int64
	if len(rows) > 0 {
		total = rows[0].Total
	}
	return &total, nil
}
