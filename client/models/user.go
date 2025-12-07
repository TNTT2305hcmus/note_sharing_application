package models

type User struct {
	ID                  string `json:"id"`
	Username            string `json:"username"`
	Token               string `json:"token"`
	PublicKey           string `json:"public_key"`
	EncryptedPrivateKey string `json:"encrypted_privKey"`
}

type RegisterRequest struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	PublicKey           string `json:"public_key"`
	EncryptedPrivateKey string `json:"encrypted_privKey"`
}

// Request gửi lên khi Đăng Nhập
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message             string `json:"message"`
	Token               string `json:"token"`
	EncryptedPrivateKey string `json:"encrypted_privKey"`
	Error               string `json:"error,omitempty"`
}
