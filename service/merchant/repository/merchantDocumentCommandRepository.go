package repository

import (
	"context"
	"errors"
	"time"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type merchantDocumentCommandRepository struct {
	db   *mongo.Database
	coll *mongo.Collection
}

func NewMerchantDocumentCommandRepository(db *mongo.Database, coll *mongo.Collection) MerchantDocumentCommandRepository {
	return &merchantDocumentCommandRepository{db: db, coll: coll}
}

func (r *merchantDocumentCommandRepository) CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*salemongo.MerchantDocument, error) {
	documentID, err := salemongo.NextID(ctx, r.db, "merchant_documents")
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	note := ""
	now := time.Now()
	doc := &salemongo.MerchantDocument{
		DocumentID:   documentID,
		MerchantID:   int32(request.MerchantID),
		DocumentType: request.DocumentType,
		DocumentUrl:  request.DocumentUrl,
		Status:       "pending",
		Note:         &note,
		UploadedAt:   now,
		UpdatedAt:    &now,
	}

	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		return nil, sharedErrors.ErrFailed("create merchant document").WithInternal(err)
	}
	return doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*salemongo.MerchantDocument, error) {
	if request.DocumentID == nil || *request.DocumentID <= 0 {
		return nil, sharedErrors.NewBadRequestError("merchant document ID is required")
	}

	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"document_id": *request.DocumentID, "deleted_at": nil},
		bson.M{"$set": bson.M{
			"document_type": request.DocumentType,
			"document_url":  request.DocumentUrl,
			"status":        request.Status,
			"note":          "",
			"updated_at":    now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var doc salemongo.MerchantDocument
	if err := result.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant document")
		}
		return nil, sharedErrors.ErrFailed("update merchant document").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*salemongo.MerchantDocument, error) {
	if request.DocumentID == nil || *request.DocumentID <= 0 {
		return nil, sharedErrors.NewBadRequestError("merchant document ID is required")
	}

	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"document_id": *request.DocumentID, "deleted_at": nil},
		bson.M{"$set": bson.M{
			"status":     request.Status,
			"note":       request.Note,
			"updated_at": now,
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var doc salemongo.MerchantDocument
	if err := result.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant document")
		}
		return nil, sharedErrors.ErrFailed("update merchant document status").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) TrashedMerchantDocument(ctx context.Context, documentID int) (*salemongo.MerchantDocument, error) {
	now := time.Now()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"document_id": documentID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var doc salemongo.MerchantDocument
	if err := result.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant document")
		}
		return nil, sharedErrors.ErrFailed("trash merchant document").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) RestoreMerchantDocument(ctx context.Context, documentID int) (*salemongo.MerchantDocument, error) {
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"document_id": documentID, "deleted_at": bson.M{"$ne": nil}},
		bson.M{"$unset": bson.M{"deleted_at": ""}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var doc salemongo.MerchantDocument
	if err := result.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant document")
		}
		return nil, sharedErrors.ErrFailed("restore merchant document").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) DeleteMerchantDocumentPermanent(ctx context.Context, documentID int) (bool, error) {
	result, err := r.coll.DeleteOne(ctx, bson.M{"document_id": documentID})
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "delete merchant document permanently")
	}
	if result.DeletedCount == 0 {
		return false, sharedErrors.ErrNotFoundResponse("merchant document")
	}
	return true, nil
}

func (r *merchantDocumentCommandRepository) RestoreAllMerchantDocument(ctx context.Context) (bool, error) {
	_, err := r.coll.UpdateMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}}, bson.M{"$unset": bson.M{"deleted_at": ""}})
	if err != nil {
		return false, sharedErrors.ErrFailed("restore all merchant documents").WithInternal(err)
	}
	return true, nil
}

func (r *merchantDocumentCommandRepository) DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error) {
	_, err := r.coll.DeleteMany(ctx, bson.M{"deleted_at": bson.M{"$ne": nil}})
	if err != nil {
		return false, sharedErrors.ErrFailed("delete all merchant documents permanently").WithInternal(err)
	}
	return true, nil
}
