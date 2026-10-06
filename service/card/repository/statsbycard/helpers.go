package repositorystatsbycard

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

// aggregateTotals groups the sum of field by month or year (groupBy) for the
// documents matching filter inside [start, end).
func aggregateTotals(ctx context.Context, db *mongo.Database, collection, timeField, field, groupBy string, filter bson.M, start, end time.Time) (map[int]int64, error) {
	match := bson.M{timeField: bson.M{"$gte": start, "$lt": end}}
	for key, value := range filter {
		match[key] = value
	}

	cursor, err := db.Collection(collection).Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.M{
			"_id":   bson.M{groupBy: bson.M{"$" + groupBy: "$" + timeField}},
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
			Year  int `bson:"year"`
		} `bson:"_id"`
		Total int64 `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	totals := make(map[int]int64, len(rows))
	for _, row := range rows {
		key := row.ID.Month
		if groupBy == "year" {
			key = row.ID.Year
		}
		totals[key] = row.Total
	}
	return totals, nil
}
