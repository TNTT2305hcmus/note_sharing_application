package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"note_id"`
	Title           string             `bson:"title" json:"title"`
	CipherText      string             `bson:"cipher_text" json:"cipher_text"`             // Nội dung ghi chú đã mã hóa
	EncryptedAesKey string             `bson:"encrypted_aes_key" json:"encrypted_aes_key"` // Key giải mã (đã bị bọc)
	OwnerID         primitive.ObjectID `bson:"owner_id" json:"owner_id"`                   // ID của người tạo (dạng string)
	ReceiverID      primitive.ObjectID `bson:"receiver_id" json:"receiver_id"`             // ID của người nhận
}

type CreateNoteInput struct {
	Title           string
	CipherText      string
	EncryptedAesKey string
	OwnerID         string // về sau cần chuyển về ObjectID để thêm vào database
}