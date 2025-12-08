package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"note_id"` // omitempty = nếu trường rỗng thì tự sinh ID
	CipherText      string             `bson:"cipher_text" json:"cipher_text"`
	EncryptedAesKey string             `bson:"encrypted_aes_key" json:"encrypted_aes_key"`
	OwnerID         string             `bson:"owner_id" json:"owner_id"`
}

type CreateNoteRequest struct {
	CipherText      string `json:"cipher_text"`
	EncryptedAesKey string `json:"encrypted_aes_key"`
}
