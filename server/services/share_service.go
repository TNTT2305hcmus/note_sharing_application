package services

import (
	"context"
	"errors"
	"fmt"
	"note_sharing_application/server/configs"
	"note_sharing_application/server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 1. Tạo URL mới
func CreateUrl(noteId, sender, receiver, sharedEncryptedAESKey, expiresIn string, maxAccess int) (string, error) {
	duration, _ := time.ParseDuration(expiresIn)
	expireTime := time.Now().Add(duration)

	newUrl := models.Url{
		NoteID:                noteId,
		SharedEncryptedAESKey: sharedEncryptedAESKey,
		ExpiresAt:             expireTime,
		MaxAccess:             maxAccess,
		Accessed:              0,
		Sender:                sender,
		Receiver:              receiver,
	}

	res, err := configs.GetCollection("urls").InsertOne(context.TODO(), newUrl)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// 2. Lấy URL đang tồn tại của Note
func GetExistingUrl(noteId, receiver string) (string, error) {
	var url models.Url
	// Tìm url của note này mà còn hạn và còn lượt xem
	filter := bson.M{
		"note_id":  noteId,
		"receiver": receiver,
	}

	err := configs.GetCollection("urls").FindOne(context.TODO(), filter).Decode(&url)
	if err != nil {
		return "", errors.New("không tìm thấy URL hợp lệ")
	}

	return url.ID.Hex(), nil
}

// 3. Xử lý xem Note
func GetNote(reqUrl models.Url) (models.Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	urlColl := configs.GetCollection("urls")
	noteColl := configs.GetCollection("notes")

	// B. Gọi hàm AccessUrl (Logic tăng view và tự xóa nằm ở đây)
	// Lưu ý: Truyền reqUrl.ID (kiểu ObjectID) vào thẳng, không cần convert từ string
	validUrl, err := models.AccessUrl(ctx, urlColl, reqUrl.ID)
	if err != nil {
		return models.Note{}, fmt.Errorf("không thể truy cập link này: %v", err)
	}

	// Lấy Note gốc dựa trên NoteID lưu trong Url
	// Vì trong struct Url, NoteID là string, cần convert sang ObjectID để query
	noteOID, err := primitive.ObjectIDFromHex(validUrl.NoteID)
	if err != nil {
		return models.Note{}, fmt.Errorf("Note ID trong dữ liệu bị lỗi")
	}

	var note models.Note
	err = noteColl.FindOne(ctx, bson.M{"_id": noteOID}).Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Note{}, fmt.Errorf("note gốc đã bị xóa khỏi hệ thống")
		}
		return models.Note{}, err
	}
	return note, nil
}
