package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Url struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"url_id"`
	NoteID                string             `bson:"note_id" json:"note_id"`       // ID của ghi chú gốc
	ExpiresAt             time.Time          `bson:"expires_at" json:"expires_at"` // Thời gian hết hạn
	MaxAccess             int                `bson:"max_access" json:"max_access"` // Số lượt truy cập tối đa
	Accessed              int                `bson:"accessed" json:"accessed"`     // Số lượt đã truy cập
	SharedEncryptedAESKey string             `bson:"shared_encrypted_aes_key" json:"shared_encrypted_aes_key"`
	Sender                string             `bson:"sender" json:"sender"`
	Receiver              string             `bson:"receiver" json:"receiver"`
}

type CreateUrlRequest struct {
	SharedEncryptedAESKey string `json:"shared_encrypted_aes_key"`
	ExpiresIn             string `json:"expires_in"` // "1h", "30m"
	MaxAccess             int    `json:"max_access"` // int (Server yêu cầu số)
	Sender                string `json:"sender"`
	Receiver              string `json:"receiver"`
}

type UrlResponse struct {
	ID         string    `json:"url_id"` // Khớp với json tag của ObjectID bên server
	NoteID     string    `json:"note_id"`
	SenderID   string    `json:"sender"`
	ReceiverID string    `json:"receiver"`
	ExpiresAt  time.Time `json:"expires_at"`
	MaxAccess  int       `json:"max_access"`
}

func CreateTTLIndex(ctx context.Context, collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// Hàm xử lý truy cập (Gọi mỗi khi user xem link)
func AccessUrl(ctx context.Context, collection *mongo.Collection, urlID primitive.ObjectID) (*Url, error) {
	var url Url
	err1 := collection.FindOne(ctx, bson.M{"_id": urlID}).Decode(&url)
	if err1 != nil {
		return nil, err1 // Lỗi kết nối hoặc không tìm thấy ID
	}

	// So sánh thời gian: Nếu hiện tại > thời gian hết hạn
	if time.Now().UTC().After(url.ExpiresAt) {
		// Xóa luôn document này
		go func() {
			collection.DeleteOne(context.Background(), bson.M{"_id": urlID})
			fmt.Printf("Deleted Expired URL %s\n", urlID.Hex())
		}()
		return nil, fmt.Errorf("link đã hết hạn")
	}

	// Tăng view
	update := bson.M{"$inc": bson.M{"accessed": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err2 := collection.FindOneAndUpdate(ctx, bson.M{"_id": urlID}, update, opts).Decode(&url)
	if err2 != nil {
		return nil, err2
	}

	// Logic xóa nếu vượt quá giới hạn
	if url.Accessed >= url.MaxAccess {
		go func() {
			collection.DeleteOne(context.Background(), bson.M{"_id": urlID})
			fmt.Printf("Deleted URL %s (Limit reached)\n", urlID.Hex())
		}()
	}

	return &url, nil
}
