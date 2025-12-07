package services

import (
	"context"
	"errors"
	"note_sharing_application/server/configs"
	"note_sharing_application/server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service: xem tất cả ghi chú do một owner
func ViewOwnedNotes(ctx context.Context, ownerIDStr string) ([]models.Note, error) {
	// lọc theo owner
	filter := bson.M{"owner_id": ownerIDStr}

	// tìm tất cả note - lúc này chỉ mới trỏ đến collection

	cursor, err := configs.GetCollection("notes").Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// load vào notes thực sự, All() = duyệt, decode, đóng cursor
	var notes []models.Note
	if err = cursor.All(ctx, &notes); err != nil {
		return nil, err
	}

	// trả về kết quả
	return notes, nil
}

// Servce: xem tất cả urls được gửi đến receiver
func ViewReceivedNoteURLs(ctx context.Context, receiverIDStr string) ([]models.Url, error) {

	// lọc theo receiver_id
	noteFilter := bson.M{"receiver_id": receiverIDStr}

	// lấy tất cả các notes thỏa mãn lưu vào allReceivedNotes
	noteCursor, err := configs.GetCollection("notes").Find(ctx, noteFilter)
	if err != nil {
		return nil, err
	}
	// load vào notes thực sự, All() = duyệt, decode, đóng cursor
	var allReceivedNotes []models.Note
	if err = noteCursor.All(ctx, &allReceivedNotes); err != nil {
		return nil, err
	}

	// Nếu không có note nào
	if len(allReceivedNotes) == 0 {
		return []models.Url{}, nil
	}

	// Lấy ID để truy vấn URLs
	noteIDStrs := make([]string, 0, len(allReceivedNotes))
	for _, note := range allReceivedNotes {
		noteIDStrs = append(noteIDStrs, note.ID.Hex())
	}

	// lọc url
	urlFilter := bson.M{
		"note_id":    bson.M{"$in": noteIDStrs}, // NoteID nằm trong danh sách truyền vào
		"expires_at": bson.M{"$gt": time.Now()}, // Chưa hết hạn
		"max_access": bson.M{"$gt": 0},          // Còn lượt truy cập
	}

	// lấy tất cả các urls thỏa mãn lưu vào validUrls
	urlCursor, err := configs.GetCollection("urls").Find(ctx, urlFilter)
	if err != nil {
		return nil, err
	}
	var validUrls []models.Url
	if err = urlCursor.All(ctx, &validUrls); err != nil {
		return nil, err
	}
	return validUrls, nil
}

func DeleteNote(ctx context.Context, noteIDStr string, ownerIDStr string) error {

	// string --> ObjectID lấy ID của node
	NoteIDObj, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		return errors.New("invalid note ID format")
	}

	// lọc note id và người sở hữu
	filter := bson.M{
		"_id":      NoteIDObj,
		"owner_id": ownerIDStr,
	}

	// Xóa note
	res, err := configs.GetCollection("notes").DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("non-exist note id")
	}

	// lọc note id
	urlFilter := bson.M{"note_id": noteIDStr}

	// Xóa URLs
	_, _ = configs.GetCollection("urls").DeleteMany(ctx, urlFilter)

	return nil

}
