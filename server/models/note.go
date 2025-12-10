package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"note_id"` // omitempty = nếu trường rỗng thì tự sinh ID
	Title           string             `bson:"title" json:"title"`
	CipherText      string             `bson:"cipher_text" json:"cipher_text"`             // Nội dung ghi chú đã mã hóa
	EncryptedAesKey string             `bson:"encrypted_aes_key" json:"encrypted_aes_key"` // Key giải mã (đã bị bọc)
	OwnerID         string             `bson:"owner_id" json:"owner_id"`                   // ID của người tạo (dạng string)
}

type CreateNoteRequest struct {
	CipherText      string `json:"cipher_text"`
	EncryptedAesKey string `json:"encrypted_aes_key_by_K"`
	Sender          string `json:"sender"`
}
