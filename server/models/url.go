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
	Sender                string             `bson:"sender_id" json:"sender_id"`
	Receiver              string             `bson:"receiver_id" json:"receiver_id"`
}

type CreateUrlRequest struct {
	SharedEncryptedAESKey string `json:"shared_encrypted_aes_key"`
	ExpiresIn             string `json:"expires_in"` // "1h", "30m"
	MaxAccess             int    `json:"max_access"` // int (Server yêu cầu số)
	Sender                string `json:"sender_id"`
	Receiver              string `json:"receiver_id"`
}

func CreateTTLIndex(ctx context.Context, collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// 2. Hàm xử lý truy cập (Gọi mỗi khi user xem link)
func AccessUrl(ctx context.Context, collection *mongo.Collection, urlID primitive.ObjectID) (*Url, error) {
	var updatedUrl Url

	// Tăng view
	filter := bson.M{"_id": urlID}
	update := bson.M{"$inc": bson.M{"accessed": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUrl)
	if err != nil {
		return nil, err
	}

	// Logic xóa nếu vượt quá giới hạn
	if updatedUrl.Accessed >= updatedUrl.MaxAccess {
		go func() {
			collection.DeleteOne(context.Background(), bson.M{"_id": urlID})
			fmt.Printf("Deleted URL %s (Limit reached)\n", urlID.Hex())
		}()
	}

	return &updatedUrl, nil
}
