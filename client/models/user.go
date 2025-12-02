package models

type User struct {
	ID              int    `json:"id"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	PublicKey       string `json:"public_key"`
	EncrypedPrivKey string `json:"encrypted_private_key"`
}

type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	PublicKey       string `json:"public_key"`
	EncrypedPrivKey string `json:"encrypted_private_key"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message             string `json:"message"`
	Token               string `json:"token"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
	Error               string `json:"error,omitempty"`
}
