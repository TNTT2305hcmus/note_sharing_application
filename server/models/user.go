package models

type User struct {
	ID                  int    `json:"id"`
	Username            string `json:"username"`
	Password            string `json:"-"`
	PublicKey           string `json:"public_key"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

type RegisterRequest struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	PublicKey           string `json:"public_key"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
