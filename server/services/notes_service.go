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

func CreateNote(cipherText string, encryptedAesKey string, ownerIDStr string) (string, error) {
	newNote := models.Note{
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
	notes := make([]models.Note, 0)
	if err = cursor.All(context.TODO(), &notes); err != nil {
		return nil, err
	}

	// trả về kết quả
	return notes, nil
}

// Servce: xem tất cả urls được gửi đến receiver
func ViewReceivedNoteURLs(receiverIDStr string) ([]models.Url, error) {

	// lọc url
	urlFilter := bson.M{
		"receiver_id": receiverIDStr,             // Receiver
		"expires_at":  bson.M{"$gt": time.Now()}, // Chưa hết hạn
		"max_access":  bson.M{"$gt": 0},          // Còn lượt truy cập
	}

	// lấy tất cả các urls thỏa mãn lưu vào validUrls
	cursor, err := configs.GetCollection("urls").Find(context.TODO(), urlFilter)
	if err != nil {
		return nil, err
	}
	receivedUrls := make([]models.Url, 0)
	if err = cursor.All(context.TODO(), &receivedUrls); err != nil {
		return nil, err
	}
	return receivedUrls, nil
}

func DeleteNote(noteIDStr string) error {

	// string --> ObjectID lấy ID của node
	NoteIDObj, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		return errors.New("invalid note ID format")
	}

	// lọc note id và người sở hữu
	filter := bson.M{
		"_id": NoteIDObj,
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

func DeleteSharedNote(noteIDStr string, ownerIDStr string) error {

	filter := bson.M{
		"note_id":   noteIDStr,
		"sender_id": ownerIDStr,
	}
	result, err := configs.GetCollection("urls").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("không tìm thấy liên kết chia sẻ nào để xóa")
	}
	return nil
}
