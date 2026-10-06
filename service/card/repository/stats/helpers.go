package repositorystats

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func monthName(m time.Month) string {
	return time.Date(0, m, 1, 0, 0, 0, 0, time.UTC).Format("Jan")
}

func yearString(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

// sumByMonth totals field per month for documents whose timeField falls inside
// [start, end).
func sumByMonth(ctx context.Context, db *mongo.Database, collection, timeField, field string, start, end time.Time) (map[int]int64, error) {
	cursor, err := db.Collection(collection).Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			timeField: bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   bson.M{"month": bson.M{"$month": "$" + timeField}},
			"total": bson.M{"$sum": "$" + field},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Month int `bson:"month"`
		} `bson:"_id"`
		Total int64 `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Month] = row.Total
	}
	return totals, nil
}

// sumByYear totals field per year for documents whose timeField falls inside
// [start, end).
func sumByYear(ctx context.Context, db *mongo.Database, collection, timeField, field string, start, end time.Time) (map[int]int64, error) {
	cursor, err := db.Collection(collection).Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			timeField: bson.M{"$gte": start, "$lt": end},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   bson.M{"year": bson.M{"$year": "$" + timeField}},
			"total": bson.M{"$sum": "$" + field},
		}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID struct {
			Year int `bson:"year"`
		} `bson:"_id"`
		Total int64 `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		totals[row.ID.Year] = row.Total
	}
	return totals, nil
}
