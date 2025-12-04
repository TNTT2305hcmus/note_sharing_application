package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"note_id"`
	Title           string             `bson:"title" json:"title"`
	CipherText      string             `bson:"cipher_text" json:"cipher_text"`             // Nội dung ghi chú đã mã hóa
	EncryptedAesKey string             `bson:"encrypted_aes_key" json:"encrypted_aes_key"` // Key giải mã (đã bị bọc)
	Owner           string             `bson:"owner" json:"owner"`                         // ID của người tạo (dạng string)
	Receiver        string             `bson:"receiver" json:"receiver"`                   // ID của người nhận (nếu gửi đích danh)
}
