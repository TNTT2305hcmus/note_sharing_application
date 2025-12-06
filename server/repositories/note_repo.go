package repositories

import (
	"context"
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
func NewNoteRepo(db *mongo.Database) *NoteRepo {
	return &NoteRepo{
		Collection: db.Collection("note"),
	}
}

func (r *NoteRepo) InsertOne(ctx context.Context, note models.Note) (string, error) {
	// Khi insert, field ID để trống (Zero Value), Mongo sẽ tự sinh
	result, err := r.Collection.InsertOne(ctx, note)
	if err != nil {
		return "", err
	}

	// Lấy ID vừa sinh ra để trả về (ép kiểu sang ObjectID rồi lấy Hex string)
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", nil
}

func (r *NoteRepo) FindByID(ctx context.Context, noteID primitive.ObjectID) (*models.Note, error) {
	var note models.Note
	err := r.Collection.FindOne(ctx, bson.M{"_id": noteID}).Decode(&note)
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *NoteRepo) DeleteByID(ctx context.Context, noteID primitive.ObjectID) error {
	_, err := r.Collection.DeleteOne(ctx, bson.M{"_id": noteID})
	return err
}

// Tìm tất cả note do owner tạo
func (r *NoteRepo) FindByOwnerID(ctx context.Context, ownerID primitive.ObjectID) ([]models.Note, error) {

	// lọc theo owner
	filter := bson.M{"owner_id": ownerID}

	// tìm tất cả note
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	// đảm bảo đóng con trỏ sau khi dùng xong
	defer cursor.Close(ctx)

	// chuyển kết quả thành slice
	var notes []models.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, err
	}

	// trả về kết quả
	return notes, nil
}

// Tìm tất cả note được gửi đến receiver
func (r *NoteRepo) FindByReceiverID(ctx context.Context, receiverID primitive.ObjectID) ([]models.Note, error) {

	// lọc theo receiver_id
	filter := bson.M{"receiver_id": receiverID}

	// tìm tất cả note
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// đảm bảo đóng con trỏ sau khi dùng xong
	defer cursor.Close(ctx)

	// chuyển kết quả thành slice
	var notes []models.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, err
	}

	// trả về kết quả
	return notes, nil
}
