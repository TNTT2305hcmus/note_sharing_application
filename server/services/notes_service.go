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
func ViewOwnedNotes(ownerIDStr string) ([]models.Note, error) {
	// lọc theo owner
	filter := bson.M{"owner_id": ownerIDStr}

	// tìm tất cả note - lúc này chỉ mới trỏ đến collection

	cursor, err := configs.GetCollection("notes").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	// load vào notes thực sự, All() = duyệt, decode, đóng cursor
	var notes []models.Note
	if err = cursor.All(context.TODO(), &notes); err != nil {
		return nil, err
	}

	// trả về kết quả
	return notes, nil
}

// Servce: xem tất cả urls được gửi đến receiver
func ViewReceivedNoteURLs(receiverIDStr string) ([]models.Url, error) {

	// lọc theo receiver_id
	noteFilter := bson.M{"receiver_id": receiverIDStr}

	// lấy tất cả các notes thỏa mãn lưu vào allReceivedNotes
	noteCursor, err := configs.GetCollection("notes").Find(context.TODO(), noteFilter)
	if err != nil {
		return nil, err
	}
	// load vào notes thực sự, All() = duyệt, decode, đóng cursor
	var allReceivedNotes []models.Note
	if err = noteCursor.All(context.TODO(), &allReceivedNotes); err != nil {
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
	urlCursor, err := configs.GetCollection("urls").Find(context.TODO(), urlFilter)
	if err != nil {
		return nil, err
	}
	var validUrls []models.Url
	if err = urlCursor.All(context.TODO(), &validUrls); err != nil {
		return nil, err
	}
	return validUrls, nil
}

func DeleteNote(noteIDStr string, ownerIDStr string) error {

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
	res, err := configs.GetCollection("notes").DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("non-exist note id")
	}

	// lọc note id
	urlFilter := bson.M{"note_id": noteIDStr}

	// Xóa URLs
	_, _ = configs.GetCollection("urls").DeleteMany(context.TODO(), urlFilter)

	return nil

}

func CreateNote(title string, cipherText string, encryptedAesKey string, ownerIDStr string) (string, error) {
	newNote := models.Note{
		Title:           title,
		CipherText:      cipherText,
		EncryptedAesKey: encryptedAesKey,
		OwnerID:         ownerIDStr,
	}

	result, err := configs.GetCollection("notes").InsertOne(context.TODO(), newNote)
	if err != nil {
		return "", err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", err

}

func DeleteSharedNote(noteIDStr string, ownerIDStr string) error {
	noteIDObj, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		return errors.New("invalid note ID format")
	}

	filter := bson.M{
		"_id":         noteIDObj,
		"owner_id":    ownerIDStr,
		"receiver_id": bson.M{"$ne": ""}, // "" thì là sharedNote, còn khác "" là chưa share
	}
	res, err := configs.GetCollection("notes").DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("cannot revoke: note not found, unauthorized, or it is an original note")
	}

	urlFilter := bson.M{"note_id": noteIDStr}
	_, _ = configs.GetCollection("urls").DeleteMany(context.TODO(), urlFilter)

	return nil
}
