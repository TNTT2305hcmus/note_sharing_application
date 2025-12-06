package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username          string             `bson:"username" json:"username"`
	EncryptedPassword string             `bson:"encrypted_password"`
	Salt              string             `bson:"salt" json:"salt"`
	EncryptedPrivKey  string             `bson:"encrypted_privKey" json:"encrypted_privKey"`
	PubKey            string             `bson:"pubKey" json:"pubKey"`
}

type RegisterRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	PublicKey        string `json:"public_key"`
	EncryptedPrivKey string `json:"encrypted_privKey"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
