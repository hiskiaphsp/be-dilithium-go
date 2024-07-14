// repositories/document_repository.go

package repositories

import (
	"context"

	"github.com/hiskiaphsp/be-dilithium-go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DocumentRepository struct {
	Collection *mongo.Collection
}

func NewDocumentRepository(db *mongo.Database, collectionName string) *DocumentRepository {
	return &DocumentRepository{
		Collection: db.Collection(collectionName),
	}
}

func (r *DocumentRepository) Create(ctx context.Context, document *models.Document) (*mongo.InsertOneResult, error) {
	return r.Collection.InsertOne(ctx, document)
}

func (r *DocumentRepository) GetById(ctx context.Context, id string) (*models.Document, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var document models.Document
	err = r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&document)
	return &document, err
}

func (r *DocumentRepository) GetAll(ctx context.Context) ([]models.Document, error) {
	var documents []models.Document
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	return documents, nil
}

func (r *DocumentRepository) Update(ctx context.Context, document *models.Document) (*mongo.UpdateResult, error) {
	objID, err := primitive.ObjectIDFromHex(document.ID.Hex())
	if err != nil {
		return nil, err
	}
	return r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": document})
}

func (r *DocumentRepository) Delete(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return r.Collection.DeleteOne(ctx, bson.M{"_id": objID})
}
