package repositories

import (
	"context"
	"errors"
	"note_sharing_application/server/models"
	"time"

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

// xóa tất cả URL liên quan đến một NoteID
func (r *UrlRepo) DeleteByNoteID(ctx context.Context, noteID string) error {
	filter := bson.M{"note_id": noteID}
	_, err := r.Collection.DeleteMany(ctx, filter)
	return err
}

// tìm các url hợp lệ từ danh sách noteIDs
func (r *UrlRepo) FindValidUrlsByNoteIDs(ctx context.Context, noteIDs []string) ([]models.Url, error) {

	//lọc theo điều kiện
	filter := bson.M{
		"note_id":    bson.M{"$in": noteIDs},    // NoteID nằm trong danh sách truyền vào
		"expires_at": bson.M{"$gt": time.Now()}, // Chưa hết hạn
		"max_access": bson.M{"$gt": 0},          // (Tuỳ chọn) Còn lượt truy cập
	}

	// tìm tất cả url
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// đảm bảo đóng con trỏ sau khi dùng xong
	defer cursor.Close(ctx)

	// đọc tất cả kết quả vào urls
	var urls []models.Url
	if err := cursor.All(ctx, &urls); err != nil {
		return nil, err
	}

	// trả về kết quả
	return urls, nil
}

