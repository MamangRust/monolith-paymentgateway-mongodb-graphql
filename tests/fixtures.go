package tests

import (
	"context"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// InsertFixture inserts doc into collection, setting the numeric id field that
// collections index as unique (e.g. "card_id") from the collection's sequence.
// When noField is non-empty a fresh uuid is stored there as well, covering the
// unique "<name>_no" indexes. Stats suites seed with raw bson.M documents and
// would otherwise collide on a null id when more than one row is inserted.
func (ts *TestSuite) InsertFixture(ctx context.Context, collection, idField, noField string, doc bson.M) error {
	id, err := salemongo.NextID(ctx, ts.db, collection)
	if err != nil {
		return err
	}

	doc[idField] = id
	if noField != "" {
		doc[noField] = uuid.New()
	}

	_, err = ts.db.Collection(collection).InsertOne(ctx, doc)
	return err
}
