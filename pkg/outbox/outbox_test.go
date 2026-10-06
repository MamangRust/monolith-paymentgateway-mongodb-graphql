package outbox

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type testLogger struct{}

func (testLogger) Info(string, ...zap.Field)                         {}
func (testLogger) Fatal(string, ...zap.Field)                        {}
func (testLogger) Debug(string, ...zap.Field)                        {}
func (testLogger) Error(string, ...zap.Field)                        {}
func (testLogger) Warn(string, ...zap.Field)                         {}
func (testLogger) Check(zapcore.Level, string) *zapcore.CheckedEntry { return nil }
func (testLogger) With(...zap.Field) logger.LoggerInterface          { return testLogger{} }
func (testLogger) Sync() error                                       { return nil }

type testPublisher struct {
	mu       sync.Mutex
	calls    int
	topics   []string
	keys     []string
	payloads [][]byte
	err      error
}

func (p *testPublisher) SendMessage(_ context.Context, topic, key string, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	p.topics = append(p.topics, topic)
	p.keys = append(p.keys, key)
	p.payloads = append(p.payloads, value)
	return p.err
}

func TestOutboxServiceNilCollectionIsNoop(t *testing.T) {
	svc := &OutboxService{logger: testLogger{}}
	ctx := context.Background()

	if err := svc.Enqueue(ctx, "topup.created", "42", []byte("{}")); err != nil {
		t.Fatalf("Enqueue() error = %v, want nil", err)
	}
	if err := svc.EnqueueInTx(ctx, nil, "topup.created", "42", []byte("{}")); err != nil {
		t.Fatalf("EnqueueInTx() error = %v, want nil", err)
	}
	if delivered, err := svc.PublishPending(ctx, 10); err != nil || delivered != 0 {
		t.Fatalf("PublishPending() = (%d, %v), want (0, nil)", delivered, err)
	}
	if err := svc.PurgeRetention(ctx); err != nil {
		t.Fatalf("PurgeRetention() error = %v, want nil", err)
	}
}

func TestOutboxRelayStopsOnContextCancel(t *testing.T) {
	svc := &OutboxService{logger: testLogger{}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		svc.Start(ctx, time.Millisecond, 1)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not stop after context cancellation")
	}
}

func TestOutboxService(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("enqueue persists a pending event", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		svc := NewOutboxService(mt.Coll.Database(), nil, testLogger{})

		payload := []byte(`{"id":42}`)
		if err := svc.Enqueue(context.Background(), "topup.created", "42", payload); err != nil {
			t.Fatalf("Enqueue() error = %v", err)
		}

		cmd, ok := startedCommand(mt, "insert")
		if !ok {
			mt.Fatal("Enqueue did not send an insert command")
		}
		if got := cmd.Lookup("insert").StringValue(); got != "outbox_events" {
			mt.Fatalf("inserted into collection %q, want outbox_events", got)
		}
		doc, err := cmd.LookupErr("documents")
		if err != nil {
			mt.Fatalf("insert command has no documents: %v", err)
		}
		first := doc.Array().Index(0).Value().Document()
		if got := first.Lookup("event_key").StringValue(); got != "42" {
			mt.Fatalf("event_key = %q, want 42", got)
		}
		if got := first.Lookup("topic").StringValue(); got != "topup.created" {
			mt.Fatalf("topic = %q, want topup.created", got)
		}
		if got := first.Lookup("status").StringValue(); got != "pending" {
			mt.Fatalf("status = %q, want pending", got)
		}
		if _, got := first.Lookup("payload").Binary(); !bytes.Equal(got, payload) {
			mt.Fatalf("payload = %s, want %s", got, payload)
		}
	})

	mt.Run("enqueue surfaces write errors", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index: 0, Code: 11000, Message: "duplicate key",
		}))
		svc := NewOutboxService(mt.Coll.Database(), nil, testLogger{})

		if err := svc.Enqueue(context.Background(), "topic", "key", []byte("{}")); err == nil {
			mt.Fatal("Enqueue() error = nil, want write error")
		}
	})

	mt.Run("enqueue-in-tx delegates to enqueue", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		svc := NewOutboxService(mt.Coll.Database(), nil, testLogger{})

		if err := svc.EnqueueInTx(context.Background(), nil, "topic", "key", []byte("{}")); err != nil {
			t.Fatalf("EnqueueInTx() error = %v", err)
		}
		if _, ok := startedCommand(mt, "insert"); !ok {
			mt.Fatal("EnqueueInTx did not send an insert command")
		}
	})

	mt.Run("publish delivers claimed events", func(mt *mtest.T) {
		mt.AddMockResponses(
			findAndModifyReply(claimEventDoc(0)),
			bson.D{{Key: "ok", Value: 1}}, // no more claimable events
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
		)
		pub := &testPublisher{}
		svc := NewOutboxService(mt.Coll.Database(), pub, testLogger{})

		delivered, err := svc.PublishPending(context.Background(), 5)
		if err != nil {
			t.Fatalf("PublishPending() error = %v", err)
		}
		if delivered != 1 {
			t.Fatalf("delivered = %d, want 1", delivered)
		}
		if pub.calls != 1 {
			t.Fatalf("publisher calls = %d, want 1", pub.calls)
		}
		if pub.topics[0] != "topup.created" || pub.keys[0] != "topup:42" || !bytes.Equal(pub.payloads[0], []byte(`{"id":42}`)) {
			t.Fatalf("publisher got topic=%s key=%s payload=%s", pub.topics[0], pub.keys[0], pub.payloads[0])
		}

		filter, update, err := firstUpdateDoc(mt)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Lookup("_id").Type == 0 {
			t.Fatal("delivered update does not target the claimed event _id")
		}
		set, err := update.LookupErr("$set")
		if err != nil {
			t.Fatal(err)
		}
		if got := set.Document().Lookup("status").StringValue(); got != "delivered" {
			t.Fatalf("delivered update status = %q, want delivered", got)
		}
	})

	mt.Run("publish failure schedules retry with backoff", func(mt *mtest.T) {
		mt.AddMockResponses(
			findAndModifyReply(claimEventDoc(2)),
			bson.D{{Key: "ok", Value: 1}},
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
		)
		pub := &testPublisher{err: errors.New("kafka broker down")}
		svc := NewOutboxService(mt.Coll.Database(), pub, testLogger{})

		delivered, err := svc.PublishPending(context.Background(), 5)
		if err != nil {
			t.Fatalf("PublishPending() error = %v", err)
		}
		if delivered != 0 {
			t.Fatalf("delivered = %d, want 0", delivered)
		}
		if pub.calls != 1 {
			t.Fatalf("publisher calls = %d, want 1", pub.calls)
		}

		_, update, err := firstUpdateDoc(mt)
		if err != nil {
			t.Fatal(err)
		}
		set, err := update.LookupErr("$set")
		if err != nil {
			t.Fatal(err)
		}
		setDoc := set.Document()
		if got := setDoc.Lookup("attempts").Int32(); got != 3 {
			t.Fatalf("retry attempts = %d, want 3", got)
		}
		nextAttempt := setDoc.Lookup("next_attempt_at").Time()
		if nextAttempt.Before(time.Now().Add(OutboxBackoff)) {
			t.Fatalf("next_attempt_at = %v, want at least one backoff interval in the future", nextAttempt)
		}
		if _, err := setDoc.LookupErr("status"); err == nil {
			t.Fatal("retry update must not change the event status")
		}
	})

	mt.Run("publish marks events dead after max attempts", func(mt *mtest.T) {
		mt.AddMockResponses(
			findAndModifyReply(claimEventDoc(OutboxMaxAttempts-1)),
			bson.D{{Key: "ok", Value: 1}},
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
		)
		pub := &testPublisher{err: errors.New("kafka broker down")}
		svc := NewOutboxService(mt.Coll.Database(), pub, testLogger{})

		delivered, err := svc.PublishPending(context.Background(), 5)
		if err != nil {
			t.Fatalf("PublishPending() error = %v", err)
		}
		if delivered != 0 {
			t.Fatalf("delivered = %d, want 0", delivered)
		}

		_, update, err := firstUpdateDoc(mt)
		if err != nil {
			t.Fatal(err)
		}
		set, err := update.LookupErr("$set")
		if err != nil {
			t.Fatal(err)
		}
		if got := set.Document().Lookup("status").StringValue(); got != "dead" {
			t.Fatalf("status = %q, want dead after %d attempts", got, OutboxMaxAttempts)
		}
	})

	mt.Run("publish is a noop without a publisher", func(mt *mtest.T) {
		svc := NewOutboxService(mt.Coll.Database(), nil, testLogger{})

		delivered, err := svc.PublishPending(context.Background(), 5)
		if err != nil {
			t.Fatalf("PublishPending() error = %v", err)
		}
		if delivered != 0 {
			t.Fatalf("delivered = %d, want 0", delivered)
		}
		if events := mt.GetAllStartedEvents(); len(events) != 0 {
			t.Fatalf("sent %d commands, want none", len(events))
		}
	})

	mt.Run("purge deletes only terminal events past retention", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(2)}))
		svc := NewOutboxService(mt.Coll.Database(), nil, testLogger{})

		if err := svc.PurgeRetention(context.Background()); err != nil {
			t.Fatalf("PurgeRetention() error = %v", err)
		}

		cmd, ok := startedCommand(mt, "delete")
		if !ok {
			mt.Fatal("PurgeRetention did not send a delete command")
		}
		if got := cmd.Lookup("delete").StringValue(); got != "outbox_events" {
			mt.Fatalf("deleted from collection %q, want outbox_events", got)
		}
		deletes, err := cmd.LookupErr("deletes")
		if err != nil {
			mt.Fatal(err)
		}
		first := deletes.Array().Index(0).Value().Document()
		filter, err := first.LookupErr("q")
		if err != nil {
			mt.Fatal(err)
		}
		statusIn, err := filter.Document().LookupErr("status", "$in")
		if err != nil {
			mt.Fatalf("purge filter does not restrict status: %v", err)
		}
		values, err := statusIn.Array().Values()
		if err != nil {
			mt.Fatal(err)
		}
		var statuses []string
		for _, v := range values {
			statuses = append(statuses, v.StringValue())
		}
		if len(statuses) != 2 || statuses[0] != "delivered" || statuses[1] != "dead" {
			t.Fatalf("purge statuses = %v, want [delivered dead]", statuses)
		}
	})
}

func TestConsumerInbox(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("reserve validates keys", func(mt *mtest.T) {
		db := mt.Coll.Database()
		cases := []struct {
			name     string
			db       *mongo.Database
			consumer string
			key      string
		}{
			{name: "nil db", db: nil, consumer: "consumer-1", key: "event-1"},
			{name: "empty consumer", db: db, consumer: "", key: "event-1"},
			{name: "empty event key", db: db, consumer: "consumer-1", key: ""},
		}
		for _, tc := range cases {
			mt.Run(tc.name, func(mt *mtest.T) {
				_, _, _, err := Reserve(context.Background(), tc.db, tc.consumer, tc.key, "topic", 0, 0)
				if !errors.Is(err, ErrInvalidInboxKey) {
					t.Fatalf("Reserve() error = %v, want ErrInvalidInboxKey", err)
				}
				if events := mt.GetAllStartedEvents(); len(events) != 0 {
					t.Fatalf("sent %d commands, want none", len(events))
				}
			})
		}
	})

	mt.Run("reserve claims a new event via upsert", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: int32(1)},
			bson.E{Key: "upserted", Value: bson.A{bson.D{
				{Key: "index", Value: int32(0)},
				{Key: "_id", Value: primitive.NewObjectID()},
			}}},
		))
		reserved, processed, version, err := Reserve(context.Background(),
			mt.Coll.Database(), "consumer-1", "event-1", "topup.created", 2, 100)
		if err != nil {
			t.Fatalf("Reserve() error = %v", err)
		}
		if !reserved || processed || version != 1 {
			t.Fatalf("Reserve() = (reserved=%t, processed=%t, version=%d), want (true, false, 1)",
				reserved, processed, version)
		}

		cmd, ok := startedCommand(mt, "update")
		if !ok {
			t.Fatal("Reserve did not send an update command")
		}
		if got := cmd.Lookup("update").StringValue(); got != "consumer_inbox" {
			t.Fatalf("updated collection %q, want consumer_inbox", got)
		}
		filter, update, err := firstUpdateDoc(mt)
		if err != nil {
			t.Fatal(err)
		}
		if got := filter.Lookup("consumer_name").StringValue(); got != "consumer-1" {
			t.Fatalf("filter consumer_name = %q, want consumer-1", got)
		}
		if got := filter.Lookup("event_key").StringValue(); got != "event-1" {
			t.Fatalf("filter event_key = %q, want event-1", got)
		}
		if _, err := update.LookupErr("$setOnInsert"); err != nil {
			t.Fatalf("upsert does not seed the inbox row on insert: %v", err)
		}
		if _, err := update.LookupErr("$inc", "attempts"); err != nil {
			t.Fatalf("upsert does not count attempts: %v", err)
		}
		if _, err := update.LookupErr("$inc", "reservation_version"); err != nil {
			t.Fatalf("upsert does not bump reservation_version: %v", err)
		}
	})

	mt.Run("reserve reports an already processed event", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
			mtest.CreateCursorResponse(1, "test.consumer_inbox", mtest.FirstBatch, bson.D{
				{Key: "status", Value: inboxStatusProcessed},
				{Key: "reservation_version", Value: int64(9)},
			}),
		)
		reserved, processed, version, err := Reserve(context.Background(),
			mt.Coll.Database(), "consumer-1", "event-1", "topup.created", 2, 100)
		if err != nil {
			t.Fatalf("Reserve() error = %v", err)
		}
		if reserved || !processed || version != 9 {
			t.Fatalf("Reserve() = (reserved=%t, processed=%t, version=%d), want (false, true, 9)",
				reserved, processed, version)
		}
	})

	mt.Run("reserve reports an active reservation for a stalled event", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
			mtest.CreateCursorResponse(1, "test.consumer_inbox", mtest.FirstBatch, bson.D{
				{Key: "status", Value: inboxStatusProcessing},
				{Key: "reservation_version", Value: int64(7)},
			}),
		)
		reserved, processed, version, err := Reserve(context.Background(),
			mt.Coll.Database(), "consumer-1", "event-1", "topup.created", 2, 100)
		if err != nil {
			t.Fatalf("Reserve() error = %v", err)
		}
		if reserved || processed || version != 7 {
			t.Fatalf("Reserve() = (reserved=%t, processed=%t, version=%d), want (false, false, 7)",
				reserved, processed, version)
		}
	})

	mt.Run("mark processed fences on the active reservation", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}))
		if err := MarkProcessed(context.Background(), mt.Coll.Database(), "consumer-1", "event-1", 3); err != nil {
			t.Fatalf("MarkProcessed() error = %v", err)
		}

		filter, update, err := firstUpdateDoc(mt)
		if err != nil {
			t.Fatal(err)
		}
		if got := filter.Lookup("status").StringValue(); got != inboxStatusProcessing {
			t.Fatalf("filter status = %q, want processing", got)
		}
		if got := filter.Lookup("reservation_version").Int64(); got != 3 {
			t.Fatalf("filter reservation_version = %d, want 3", got)
		}
		set, err := update.LookupErr("$set")
		if err != nil {
			t.Fatal(err)
		}
		setDoc := set.Document()
		if got := setDoc.Lookup("status").StringValue(); got != inboxStatusProcessed {
			t.Fatalf("status = %q, want processed", got)
		}
		if _, err := setDoc.LookupErr("processed_at"); err != nil {
			t.Fatalf("processed update does not stamp processed_at: %v", err)
		}
	})

	mt.Run("release returns the event to pending with the failure", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}))
		err := Release(context.Background(), mt.Coll.Database(), "consumer-1", "event-1", 3,
			errors.New("handler exploded"))
		if err != nil {
			t.Fatalf("Release() error = %v", err)
		}

		filter, update, err := firstUpdateDoc(mt)
		if err != nil {
			t.Fatal(err)
		}
		if got := filter.Lookup("reservation_version").Int64(); got != 3 {
			t.Fatalf("filter reservation_version = %d, want 3", got)
		}
		set, err := update.LookupErr("$set")
		if err != nil {
			t.Fatal(err)
		}
		setDoc := set.Document()
		if got := setDoc.Lookup("status").StringValue(); got != inboxStatusPending {
			t.Fatalf("status = %q, want pending", got)
		}
		if got := setDoc.Lookup("last_error").StringValue(); got != "handler exploded" {
			t.Fatalf("last_error = %q, want handler exploded", got)
		}
	})

	mt.Run("mongo inbox adapter delegates to the package functions", func(mt *mtest.T) {
		mt.AddMockResponses(
			// Reserve: fresh upserted claim
			mtest.CreateSuccessResponse(
				bson.E{Key: "n", Value: int32(1)},
				bson.E{Key: "upserted", Value: bson.A{bson.D{
					{Key: "index", Value: int32(0)},
					{Key: "_id", Value: primitive.NewObjectID()},
				}}},
			),
			// MarkProcessed
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
			// Release
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
		)
		inbox, err := NewMongoInbox(mt.Coll.Database())
		if err != nil {
			t.Fatalf("NewMongoInbox() error = %v", err)
		}
		ctx := context.Background()
		reserved, processed, version, err := inbox.Reserve(ctx, "consumer-1", "event-1", "topup.created", 2, 100)
		if err != nil || !reserved || processed || version != 1 {
			t.Fatalf("inbox.Reserve() = (%t, %t, %d, %v), want (true, false, 1, nil)",
				reserved, processed, version, err)
		}
		if err := inbox.MarkProcessed(ctx, "consumer-1", "event-1", 1); err != nil {
			t.Fatalf("inbox.MarkProcessed() error = %v", err)
		}
		if err := inbox.Release(ctx, "consumer-1", "event-1", 1, nil); err != nil {
			t.Fatalf("inbox.Release() error = %v", err)
		}
	})
}

func TestNewMongoInboxRejectsNilDB(t *testing.T) {
	if _, err := NewMongoInbox(nil); err == nil {
		t.Fatal("NewMongoInbox(nil) error = nil, want error")
	}
}

// findAndModifyReply builds a findAndModify response carrying the given document.
func findAndModifyReply(value bson.D) bson.D {
	return bson.D{
		{Key: "lastErrorObject", Value: bson.D{{Key: "n", Value: int32(1)}}},
		{Key: "value", Value: value},
		{Key: "ok", Value: 1},
	}
}

// claimEventDoc is an outbox event as claimed by PublishPending.
func claimEventDoc(attempts int32) bson.D {
	return bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "outbox_id", Value: int64(1)},
		{Key: "topic", Value: "topup.created"},
		{Key: "event_key", Value: "topup:42"},
		{Key: "payload", Value: []byte(`{"id":42}`)},
		{Key: "status", Value: "pending"},
		{Key: "attempts", Value: attempts},
	}
}

func startedCommand(mt *mtest.T, name string) (bson.Raw, bool) {
	for _, event := range mt.GetAllStartedEvents() {
		if event.CommandName == name {
			return event.Command, true
		}
	}
	return nil, false
}

// firstUpdateDoc returns the filter and update document of the first update command sent.
func firstUpdateDoc(mt *mtest.T) (filter, update bson.Raw, err error) {
	cmd, ok := startedCommand(mt, "update")
	if !ok {
		return nil, nil, errors.New("no update command was sent")
	}
	updates, err := cmd.LookupErr("updates")
	if err != nil {
		return nil, nil, err
	}
	first, ok := updates.Array().Index(0).Value().DocumentOK()
	if !ok {
		return nil, nil, errors.New("update command is malformed")
	}
	q, ok := first.Lookup("q").DocumentOK()
	if !ok {
		return nil, nil, errors.New("update command has no filter document")
	}
	u, ok := first.Lookup("u").DocumentOK()
	if !ok {
		return nil, nil, errors.New("update command has no update document")
	}
	return q, u, nil
}
