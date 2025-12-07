package services

import (
	"context"
	"errors"
	"note_sharing_application/server/configs"
	"note_sharing_application/server/models"
	"note_sharing_application/server/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShareService struct {
	UrlRepo  *repositories.UrlRepo
	NoteRepo *repositories.NoteRepo
}

func NewShareService(u *repositories.UrlRepo, n *repositories.NoteRepo) *ShareService {
	return &ShareService{UrlRepo: u, NoteRepo: n}
}

// 1. Tạo URL mới
func CreateUrl(noteId string, expiresIn string, maxAccess int) (string, error) {
	duration, _ := time.ParseDuration(expiresIn)
	expireTime := time.Now().Add(duration)
	noteID, err := primitive.ObjectIDFromHex(noteId)
	if err != nil {
		return "", err
	}

	newUrl := models.Url{
		NoteID:    noteID,
		ExpiresAt: expireTime,
		MaxAccess: maxAccess,
	}

	res, err := configs.GetCollection("urls").InsertOne(context.TODO(), newUrl)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// 2. Lấy URL đang tồn tại của Note
func GetExistingUrl(noteID string) (string, error) {
	var url models.Url
	// Tìm url của note này mà còn hạn và còn lượt xem
	filter := bson.M{
		"note_id":    noteID,
		"exp":        bson.M{"$gt": time.Now()}, // Còn hạn
		"max_access": bson.M{"$gt": 0},          // Còn lượt
	}

	err := configs.GetCollection("urls").FindOne(context.TODO(), filter).Decode(&url)
	if err != nil {
		return "", errors.New("không tìm thấy URL hợp lệ")
	}

	return url.ID.Hex(), nil
}

// 3. Xử lý xem Note (Check URL, Check quyền, Trả về nội dung)
func GetNote(url models.Url) (*models.Note, error) {
	urlId := url.ID.Hex()

	filter := bson.M{
		"_id":        urlId,
		"max_access": bson.M{"$gt": 0},
	}

	update := bson.M{"$inc": bson.M{"max_access": -1}}

	coll := configs.GetCollection("urls")
	err := coll.FindOneAndUpdate(context.TODO(), filter, update).Decode(&url)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("URL đã hết hạn hoặc hết lượt truy cập")
		}
		return nil, err
	}

	var note models.Note
	noteId := url.NoteID

	err = configs.GetCollection("notes").FindOne(context.TODO(), bson.M{"_id": noteId}).Decode(&note)
	if err != nil {
		return nil, errors.New("note gốc đã bị xóa")
	}

	return &note, nil
}
