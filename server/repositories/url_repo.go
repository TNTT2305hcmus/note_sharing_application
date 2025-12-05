package repositories

import (
	"context"
	"errors"
	"note_sharing_application/server/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository cho URL
type UrlRepo struct {
	Collection *mongo.Collection
}

// Khởi tạo repository, trỏ đúng collection "url"
func NewUrlRepo(db *mongo.Database) *UrlRepo {
	return &UrlRepo{
		Collection: db.Collection("url"),
	}
}

// Insert một URL
func (r *UrlRepo) InsertOne(ctx context.Context, url models.Url) (string, error) {
	result, err := r.Collection.InsertOne(ctx, url)
	if err != nil {
		return "", err
	}

	// Lấy ObjectID vừa sinh
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", nil
}

// Tìm URL theo ID
func (r *UrlRepo) FindByID(ctx context.Context, strId string) (*models.Url, error) {

	// Chuyển string -> ObjectID
	hexId, err := primitive.ObjectIDFromHex(strId)
	if err != nil {
		return nil, errors.New("Invalid URL ID format")
	}

	var url models.Url

	filter := bson.M{"_id": hexId}

	err = r.Collection.FindOne(ctx, filter).Decode(&url)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("URL not found")
		}
		return nil, err
	}

	return &url, nil
}

// Xóa URL theo ID
func (r *UrlRepo) DeleteByID(ctx context.Context, idStr string) error {
	hexID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return errors.New("Invalid URL ID format")
	}

	filter := bson.M{"_id": hexID}

	result, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("URL not found to delete")
	}

	return nil
}
