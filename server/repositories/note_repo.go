package repositories

import (
	"context"
	"errors"
	"note_sharing_application/server/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Đại diện cho Note
type NoteRepo struct {
	Collection *mongo.Collection
}

// Nhận đúng bảng note
func newNoteRepo(db *mongo.Database) *NoteRepo {
	return &NoteRepo{
		Collection: db.Collection("note"),
	}
}

func (r *NoteRepo) InsertOne(ctx context.Context, student models.Note) (string, error) {
	// Khi insert, field ID để trống (Zero Value), Mongo sẽ tự sinh
	result, err := r.Collection.InsertOne(ctx, student)
	if err != nil {
		return "", err
	}

	// Lấy ID vừa sinh ra để trả về (ép kiểu sang ObjectID rồi lấy Hex string)
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", nil
}

func (r *NoteRepo) FindByID(ctx context.Context, strId string) (*models.Note, error) {

	// Chuyển từ string sang hex
	hexId, err := primitive.ObjectIDFromHex(strId)
	if err != nil {
		return nil, errors.New("Invalid ID format")
	}

	var note models.Note

	// Tạo filter
	filter := bson.M{"_id": hexId}

	err = r.Collection.FindOne(ctx, filter).Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Can not find note ID")
		}
		return nil, err
	}

	return &note, nil
}

func (r *NoteRepo) DeleteByID(ctx context.Context, idStr string) error {
	// string -> hex
	hexID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return errors.New("Invalid ID format")
	}

	// 2. Delete
	filter := bson.M{"_id": hexID}
	result, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("Can not find note ID to delete")
	}

	return nil
}
