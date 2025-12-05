package repositories

import (
	"context"
	"errors"
	"note_sharing_application/server/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	Collection *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{
		Collection: db.Collection("user"),
	}
}

// Insert User
func (r *UserRepo) InsertOne(ctx context.Context, user models.User) error {
	_, err := r.Collection.InsertOne(ctx, user)
	return err
}

// Find user by string ID
func (r *UserRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User

	filter := bson.M{"id": id}

	err := r.Collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("User not found")
		}
		return nil, err
	}

	return &user, nil
}

// Find user by username
func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	filter := bson.M{"username": username}

	err := r.Collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("User not found")
		}
		return nil, err
	}

	return &user, nil
}

// Delete user by string ID
func (r *UserRepo) DeleteByID(ctx context.Context, id string) error {
	filter := bson.M{"id": id}

	result, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("User not found to delete")
	}

	return nil
}
